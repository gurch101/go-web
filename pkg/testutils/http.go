package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gurch101.github.io/go-web/pkg/httputils"
)

func createRequestWithBody(t *testing.T, method, url string, payload interface{}) *http.Request {
	t.Helper()

	requestBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(requestBody))
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func CreatePostRequest(t *testing.T, url string, payload interface{}) *http.Request {
	t.Helper()

	return createRequestWithBody(t, http.MethodPost, url, payload)
}

func CreatePatchRequest(t *testing.T, url string, payload interface{}) *http.Request {
	t.Helper()

	return createRequestWithBody(t, http.MethodPatch, url, payload)
}

func CreateGetRequest(url string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func CreateDeleteRequest(url string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func AssertError(t *testing.T, resp map[string]interface{}, expectedErrorField string, expectedErrorMessage string) {
	t.Helper()

	err, ok := resp["errors"].([]interface{})[0].(map[string]interface{})
	if !ok {
		t.Errorf("expected error; got %v", resp)
	}

	errorKey, ok := err["field"]
	if !ok || errorKey != expectedErrorField {
		t.Errorf("expected error field %s; got %v", expectedErrorField, resp)
	}

	errorMessage, ok := err["message"]
	if !ok || errorMessage != expectedErrorMessage {
		t.Errorf("expected error message %s; got %v", expectedErrorMessage, resp)
	}
}
