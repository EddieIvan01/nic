package nic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	// H struct is options for request and http client
	H struct {
		Data    KV
		Raw     string
		Headers KV
		Cookies KV
		Auth    KV
		Proxy   string

		AllowRedirect bool
		Timeout       int64
		Chunked       bool

		JSON  KV
		Files F
	}

	// KV is used for H struct
	KV map[string]interface{}

	// F is for file-upload request
	// map[string]KV{
	//     "file1" : KV{
	//                  // path of file
	//                  "filename" : "1.txt",
	//                  "token" : "abc",
	//               },
	//     "file2" : KV{...},
	// }
	F map[string]KV
)

// Option is the interface implemented by `H` and `*H`
type Option interface {
	setRequestOpt(*http.Request) error
	setClientOpt(*http.Client) error
}

// could only contains one of Data, Raw, Files, Json
func (h H) isConflict() bool {
	count := 0
	if h.Data != nil {
		count++
	}
	if h.Raw != "" {
		count++
	}
	if h.Files != nil {
		count++
	}
	if h.JSON != nil {
		count++
	}
	return count > 1
}

func setData(req *http.Request, d KV, chunked bool) error {
	data := ""
	for k, v := range d {
		k = url.QueryEscape(k)

		vs, ok := v.(string)
		if !ok {
			return fmt.Errorf("nic: post data %v[%T] must be string type", v, v)
		}
		vs = url.QueryEscape(vs)
		data = fmt.Sprintf("%s&%s=%s", data, k, vs)
	}

	data = data[1:]
	v := strings.NewReader(data)
	req.Body = ioutil.NopCloser(v)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if !chunked {
		req.ContentLength = int64(v.Len())
	}

	return nil
}

func setFiles(req *http.Request, f F, chunked bool) error {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	for name, fileInfo := range f {
		filenameI := fileInfo["filename"]

		filename, ok := filenameI.(string)
		if !ok {
			return fmt.Errorf("nic: filename %v[%T] must be string type", filenameI, filenameI)
		}

		if len(fileInfo) < 1 || filename == "" {
			return ErrFileInfo
		}

		fp, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fp.Close()

		part, err := writer.CreateFormFile(name, filepath.Base(filename))
		if err != nil {
			return err
		}

		_, err = io.Copy(part, fp)
		if err != nil {
			return err
		}

		if len(fileInfo) > 1 {
			delete(fileInfo, "filename")
			for k, v := range fileInfo {
				vs, ok := v.(string)
				if !ok {
					return fmt.Errorf("nic: %v[%T] param must be string type", v, v)
				}
				_ = writer.WriteField(k, vs)
			}
		}
	}
	err := writer.Close()
	if err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(buffer)
	contentType := writer.FormDataContentType()
	req.Header.Set("Content-Type", contentType)
	if !chunked {
		req.ContentLength = int64(buffer.Len())
	}
	return nil
}

func setJSON(req *http.Request, j KV, chunked bool) error {
	jsonV, err := json.Marshal(j)
	if err != nil {
		return err
	}

	v := bytes.NewBuffer(jsonV)
	req.Body = ioutil.NopCloser(v)
	req.Header.Set("Content-Type", "application/json")
	if !chunked {
		req.ContentLength = int64(v.Len())
	}
	return nil
}

func (h H) setRequestOpt(req *http.Request) error {
	// set option to request
	// data, header, cookie, auth, file, json
	if h.isConflict() {
		return ErrParamConflict
	}

	if h.Data != nil {
		err := setData(req, h.Data, h.Chunked)
		if err != nil {
			return err
		}
	}

	if h.Raw != "" {
		v := strings.NewReader(h.Raw)
		req.Body = ioutil.NopCloser(v)
		if !h.Chunked {
			req.ContentLength = int64(v.Len())
		}
	}

	if h.Headers != nil {
		for headerK, headerV := range h.Headers {
			headerVS, ok := headerV.(string)
			if !ok {
				return fmt.Errorf("nic: header %v[%T] must be string type", headerV, headerV)
			}
			req.Header.Add(headerK, headerVS)
		}
	}

	if h.Cookies != nil {
		req.Header.Set("Cookies", "")
		for cookieK, cookieV := range h.Cookies {
			cookieVS, ok := cookieV.(string)
			if !ok {
				return fmt.Errorf("nic: cookie %v[%T] must be string type", cookieV, cookieV)
			}
			c := &http.Cookie{
				Name:  cookieK,
				Value: cookieVS,
			}
			req.AddCookie(c)
		}
	}

	if h.Auth != nil {
		for k, v := range h.Auth {
			vs, ok := v.(string)
			if !ok {
				return fmt.Errorf("nic: basic-auth %v[%T] must be string type", v, v)
			}
			req.SetBasicAuth(k, vs)
		}
	}

	if h.Files != nil {
		err := setFiles(req, h.Files, h.Chunked)
		if err != nil {
			return err
		}
	}

	if h.JSON != nil {
		err := setJSON(req, h.JSON, h.Chunked)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h H) setClientOpt(client *http.Client) error {
	if !h.AllowRedirect {
		client.CheckRedirect = disableRedirect
	}

	client.Timeout = time.Duration(h.Timeout) * time.Second

	if h.Proxy != "" {
		urli := url.URL{}
		urlproxy, err := urli.Parse(h.Proxy)
		if err != nil {
			return err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}
	}
	return nil
}
