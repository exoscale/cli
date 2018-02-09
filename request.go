package egoscale

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Command represents a CloudStack request
type Command interface {
	// CloudStack API command name
	APIName() string
}

// SyncCommand represents a CloudStack synchronous request
type syncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	response() interface{}
}

// asyncCommand represents a async CloudStack request
type asyncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	asyncResponse() interface{}
}

// onBeforeHook represents an action to be done on the params before sending them
//
// This little took helps with issue of relying on JSON serialization logic only.
// `omitempty` may make sense in some cases but not all the time.
type onBeforeHook interface {
	onBeforeSend(params *url.Values) error
}

const (
	// Pending represents a job in progress
	Pending JobStatusType = iota
	// Success represents a successfully completed job
	Success
	// Failure represents a job that has failed to complete
	Failure
)

// JobStatusType represents the status of a Job
type JobStatusType int

const (
	// Unauthorized represents ... (TODO)
	Unauthorized ErrorCode = 401
	// MethodNotAllowed represents ... (TODO)
	MethodNotAllowed = 405
	// UnsupportedActionError represents ... (TODO)
	UnsupportedActionError = 422
	// APILimitExceeded represents ... (TODO)
	APILimitExceeded = 429
	// MalformedParameterError represents ... (TODO)
	MalformedParameterError = 430
	// ParamError represents ... (TODO)
	ParamError = 431

	// InternalError represents a server error
	InternalError = 530
	// AccountError represents ... (TODO)
	AccountError = 531
	// AccountResourceLimitError represents ... (TODO)
	AccountResourceLimitError = 532
	// InsufficientCapacityError represents ... (TODO)
	InsufficientCapacityError = 533
	// ResourceUnavailableError represents ... (TODO)
	ResourceUnavailableError = 534
	// ResourceAllocationError represents ... (TODO)
	ResourceAllocationError = 535
	// ResourceInUseError represents ... (TODO)
	ResourceInUseError = 536
	// NetworkRuleConflictError represents ... (TODO)
	NetworkRuleConflictError = 537
)

// ErrorCode represents the CloudStack ApiErrorCode enum
//
// See: https://github.com/apache/cloudstack/blob/master/api/src/org/apache/cloudstack/api/ApiErrorCode.java
type ErrorCode int

// JobResultResponse represents a generic response to a job task
type JobResultResponse struct {
	AccountID     string           `json:"accountid,omitempty"`
	Cmd           string           `json:"cmd"`
	Created       string           `json:"created"`
	JobID         string           `json:"jobid"`
	JobProcStatus int              `json:"jobprocstatus"`
	JobResult     *json.RawMessage `json:"jobresult"`
	JobStatus     JobStatusType    `json:"jobstatus"`
	JobResultType string           `json:"jobresulttype"`
	UserID        string           `json:"userid,omitempty"`
}

// ErrorResponse represents the standard error response from CloudStack
type ErrorResponse struct {
	ErrorCode   ErrorCode  `json:"errorcode"`
	CsErrorCode int        `json:"cserrorcode"`
	ErrorText   string     `json:"errortext"`
	UUIDList    []UUIDItem `json:"uuidList,omitempty"` // uuid*L*ist is not a typo
}

// UUIDItem represents an item of the UUIDList part of an ErrorResponse
type UUIDItem struct {
	Description      string `json:"description,omitempty"`
	SerialVersionUID int64  `json:"serialVersionUID,omitempty"`
	UUID             string `json:"uuid"`
}

// Error formats a CloudStack error into a standard error
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error %d (internal code: %d): %s", e.ErrorCode, e.CsErrorCode, e.ErrorText)
}

// booleanAsyncResponse represents a boolean response (usually after a deletion)
type booleanAsyncResponse struct {
	Success     bool   `json:"success"`
	DisplayText string `json:"diplaytext,omitempty"`
}

// Error formats a CloudStack job response into a standard error
func (e *booleanAsyncResponse) Error() error {
	if e.Success {
		return nil
	}
	return fmt.Errorf("API error: %s", e.DisplayText)
}

// booleanAsyncResponse represents a boolean response for sync calls
type booleanSyncResponse struct {
	Success     string `json:"success"`
	DisplayText string `json:"displaytext,omitempty"`
}

func (e *booleanSyncResponse) Error() error {
	if e.Success == "true" {
		return nil
	}

	return fmt.Errorf("API error: %s", e.DisplayText)
}

type syncJob struct {
	command      syncCommand
	responseChan chan<- interface{}
	errorChan    chan<- error
	ctx          context.Context
}

type asyncJob struct {
	command      asyncCommand
	responseChan chan<- *AsyncJobResult
	errorChan    chan<- error
	ctx          context.Context
}

