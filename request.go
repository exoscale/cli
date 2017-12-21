package egoscale

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Request represents a CloudStack request
type Request interface {
	Command() string
}

const (
	// PENDING represents a job in progress
	PENDING JobStatusType = iota
	// SUCCESS represents a successfully completed job
	SUCCESS
	// FAILURE represents a job that has failed to complete
	FAILURE
)

// JobStatusType represents the status of a Job
type JobStatusType int

// JobResultResponse represents a generic response to a job task
type JobResultResponse struct {
	AccountId     string           `json:"accountid,omitempty"`
	Cmd           string           `json:"cmd,omitempty"`
	CreatedAt     string           `json:"created,omitempty"`
	JobId         string           `json:"jobid,omitempty"`
	JobProcStatus int              `json:"jobprocstatus,omitempty"`
	JobResult     *json.RawMessage `json:"jobresult,omitempty"`
	JobStatus     JobStatusType    `json:"jobstatus,omitempty"`
	JobResultType string           `json:"jobresulttype,omitempty"`
	UserId        string           `json:"userid,omitempty"`
}

// ErrorResponse represents the standard error response from CloudStack
type ErrorResponse struct {
	ErrorCode   int      `json:"errorcode"`
	CsErrorCode int      `json:"cserrorcode"`
	ErrorText   string   `json:"errortext"`
	UuidList    []string `json:"uuidList,omitempty"` // uuid*L*ist is not a typo
}

// Error formats a CloudStack error into a standard error
func (e *ErrorResponse) Error() error {
	return fmt.Errorf("API error %d (internal code: %d): %s", e.ErrorCode, e.CsErrorCode, e.ErrorText)
}

// BooleanResponse represents a boolean response (usually after a deletion)
type BooleanResponse struct {
	Success     bool   `json:"success"`
	DisplayText string `json:"diplaytext,omitempty"`
}

// Error formats a CloudStack job response into a standard error
func (e *BooleanResponse) Error() error {
	if e.Success {
		return nil
	}
	return fmt.Errorf("API error: %s", e.DisplayText)
}

// AsyncInfo represents the details for any async call
//
// It retries at most Retries time and waits for Delay between each retry
type AsyncInfo struct {
	Retries int
	Delay   int
}

func csQuotePlus(s string) string {
	return strings.Replace(s, "+", "%20", -1)
}

func csEncode(s string) string {
	return csQuotePlus(url.QueryEscape(s))
}

func rawValue(b json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	return nil, nil
}

func rawValues(b json.RawMessage) (json.RawMessage, error) {
	var i []json.RawMessage

	if err := json.Unmarshal(b, &i); err != nil {
		return nil, nil
	}

	return i[0], nil
}

func (exo *Client) ParseResponse(resp *http.Response) (json.RawMessage, error) {
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

		/* Need to account for different error types */
		if e.ErrorCode != 0 {
			return nil, e.Error()
		} else {
			var de DNSError
			if err := json.Unmarshal(b, &de); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("Exoscale error (%d): %s", resp.StatusCode, strings.Join(de.Name, "\n"))
		}
	}

	return b, nil
}

// AsyncRequest performs an asynchronous request and polls it for retries * day [s]
func (exo *Client) AsyncRequest(command Request, v interface{}, async AsyncInfo) error {
	body, err := exo.request(command)
	if err != nil {
		return err
	}

	// Is it a Job?
	var job JobResultResponse
	if err := json.Unmarshal(body, &job); err != nil {
		return err
	}

	if job.JobId == "" || job.JobStatus != PENDING {
		return json.Unmarshal(*job.JobResult, v)
	}

	// we've go a pending job
	result := new(QueryAsyncJobResultResponse)
	result.JobStatus = job.JobStatus
	for async.Retries > 0 && result.JobStatus == PENDING {
		time.Sleep(time.Duration(async.Delay) * time.Second)

		async.Retries--

		req := &QueryAsyncJobResultRequest{JobId: job.JobId}
		err = exo.Request(req, result)
		if err != nil {
			return err
		}
	}

	if result.JobStatus == FAILURE {
		var r ErrorResponse
		if err := json.Unmarshal(*result.JobResult, &r); err != nil {
			return err
		}
		return r.Error()
	}

	if result.JobStatus == PENDING {
		return fmt.Errorf("Maximum number of retries reached")
	}

	return json.Unmarshal(*job.JobResult, v)
}

// BooleanRequest performs a sync request on a boolean call
func (exo *Client) BooleanRequest(command Request) error {
	var r BooleanResponse
	err := exo.Request(command, &r)
	if err != nil {
		return err
	}

	return r.Error()
}

