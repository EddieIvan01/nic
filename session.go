package nic

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	// Session is the wrapper for http.Client and http.Request
	Session struct {
		client  *http.Client
		request *http.Request
		cookies []*http.Cookie

		// default to false
		allowRedirect bool
		// default to 0
		timeout int64
	}
)

var (
	disableRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
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

		// https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
		s.request.Close = true
		s.client = &http.Client{}
		for _, cookie := range s.cookies {
			s.request.AddCookie(cookie)
		}
		// add options of Request struct
		err = addOptions(s.request, options)
		if err != nil {
			return nil, err
		}

		// add options of Client struct
		if options != nil {
			// allowRedirect and timeout hould be restored to the default value
			// after every request
			if !options.AllowRedirect {
				s.client.CheckRedirect = disableRedirect
			}
			s.client.Timeout = time.Duration(options.Timeout) * time.Second

			if options.Proxy != "" {
				urli := url.URL{}
				urlproxy, err := urli.Parse(options.Proxy)
				if err != nil {
					return nil, err
				}
				s.client.Transport = &http.Transport{
					Proxy: http.ProxyURL(urlproxy),
				}
			}
		}

	default:
		return nil, ErrInvalidMethod
	}

	resp, err := s.client.Do(s.request)
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

	s.client.CheckRedirect = disableRedirect
	s.client.Timeout = 0
	s.cookies = append(s.cookies, resp.Cookies()...)
	return retResp, nil
}

// ClearCookies deletes all cookies
func (s *Session) ClearCookies() {
	s.cookies = []*http.Cookie{}
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
