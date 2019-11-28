package nic

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
)

type (
	// Session is the wrapper for http.Client and http.Request
	Session struct {
		Client                 *http.Client
		request                *http.Request
		beforeRequestHookFuncs []BeforeRequestHookFunc
		afterResponseHookFuncs []AfterResponseHookFunc
		sync.Mutex
	}
)

var (
	// disable automatic redirect
	disableRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// from HTTP std lib
	// automatic redirection allowed 10 times
	defaultCheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
)

// NewSession returns an empty Session
func NewSession() *Session {
	client := &http.Client{}
	jar, _ := cookiejar.New(nil)
	client.Jar = jar
	client.Transport = &http.Transport{}

	return &Session{
		Client: client,
	}
}

// Request is the base method
func (s *Session) Request(method string, urlStr string, option Option) (*Response, error) {
	s.Lock()
	defer s.Unlock()

	method = strings.ToUpper(method)
	switch method {
	case HEAD, GET, POST, DELETE, OPTIONS, PUT, PATCH:
		// url encode the query string
		urlStrParsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		urlStrParsed.RawQuery = urlStrParsed.Query().Encode()

		s.request, err = http.NewRequest(method, urlStrParsed.String(), nil)
		if err != nil {
			return nil, err
		}
		s.request.Header.Set("User-Agent", userAgent)
		s.request.Close = true

		if s.Client == nil {
			s.Client = &http.Client{}
			jar, _ := cookiejar.New(nil)
			s.Client.Jar = jar
			s.Client.Transport = &http.Transport{}
		}

		if option != nil {
			// set options of http.Request
			err = option.setRequestOpt(s.request)
			if err != nil {
				return nil, err
			}

			// all client config will be restored to the default value after every request
			defer func() {
				s.Client.CheckRedirect = defaultCheckRedirect
				s.Client.Timeout = 0
				s.Client.Transport = &http.Transport{}
			}()

			// set options of http.Client
			err = option.setClientOpt(s.Client)
			if err != nil {
				return nil, err
			}
		}

		for _, fn := range s.beforeRequestHookFuncs {
			err = fn(s.request)
			if err != nil {
				break
			}
		}

	default:
		return nil, ErrInvalidMethod
	}

	// do request then parse response
	r, err := s.Client.Do(s.request)
	if err != nil {
		return nil, err
	}

	for _, fn := range s.afterResponseHookFuncs {
		err = fn(r)
		if err != nil {
			break
		}
	}

	resp, err := NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetRequest returns nic.Session.request
func (s *Session) GetRequest() *http.Request {
	return s.request
}

type (
	BeforeRequestHookFunc func(*http.Request) error
	AfterResponseHookFunc func(*http.Response) error
)

// Register the before request hook
func (s *Session) RegisterBeforeReqHook(fn BeforeRequestHookFunc) error {
	if s.beforeRequestHookFuncs == nil {
		s.beforeRequestHookFuncs = make([]BeforeRequestHookFunc, 0, 8)
	}
	if len(s.beforeRequestHookFuncs) > 7 {
		return ErrHookFuncMaxLimit
	}
	s.beforeRequestHookFuncs = append(s.beforeRequestHookFuncs, fn)
	return nil
}

// Unregister the request hook, pass the function's index(start at 0)
func (s *Session) UnregisterBeforeReqHook(index int) error {
	if index >= len(s.beforeRequestHookFuncs) {
		return ErrIndexOutofBound
	}
	s.beforeRequestHookFuncs = append(s.beforeRequestHookFuncs[:index], s.beforeRequestHookFuncs[index+1:]...)
	return nil
}

// Reset all before request hook
func (s *Session) ResetBeforeReqHook() {
	s.beforeRequestHookFuncs = []BeforeRequestHookFunc{}
}

// Register the after response hook
func (s *Session) RegisterAfterRespHook(fn AfterResponseHookFunc) error {
	if s.afterResponseHookFuncs == nil {
		s.afterResponseHookFuncs = make([]AfterResponseHookFunc, 0, 8)
	}
	if len(s.afterResponseHookFuncs) > 7 {
		return ErrHookFuncMaxLimit
	}
	s.afterResponseHookFuncs = append(s.afterResponseHookFuncs, fn)
	return nil
}

// Unregister the response hook, pass the function's index(start at 0)
func (s *Session) UnregisterAfterRespHook(index int) error {
	if index >= len(s.afterResponseHookFuncs) {
		return ErrIndexOutofBound
	}
	s.afterResponseHookFuncs = append(s.afterResponseHookFuncs[:index], s.afterResponseHookFuncs[index+1:]...)
	return nil
}

// Reset all after response hook
func (s *Session) ResetAfterRespHook() {
	s.afterResponseHookFuncs = []AfterResponseHookFunc{}
}

// Get is a shortcut for get method
func (s *Session) Get(url string, option Option) (*Response, error) {
	return s.Request("get", url, option)
}

// Post is a shortcut for get method
func (s *Session) Post(url string, option Option) (*Response, error) {
	return s.Request("post", url, option)
}

// Head is a shortcut for get method
func (s *Session) Head(url string, option Option) (*Response, error) {
	return s.Request("head", url, option)
}

// Delete is a shortcut for get method
func (s *Session) Delete(url string, option Option) (*Response, error) {
	return s.Request("delete", url, option)
}

// Options is a shortcut for get method
func (s *Session) Options(url string, option Option) (*Response, error) {
	return s.Request("options", url, option)
}

// Put is a shortcut for get method
func (s *Session) Put(url string, option Option) (*Response, error) {
	return s.Request("put", url, option)
}

// Patch is a shortcut for get method
func (s *Session) Patch(url string, option Option) (*Response, error) {
	return s.Request("patch", url, option)
}
