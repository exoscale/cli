package egoscale

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

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

		/* Need to account for differet error types */
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
func (exo *Client) AsyncRequest(command string, params url.Values, async AsyncInfo) (json.RawMessage, error) {
	body, err := exo.request(command, params)
	if err != nil {
		return nil, err
	}

	// This is not a Job
	var job JobResultResponse
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, err
	}

	if job.JobId != "" {
		if job.JobStatus == SUCCESS {
			return *job.JobResult, nil
		} else if job.JobStatus == FAILURE {
			return nil, fmt.Errorf("Job %s failed. %s", job.JobId, job.JobResultType)
		}

		// we've go a pending job
		for async.Retries > 0 {
			time.Sleep(time.Duration(async.Delay) * time.Second)

			async.Retries--

			resp, err := exo.PollAsyncJob(job.JobId)
			if err != nil {
				return nil, err
			}

			if resp.JobStatus == SUCCESS {
				return *resp.JobResult, nil
			} else if resp.JobStatus == FAILURE {
				return nil, fmt.Errorf("Job %s failed. %s", job.JobId, resp.JobResultType)
			}
		}

		return nil, fmt.Errorf("Maximum number of retries reached")
	} else {
		// the job is done
		return body, nil
	}
}

// Request performs a sync request (one try only)
func (exo *Client) Request(command string, params url.Values) (json.RawMessage, error) {
	return exo.AsyncRequest(command, params, AsyncInfo{})
}

// request makes a Request while being close to the metal
func (exo *Client) request(command string, params url.Values) (json.RawMessage, error) {
	mac := hmac.New(sha1.New, []byte(exo.apiSecret))

	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
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

	sign_string := strings.ToLower(strings.Join(unencoded, "&"))

	mac.Write([]byte(sign_string))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	query := params.Encode()
	url := fmt.Sprintf("%s?%s&signature=%s", exo.endpoint, csQuotePlus(query), signature)

	resp, err := exo.client.Get(url)
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
