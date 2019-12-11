// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package rest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func Test_handler_uploadHandler(t *testing.T) {
	h := handler{
		dir: "data",
	}

	origTestHookCheckZipFile := testHookCheckZipDir
	defer func() { testHookCheckZipDir = origTestHookCheckZipFile }()
	testHookCheckZipDir = func() {
		// TODO write your code here to check if zip files exists in right dir
		if _, err := os.Stat("data/france.zip"); os.IsNotExist(err) {
			// path/to/whatever does not exis
			fmt.Println(err)
			t.Fatal(err)
		}
	}
	// Use Client & URL from our local test server
	request, err := newfileUploadRequest("/api/v1/import", "file", "testdata/france.zip")
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := httprouter.New()
	// TODO add a import handler here
	router.POST("/api/v1/import", h.uploadHandler)

	// Populate the request's context with our test data.
	router.ServeHTTP(rr, request)
	// Check the status code is what we expect.
	err = os.RemoveAll("data")
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func newfileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
