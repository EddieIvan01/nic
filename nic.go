package nic

import (
	"errors"
)

const (
	version   = "0.1.1"
	userAgent = "golang-nic/0.1.1"
	author    = "Iv4n"
	copyright = "Copyright 2019 Iv4n"
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
)

// Get implemented by Session.Get
func Get(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Get(url, options)
}

// Post implemented by Session.Post
func Post(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Post(url, options)
}

// Head implemented by Session.Head
func Head(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Head(url, options)
}

// Delete implemented by Session.Delete
func Delete(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Delete(url, options)
}

// Options implemented by Session.Options
func Options(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Options(url, options)
}

// Put implemented by Session.Put
func Put(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Put(url, options)
}

// Patch implemented by Session.Patch
func Patch(url string, options *H) (*Response, error) {
	session := &Session{}
	return session.Patch(url, options)
}