func (exo *Client) parseResponse(resp *http.Response) (json.RawMessage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	a, err := rawValues(b)

	if a == nil {
		b, err = rawValue(b)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 {
		var e ErrorResponse
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		return b, &e
	}
	return b, nil
}

func (exo *Client) processSyncJob(ctx context.Context, job *syncJob) {
	defer close(job.responseChan)
	defer close(job.errorChan)

	body, err := exo.request(ctx, job.command.APIName(), job.command)
	if err != nil {
		job.errorChan <- err
		return
	}

	resp := job.command.response()
	if err := json.Unmarshal(body, resp); err != nil {
		r := new(ErrorResponse)
		if e := json.Unmarshal(body, r); e != nil {
			job.errorChan <- r
			return
		}
		job.errorChan <- err
		return
	}

	job.responseChan <- resp.(interface{})
}

func (exo *Client) processAsyncJob(ctx context.Context, job *asyncJob) {
	defer close(job.responseChan)
	defer close(job.errorChan)

	body, err := exo.request(ctx, job.command.APIName(), job.command)
	if err != nil {
		job.errorChan <- err
		return
	}

	jobResult := new(AsyncJobResult)
	if err := json.Unmarshal(body, jobResult); err != nil {
		r := new(ErrorResponse)
		if e := json.Unmarshal(body, r); e != nil {
			job.errorChan <- r
			return
		}
		job.errorChan <- err
		return
	}

	// Successful response
	if jobResult.JobID == "" || jobResult.JobStatus != Pending {
		job.responseChan <- jobResult
		return
	}

	for iteration := 0; ; iteration++ {
		select {
		case <-ctx.Done():
			job.errorChan <- ctx.Err()
			return
		default:
			time.Sleep(exo.RetryStrategy(int64(iteration)))

			req := &QueryAsyncJobResult{JobID: jobResult.JobID}
			resp, err := exo.Request(req)
			if err != nil {
				job.errorChan <- err
				return
			}

			result := resp.(*QueryAsyncJobResultResponse)
			if result.JobStatus == Success {
				job.responseChan <- (*AsyncJobResult)(result)
				return
			} else if result.JobStatus == Failure {
				r := new(ErrorResponse)
				e := json.Unmarshal(*result.JobResult, r)
				if e != nil {
					job.errorChan <- e
					return
				}
				job.errorChan <- r
				return
			}
		}
	}
}

// asyncRequest perform an asynchronous job with a context
func (exo *Client) asyncRequest(ctx context.Context, req asyncCommand) (interface{}, error) {
	responseChan := make(chan *AsyncJobResult, 1)
	errorChan := make(chan error, 1)

	go exo.processAsyncJob(ctx, &asyncJob{
		command:      req,
		responseChan: responseChan,
		errorChan:    errorChan,
		ctx:          ctx,
	})

	select {
	case result := <-responseChan:
		resp := req.asyncResponse()
		if err := json.Unmarshal(*(result.JobResult), resp); err != nil {
			return nil, err
		}
		return resp, nil

	case err := <-errorChan:
		return nil, err

	case <-ctx.Done():
		err := <-errorChan
		return nil, err
	}
}

// syncRequest performs a sync request with a context
func (exo *Client) syncRequest(ctx context.Context, req syncCommand) (interface{}, error) {
	responseChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	go exo.processSyncJob(ctx, &syncJob{
		command:      req,
		responseChan: responseChan,
		errorChan:    errorChan,
		ctx:          ctx,
	})

	select {
	case result := <-responseChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		err := <-errorChan
		return nil, err
	}
}

// BooleanRequest performs the given boolean command
func (exo *Client) BooleanRequest(req Command) error {
	resp, err := exo.Request(req)
	if err != nil {
		return err
	}

	// CloudStack returns a different type between sync and async success responses
	if b, ok := resp.(*booleanSyncResponse); ok {
		return b.Error()
	}
	if b, ok := resp.(*booleanAsyncResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("The command %s is not a proper boolean response. %#v", req.APIName(), resp))
}

// Request performs the given command
func (exo *Client) Request(req Command) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exo.Timeout)
	defer cancel()

	switch req.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, req.(syncCommand))
	case asyncCommand:
		return exo.asyncRequest(ctx, req.(asyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", req.APIName()))
	}
}

// RequestWithContext preforms a request with a context
func (exo *Client) RequestWithContext(ctx context.Context, req Command) (interface{}, error) {
	switch req.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, req.(syncCommand))
	case asyncCommand:
		return exo.asyncRequest(ctx, req.(asyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", req.APIName()))
	}
}

// request makes a Request while being close to the metal
func (exo *Client) request(ctx context.Context, command string, req interface{}) (json.RawMessage, error) {
	params := url.Values{}
	err := prepareValues("", &params, req)
	if err != nil {
		return nil, err
	}
	if hookReq, ok := req.(onBeforeHook); ok {
		hookReq.onBeforeSend(&params)
	}
	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
	params.Set("response", "json")

	// This code is borrowed from net/url/url.go
	// The way it's encoded by net/url doesn't match
	// how CloudStack works.
	var buf bytes.Buffer
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		prefix := csEncode(k) + "="
		for _, v := range params[k] {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(csEncode(v))
		}
	}

	query := buf.String()

	mac := hmac.New(sha1.New, []byte(exo.apiSecret))
	mac.Write([]byte(strings.ToLower(query)))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	payload := fmt.Sprintf("%s&signature=%s", csQuotePlus(query), signature)

	request, err := http.NewRequest("POST", exo.endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(payload)))

	resp, err := exo.client.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := exo.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return body, nil
}
