package retry

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryWhenImmediatelySuccessful(t *testing.T) {
	assert := assert.New(t)

	r := New(2, TestBackoff)
	err := r.Retry(func(attempt int) (error, bool) {
		// no error
		return nil, false
	})

	assert.Nil(err)
}

func TestRetryWhenEventuallySuccessful(t *testing.T) {
	assert := assert.New(t)

	do := func(attempt int) (error, bool) {
		switch attempt {
		case 3:
			// Third attempt is ok
			return nil, false
		default:
			// Other attempts (including #1 and #2) fail but a retry is possible
			return fmt.Errorf("Error on attempt #%d", attempt), true
		}
	}

	r1 := New(3, TestBackoff)
	err1 := r1.Retry(do)
	assert.Nil(err1)

	// fails if not enough attempts
	r2 := New(2, TestBackoff)
	err2 := r2.Retry(do)
	assert.Error(err2)
	assert.Equal("Error on attempt #2", err2.Error())
}

func TestRetryWhenNoPossibleRetry(t *testing.T) {
	assert := assert.New(t)

	r := New(3, TestBackoff)

	do := func(attempt int) (error, bool) {
		return fmt.Errorf("Error on attempt #%d", attempt), false
	}

	err := r.Retry(do)
	assert.Error(err)
	// gave up at first call
	assert.Equal("Error on attempt #1", err.Error())
}

func TestSendHTTPRequestWhenItEventuallySucceeds(t *testing.T) {
	assert := assert.New(t)

	// this client will fail twice with a 500 error then will succeed
	body := ioutil.NopCloser(strings.NewReader("blah"))
	httpClient := &mockHTTPDoer{}
	httpClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 500, Body: body}, nil).Twice()
	httpClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 200, Body: body}, nil).Once()

	r := New(3, TestBackoff)

	res, err := SendHTTPRequest(r, httpClient, &http.Request{})

	assert.Nil(err)
	assert.Equal(200, res.StatusCode)
}

func TestSendHTTPRequestWhenItEventuallyFails(t *testing.T) {
	assert := assert.New(t)

	body := ioutil.NopCloser(strings.NewReader("blah"))
	httpClient := &mockHTTPDoer{}
	httpClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 500, Body: body}, nil).Times(3)

	r := New(3, TestBackoff)

	res, err := SendHTTPRequest(r, httpClient, &http.Request{})

	assert.NotNil(err)
	assert.Equal(500, res.StatusCode)
}

func TestSendHTTPRequestDoesNotRetryInCaseOf400Status(t *testing.T) {
	assert := assert.New(t)

	body := ioutil.NopCloser(strings.NewReader("blah"))
	httpClient := &mockHTTPDoer{}
	httpClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 400, Body: body}, nil).Once()

	r := New(3, TestBackoff)

	res, err := SendHTTPRequest(r, httpClient, &http.Request{})

	assert.NotNil(err)
	assert.Equal(400, res.StatusCode)
}

func TestSendHTTPRequestDoesNotRetryInCaseOfError(t *testing.T) {
	assert := assert.New(t)

	httpClient := &mockHTTPDoer{}
	httpClient.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Times(3)

	r := New(3, TestBackoff)

	res, err := SendHTTPRequest(r, httpClient, &http.Request{})

	assert.NotNil(err)
	assert.Nil(res)
}

func TestBodyReaderIsStillReadeableAfterAFailedAttempt(t *testing.T) {
	assert := assert.New(t)

	body := ioutil.NopCloser(strings.NewReader("blah"))
	req := &http.Request{
		Body: body,
	}

	httpClient := &mockHTTPDoer{}
	httpClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 200, Body: body}, nil).Once()

	r := New(3, TestBackoff)
	res, err := SendHTTPRequest(r, httpClient, req)

	assert.NoError(err)
	assert.Equal(200, res.StatusCode)

	finalBody, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)
	assert.Equal([]byte("blah"), finalBody)
}

type mockHTTPDoer struct {
	mock.Mock
}

func (m *mockHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	// the req body is consumed to simulate the real method's side effect
	if req.Body != nil {
		ioutil.ReadAll(req.Body)
	}

	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}
