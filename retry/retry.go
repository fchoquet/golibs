package retry

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// BackOffFunc return the time before the next attempt
// it is not called before the 1st attempt (index zero)
// but only before the second one. So it is initially called with i = 1
type BackOffFunc func(i int) time.Duration

// AttemptFunc is the closure called at every attempt.
// In case of error it states if a retry is possible or if we should give up
// For instance when calling an Http Api, a 504 should be retried, not a 400
type AttemptFunc func(attempt int) (err error, retry bool)

// Retrier retries an AttemptFunc and return the final results
// This is the implementer's responsibility to limit the number of attempts
// And use an appropriate back off strategy
type Retrier interface {
	Retry(do AttemptFunc) error
}

// New returns a default Retrier implementation
func New(maxAttempts int, backoff BackOffFunc) Retrier {
	return &retrier{
		maxAttempts: maxAttempts,
		backoff:     backoff,
	}
}

type retrier struct {
	maxAttempts int
	backoff     BackOffFunc
}

func (r *retrier) Retry(do AttemptFunc) error {
	var lastErr error

	for i := 0; i < r.maxAttempts; i++ {
		if i > 0 {
			time.Sleep(r.backoff(i))
		}

		var retry bool
		lastErr, retry = do(i + 1)

		if lastErr == nil || !retry {
			// no error or no retry
			break
		}
	}
	return lastErr
}

// SendHTTPRequest sends an http request using the provided Retrier
func SendHTTPRequest(r Retrier, client httpDoer, req *http.Request) (res *http.Response, err error) {

	err = r.Retry(func(attempt int) (err error, retry bool) {
		// req.Body is consumed when we call Do. Let's use a clone
		newReq, err1 := cloneRequest(req)
		if err1 != nil {
			return err1, false
		}

		res, err = client.Do(newReq)
		if err != nil {
			return err, true
		}

		if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 202 {
			body, _ := ioutil.ReadAll(res.Body)
			// Will not retry 4xx statuses but will retry 5xx
			if res.StatusCode >= 500 {
				retry = true
			}
			err = fmt.Errorf("http error %d : %s", res.StatusCode, body)
		}
		return
	})

	return
}

// clone an http request and preserves body
func cloneRequest(req *http.Request) (*http.Request, error) {
	newReq := *req
	if req.Body != nil {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		newReq.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	return &newReq, nil
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
