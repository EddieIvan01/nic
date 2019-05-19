package nic

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	// Session is the wrapper for http.Client and http.Request
	Session struct {
		Client  *http.Client
		request *http.Request
		Cookies []*http.Cookie

		// default to true
		allowRedirect bool
		// default to 0
		timeout int64
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

// Request is the base method
func (s *Session) Request(method string, urlStr string, options *H) (*Response, error) {
	method = strings.ToUpper(method)

	switch method {
	case "HEAD", "GET", "POST", "DELETE", "OPTIONS", "PUT", "PATCH",
		"CONNECT", "TRACE":
		// urlencode the query string
		urlStrParsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		urlStrParsed.RawQuery = urlStrParsed.Query().Encode()

		s.request, _ = http.NewRequest(method, urlStrParsed.String(), nil)
		s.request.Header.Set("User-Agent", userAgent)

		// using one session multiply
		// https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
		s.request.Close = true
		s.Client = &http.Client{}
		for _, cookie := range s.Cookies {
			s.request.AddCookie(cookie)
		}

		// add options of Request struct
		err = addOptions(s.request, options)
		if err != nil {
			return nil, err
		}

		// add options of Client struct
		if options != nil {
			if !options.AllowRedirect {
				s.Client.CheckRedirect = disableRedirect
			}
			s.Client.Timeout = time.Duration(options.Timeout) * time.Second

			if options.Proxy != "" {
				urli := url.URL{}
				urlproxy, err := urli.Parse(options.Proxy)
				if err != nil {
					return nil, err
				}
				s.Client.Transport = &http.Transport{
					Proxy: http.ProxyURL(urlproxy),
				}
			}
		}

	default:
		return nil, ErrInvalidMethod
	}

	// do request and parse response
	resp, err := s.Client.Do(s.request)
	if err != nil {
		return nil, err
	}

	retResp := &Response{
		resp,
		"utf-8",
		"",
		[]byte{},
	}
	err = retResp.bytes()
	if err != nil {
		return nil, err
	}
	retResp.text()

	// allowRedirect and timeout will be restored to the default value after every request
	s.Client.CheckRedirect = defaultCheckRedirect
	s.Client.Timeout = 0

	// store cookies in the session structure
	s.Cookies = append(s.Cookies, resp.Cookies()...)
	return retResp, nil
}

// ClearCookies deletes all cookies
func (s *Session) ClearCookies() {
	s.Cookies = []*http.Cookie{}
}

// GetRequest returns nic.Session.request
func (s *Session) GetRequest() *http.Request {
	return s.request
}

// Get is a shortcut for get method
func (s *Session) Get(url string, options *H) (*Response, error) {
	return s.Request("get", url, options)
}

// Post is a shortcut for get method
func (s *Session) Post(url string, options *H) (*Response, error) {
	return s.Request("post", url, options)
}

// Head is a shortcut for get method
func (s *Session) Head(url string, options *H) (*Response, error) {
	return s.Request("head", url, options)
}

// Delete is a shortcut for get method
func (s *Session) Delete(url string, options *H) (*Response, error) {
	return s.Request("delete", url, options)
}

// Options is a shortcut for get method
func (s *Session) Options(url string, options *H) (*Response, error) {
	return s.Request("options", url, options)
}

// Put is a shortcut for get method
func (s *Session) Put(url string, options *H) (*Response, error) {
	return s.Request("put", url, options)
}

// Patch is a shortcut for get method
func (s *Session) Patch(url string, options *H) (*Response, error) {
	return s.Request("patch", url, options)
}

// Connect is a shortcut for get method
func (s *Session) Connect(url string, options *H) (*Response, error) {
	return s.Request("connect", url, options)
}

// Trace is a shortcut for get method
func (s *Session) Trace(url string, options *H) (*Response, error) {
	return s.Request("trace", url, options)
}
