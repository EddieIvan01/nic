package nic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/axgle/mahonia"
)

// Response is the wrapper for http.Response
type Response struct {
	*http.Response
	encoding string
	Text     string
	Bytes    []byte
}

func (r *Response) text() {
	r.Text = string(r.Bytes)
}

func (r *Response) bytes() error {
	data, err := ioutil.ReadAll(r.Body)
	// for multiple reading
	// e.g. goquery.NewDocumentFromReader
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	r.Bytes = data
	return err
}

// JSON could parse http json response
// if is not a json response, returns ErrNotJsonResponse
func (r *Response) JSON(s interface{}) error {
	// JSON response not must be `application/json` type
	// maybe `text/plain`, `text/html`...etc.
	/*
		cType := r.Header.Get("Content-Type")
		if !strings.Contains(cType, "json") {
			return ErrNotJsonResponse
		}
	*/
	err := json.Unmarshal(r.Bytes, s)
	return err
}

// SetEncode changes Response.encoding
// and it changes Response.Text every times be invoked
func (r *Response) SetEncode(e string) error {
	if r.encoding != e {
		r.encoding = strings.ToLower(e)
		decoder := mahonia.NewDecoder(e)
		if decoder == nil {
			return ErrUnrecognizedEncoding
		}
		r.Text = decoder.ConvertString(r.Text)
	}
	return nil
}

// GetEncode returns Response.encoding
func (r *Response) GetEncode() string {
	return r.encoding
}

func (r *Response) Read(x []byte) (int, error) {
	fmt.Println(len(x), cap(x))
	fmt.Println(len(r.Bytes), cap(r.Bytes))
	fmt.Println(r.Bytes[4095])
	for i := 0; i < len(x) && i < len(r.Bytes); i++ {
		//fmt.Println(i)
		x[i] = r.Bytes[i]
	}
	return len(r.Bytes), nil
}
