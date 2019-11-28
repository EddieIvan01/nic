package nic

import (
	"errors"
)

const (
	version   = "0.3"
	userAgent = "golang-nic/0.3"
	author    = "iv4n"
	copyright = "Copyright 2019 iv4n"
)

var (
	// ErrInvalidMethod will be throwed when method not in
	// [HEAD, GET, POST, DELETE, OPTIONS, PUT, PATCH, CONNECT, TRACE]
	ErrInvalidMethod = errors.New("nic: Method is invalid")

	// ErrFileInfo will be throwed when fileinfo is invalid
	ErrFileInfo = errors.New("nic: Invalid file information")

	// ErrParamConflict will be throwed when options params conflict
	// e.g. files + data
	//      json + data
	//      ...
	ErrParamConflict = errors.New("nic: Options param conflict")

	// ErrUnrecognizedEncoding will be throwed while changing response encoding
	// if encoding is not recognized
	ErrUnrecognizedEncoding = errors.New("nic: Unrecognized encoding")

	// ErrNotJsonResponse will be throwed when response not a json
	// but invoke Json() method
	ErrNotJsonResponse = errors.New("nic: Not a Json response")

	// ErrHookFuncMaxLimit will be throwed when the number of hook functions
	// more than MaxLimit = 8
	ErrHookFuncMaxLimit = errors.New("nic: The number of hook functions must be less than 8")

	// ErrIndexOutofBound means the index out of bound
	ErrIndexOutofBound = errors.New("nic: Index out of bound")
)

const (
	HEAD    = "HEAD"
	GET     = "GET"
	POST    = "POST"
	DELETE  = "DELETE"
	OPTIONS = "OPTIONS"
	PUT     = "PUT"
	PATCH   = "PATCH"
)

// Get implemented by Session.Get
func Get(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Get(url, option)
}

// Post implemented by Session.Post
func Post(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Post(url, option)
}

// Head implemented by Session.Head
func Head(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Head(url, option)
}

// Delete implemented by Session.Delete
func Delete(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Delete(url, option)
}

// Options implemented by Session.Options
func Options(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Options(url, option)
}

// Put implemented by Session.Put
func Put(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Put(url, option)
}

// Patch implemented by Session.Patch
func Patch(url string, option Option) (*Response, error) {
	session := NewSession()
	return session.Patch(url, option)
}
