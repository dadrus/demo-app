package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShowIndexPageUnauthenticated(t *testing.T) {
	r := getRouter(true)

	r.GET("/", ShowIndexPage)

	req, _ := http.NewRequest("GET", "/", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOk := w.Code == http.StatusOK

		p, err := ioutil.ReadAll(w.Body)

		pageOk := err == nil && strings.Index(string(p), "<title>Home Page</title>") > 0

		return statusOk && pageOk
	})
}
