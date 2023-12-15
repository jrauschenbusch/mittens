//Copyright 2019 Expedia, Inc.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package http

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"mittens/internal/pkg/placeholders"
	"strings"
)

// Request represents an HTTP request.
type Request struct {
	Method string
	Path   string
	Body   io.Reader
}

var allowedHTTPMethods = map[string]interface{}{
	"GET":     nil,
	"HEAD":    nil,
	"POST":    nil,
	"PUT":     nil,
	"PATCH":   nil,
	"DELETE":  nil,
	"CONNECT": nil,
	"OPTIONS": nil,
	"TRACE":   nil,
}

//
// ToHTTPRequest parses an HTTP request which is in a string format and stores it in a struct.
func ToHTTPRequest(requestString string, gzipCompression bool) (Request, error) {
	parts := strings.SplitN(requestString, ":", 3)
	if len(parts) < 2 {
		return Request{}, fmt.Errorf("invalid request flag: %s, expected format <http-method>:<path>[:body]", requestString)
	}

	method := strings.ToUpper(parts[0])
	_, ok := allowedHTTPMethods[method]
	if !ok {
		return Request{}, fmt.Errorf("invalid request flag: %s, method %s is not supported", requestString, method)
	}

	// <method>:<path>
	if len(parts) == 2 {
		path := placeholders.InterpolatePlaceholders(parts[1])

		return Request{
			Method: method,
			Path:   path,
			Body:   nil,
		}, nil
	}

	path := placeholders.InterpolatePlaceholders(parts[1])
	// the body of the request can either be inlined, or come from a file
	rawBody, err := placeholders.GetBodyFromFileOrInlined(parts[2])
	if err != nil {
		return Request{}, fmt.Errorf("unable to parse body for request: %s", parts[2])
	}
	var interpolatedBody = placeholders.InterpolatePlaceholders(*rawBody)

	var reader io.Reader
	if gzipCompression == true {
		reader = GZipCompressBody(interpolatedBody)
	} else {
		reader = bytes.NewBufferString(interpolatedBody)
	}

	return Request{
		Method: method,
		Path:   path,
		Body:   reader,
	}, nil
}

func GZipCompressBody(body string) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		gz := gzip.NewWriter(pw)
		_, err := gz.Write([]byte(body))
		gz.Close()
		pw.CloseWithError(err)
	}()
	return pr
}
