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
)

type (
	// H struct is options for request and http client
	H struct {
		AllowRedirect bool
		Timeout       int64
		Data          KV
		Raw           string
		Headers       KV
		Cookies       KV
		Auth          KV
		Proxy         string

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

// could only contains one of Data, Raw, Files, Json
func (h *H) isConflict() bool {
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

//========================================================
// functions for adding options
// vvvvvvvvvvvvvvvvvvvvv
//========================================================
func addData(req *http.Request, d KV) error {
	data := ""
	for k, v := range d {
		k = url.QueryEscape(k)

		v, ok := v.(string)
		if !ok {
			return fmt.Errorf("nic: %v param must be string type", v)
		}
		v = url.QueryEscape(v)
		data = fmt.Sprintf("%s&%s=%s", data, k, v)
	}

	data = data[1:]
	req.Body = ioutil.NopCloser(strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return nil
}

func addFiles(req *http.Request, f F) error {
	for name, fileInfo := range f {
		filenameI := fileInfo["filename"]

		filename, ok := filenameI.(string)
		if !ok {
			return fmt.Errorf("nic: filename %v must be string type", filenameI)
		}

		if len(fileInfo) < 1 || filename == "" {
			return ErrFileInfo
		}

		fp, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fp.Close()

		buffer := &bytes.Buffer{}
		writer := multipart.NewWriter(buffer)
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
				v, ok := v.(string)
				if !ok {
					return fmt.Errorf("nic: %v param must be string type", v)
				}
				_ = writer.WriteField(k, v)
			}
		}
		err = writer.Close()
		if err != nil {
			return err
		}

		req.Body = ioutil.NopCloser(buffer)
		contentType := writer.FormDataContentType()
		req.Header.Set("Content-Type", contentType)
	}
	return nil
}

func addJSON(req *http.Request, j KV) error {
	jsonV, err := json.Marshal(j)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonV))
	req.Header.Set("Content-Type", "application/json")
	return nil
}

//========================================================
// ^^^^^^^^^^^^^^^^^^^^^
// functions for adding options
//========================================================

func addOptions(req *http.Request, h *H) error {
	// add option to request
	// data, header, cookie, auth, file, json
	if h == nil {
		return nil
	}
	if h.isConflict() {
		return ErrParamConflict
	}

	if h.Data != nil {
		addData(req, h.Data)
	}

	if h.Raw != "" {
		req.Body = ioutil.NopCloser(strings.NewReader(h.Raw))
	}

	if h.Headers != nil {
		for headerK, headerV := range h.Headers {
			headerV, ok := headerV.(string)
			if !ok {
				return fmt.Errorf("nic: header %v must be string type", headerV)
			}
			req.Header.Add(headerK, headerV)
		}
	}

	if h.Cookies != nil {
		req.Header.Set("Cookies", "")
		for cookieK, cookieV := range h.Cookies {
			cookieV, ok := cookieV.(string)
			if !ok {
				return fmt.Errorf("nic: cookie %v must be string type", cookieV)
			}
			c := &http.Cookie{
				Name:  cookieK,
				Value: cookieV,
			}
			req.AddCookie(c)
		}
	}

	if h.Auth != nil {
		for k, v := range h.Auth {
			v, ok := v.(string)
			if !ok {
				return fmt.Errorf("nic: basic-auth %v must be string type", v)
			}
			req.SetBasicAuth(k, v)
		}
	}

	if h.Files != nil {
		err := addFiles(req, h.Files)
		if err != nil {
			return err
		}
	}

	if h.JSON != nil {
		err := addJSON(req, h.JSON)
		if err != nil {
			return err
		}
	}

	return nil
}
