package nic

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	// H struct is options for request and http client
	H struct {
		Params  KV
		Data    KV
		Raw     string
		Headers KV
		Cookies KV
		Auth    KV
		Proxy   string
		JSON    KV
		Files   KV

		AllowRedirect      bool
		Timeout            int64
		Chunked            bool
		DisableKeepAlives  bool
		DisableCompression bool
		SkipVerifyTLS      bool
	}

	// KV is used for H struct
	KV map[string]interface{}

	// when upload a file, we use nic.KV again
	// nic.File returns F struct
	//
	// nic.KV {
	//     "file1" :"file" : nic.FileFromPath("test.go"),
	//     "file2" : nic.File("test.go", []byte("package nic")).
	// 					FName("nic.go").
	//					MIME("text/plain"),
	//     "token" : "abc",
	// }
	//
	//
	// the POST body is:
	//
	// Content-Type: multipart/form-data; boundary=e7d105eae032bdc774a787f1d874269d04499cb284477d6d77889be73caf
	//
	// --e7d105eae032bdc774a787f1d874269d04499cb284477d6d77889be73caf
	// Content-Disposition: form-data; name="file1"; filename="test.go"
	// Content-Type: application/octet-stream
	//
	// package test
	// --e7d105eae032bdc774a787f1d874269d04499cb284477d6d77889be73caf
	// Content-Disposition: form-data; name="token"
	//
	// abc
	// --e7d105eae032bdc774a787f1d874269d04499cb284477d6d77889be73caf
	// Content-Disposition: form-data; name="file2"; filename="nic.go"
	// Content-Type: text/plain
	//
	// package test
	// --e7d105eae032bdc774a787f1d874269d04499cb284477d6d77889be73caf--
	//
	//
	// F struct saves file form information
	F struct {
		Src      []byte
		FilePath string
		FileName string
		MimeType string
	}
)

// File returns a new file struct
func File(filename string, src []byte) *F {
	return &F{
		Src:      src,
		FileName: filename,
	}
}

// FileFromPath returns a file struct from file path
func FileFromPath(path string) *F {
	return &F{
		FilePath: path,
		FileName: filepath.Base(path),
	}
}

// FName changes file's filename in multipart form
// invoke it in a chain
func (f *F) FName(filename string) *F {
	f.FileName = filename
	return f
}

// MIME changes file's mime type in multipart form
// invoke it in a chain
func (f *F) MIME(mimetype string) *F {
	f.MimeType = mimetype
	return f
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

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

func setQuery(req *http.Request, p KV) error {
	originURL := req.URL
	extendQuery := make([]byte, 0)

	for k, v := range p {
		kEscaped := url.QueryEscape(k)
		vs, ok := v.(string)
		if !ok {
			return fmt.Errorf("nic: query param %v[%T] must be string type", v, v)
		}
		vEscaped := url.QueryEscape(vs)

		extendQuery = append(extendQuery, '&')
		extendQuery = append(extendQuery, []byte(kEscaped)...)
		extendQuery = append(extendQuery, '=')
		extendQuery = append(extendQuery, []byte(vEscaped)...)
	}

	// trim the `&`
	if originURL.RawQuery == "" {
		extendQuery = extendQuery[1:]
	}

	originURL.RawQuery += string(extendQuery)
	return nil
}

func setData(req *http.Request, d KV, chunked bool) error {
	data := ""
	for k, v := range d {
		k = url.QueryEscape(k)

		vs, ok := v.(string)
		if !ok {
			return fmt.Errorf(
				"nic: post data %v[%T] must be string type", v, v)
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

func setFiles(req *http.Request, files KV, chunked bool) error {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	for name, value := range files {
		switch value := value.(type) {
		case *F:
			mimetype := value.MimeType
			if mimetype == "" {
				mimetype = "application/octet-stream"
			}

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition",
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
					escapeQuotes(name), escapeQuotes(value.FileName)))
			h.Set("Content-Type", mimetype)

			part, err := writer.CreatePart(h)
			if err != nil {
				return err
			}

			if len(value.Src) != 0 {
				_, err = part.Write(value.Src)
				if err != nil {
					return err
				}
			} else {
				fp, err := os.Open(value.FilePath)
				if err != nil {
					return err
				}
				defer fp.Close()

				_, err = io.Copy(part, fp)
				if err != nil {
					return err
				}
			}

		case string:
			err := writer.WriteField(name, value)
			if err != nil {
				return err
			}

		default:
			return ErrFileInfo
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

// set option for http.Request
// data, header, cookie, auth, file, json
func (h H) setRequestOpt(req *http.Request) error {
	if h.isConflict() {
		return ErrParamConflict
	}

	if h.Params != nil {
		err := setQuery(req, h.Params)
		if err != nil {
			return err
		}
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
				return fmt.Errorf(
					"nic: header %v[%T] must be string type",
					headerV, headerV)
			}
			req.Header.Set(headerK, headerVS)
		}
	}

	if h.Cookies != nil {
		for cookieK, cookieV := range h.Cookies {
			cookieVS, ok := cookieV.(string)
			if !ok {
				return fmt.Errorf(
					"nic: cookie %v[%T] must be string type",
					cookieV, cookieV)
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
				return fmt.Errorf(
					"nic: basic-auth %v[%T] must be string type",
					v, v)
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

// set option for http.Client
// proxy, timeout, redirect
func (h H) setClientOpt(client *http.Client) error {
	if !h.AllowRedirect {
		client.CheckRedirect = disableRedirect
	}

	client.Timeout = time.Duration(h.Timeout) * time.Second

	transport := client.Transport.(*http.Transport)
	transport.DisableKeepAlives = h.DisableKeepAlives
	transport.DisableCompression = h.DisableCompression

	if h.SkipVerifyTLS {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	if h.Proxy != "" {
		urli := url.URL{}
		urlproxy, err := urli.Parse(h.Proxy)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(urlproxy)
	}

	return nil
}