// BooleanAsyncRequest performs a sync request on a boolean call
func (exo *Client) BooleanAsyncRequest(command Request, async AsyncInfo) error {
	var r BooleanResponse
	err := exo.AsyncRequest(command, &r, async)
	if err != nil {
		return err
	}

	return r.Error()
}

// Request performs a simple request
func (exo *Client) Request(command Request, v interface{}) error {
	resp, err := exo.request(command)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, v); err != nil {
		var r ErrorResponse
		if e := json.Unmarshal(resp, &r); e == nil {
			return r.Error()
		}
		return err
	}

	return nil
}

// request makes a Request while being close to the metal
func (exo *Client) request(command Request) (json.RawMessage, error) {
	mac := hmac.New(sha1.New, []byte(exo.apiSecret))

	params := url.Values{}
	err := prepareValues("", &params, command)
	if err != nil {
		return nil, err
	}

	params.Set("apikey", exo.apiKey)
	params.Set("command", command.Command())
	params.Set("response", "json")

	keys := make([]string, 0)
	unencoded := make([]string, 0)
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		arg := fmt.Sprintf("%s=%s", k, csEncode(params[k][0]))
		unencoded = append(unencoded, arg)
	}

	signString := strings.ToLower(strings.Join(unencoded, "&"))

	mac.Write([]byte(signString))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	query := params.Encode()
	reader := strings.NewReader(fmt.Sprintf("%s&signature=%s", csQuotePlus(query), signature))

	// Use PostForm?
	resp, err := exo.client.Post(exo.endpoint, "application/x-www-form-urlencoded", reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := exo.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (exo *Client) DetailedRequest(uri string, params string, method string, header http.Header) (json.RawMessage, error) {
	url := exo.endpoint + uri

	req, err := http.NewRequest(method, url, strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	req.Header = header

	response, err := exo.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return exo.ParseResponse(response)
}

// prepareValues uses a command to build a POST request
//
// command is not a Command so it's easier to Test
func prepareValues(prefix string, params *url.Values, command interface{}) error {
	value := reflect.ValueOf(command)
	typeof := reflect.TypeOf(command)
	// Going up the pointer chain to find the underlying struct
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		value = value.Elem()
	}

	for i := 0; i < typeof.NumField(); i++ {
		field := typeof.Field(i)
		val := value.Field(i)
		tag := field.Tag
		if json, ok := tag.Lookup("json"); ok {
			n, required := extractJsonTag(field.Name, json)
			name := prefix + n

			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v := val.Int()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0.", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatInt(v, 10))
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v := val.Uint()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0.", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatUint(v, 10))
				}
			case reflect.Float32, reflect.Float64:
				v := val.Float()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0.", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatFloat(v, 'f', -1, 64))
				}
			case reflect.String:
				v := val.String()
				if v == "" {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got \"\".", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, v)
				}
			case reflect.Bool:
				v := val.Bool()
				if v == false {
					if required {
						params.Set(name, "false")
					}
				} else {
					(*params).Set(name, "true")
				}
			case reflect.Slice:
				switch field.Type.Elem().Kind() {
				case reflect.Uint8:
					if val.Len() == 0 {
						if required {
							return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), field.Name, val.Kind())
						}
					} else {
						v := val.Bytes()
						(*params).Set(name, base64.StdEncoding.EncodeToString(v))
					}
				case reflect.Ptr:
					if val.Len() == 0 {
						if required {
							return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), field.Name, val.Kind())
						}
					} else {
						err := prepareList(name, params, val.Interface())
						if err != nil {
							return err
						}
					}
				default:
					if required {
						return fmt.Errorf("Unsupported type %s.%s ([]%s)", typeof.Name(), field.Name, field.Type.Elem().Kind())
					}
				}
			default:
				if required {
					return fmt.Errorf("Unsupported type %s.%s (%v)", typeof.Name(), field.Name, val.Kind())
				}
			}
		} else {
			log.Printf("[SKIP] %s.%s no json label found", typeof.Name(), field.Name)
		}
	}

	return nil
}

func prepareList(prefix string, params *url.Values, slice interface{}) error {
	value := reflect.ValueOf(slice)

	for i := 0; i < value.Len(); i++ {
		prepareValues(fmt.Sprintf("%s[%d].", prefix, i), params, value.Index(i).Interface())
	}

	return nil
}

// extractJsonTag returns the variable name or defaultName as well as if the field is required (!omitempty)
func extractJsonTag(defaultName, jsonTag string) (string, bool) {
	tags := strings.Split(jsonTag, ",")
	name := tags[0]
	required := true
	for _, tag := range tags {
		if tag == "omitempty" {
			required = false
		}
	}

	if name == "" || name == "omitempty" {
		name = defaultName
	}
	return name, required
}
