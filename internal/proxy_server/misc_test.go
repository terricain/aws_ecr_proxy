package proxy_server

import (
	"encoding/json"
	"github.com/terricain/aws_ecr_proxy/internal/ecr_token"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Healthz)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("healthz handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	//// Check the response body is what we expect.
	//expected := `{"alive": true}`
	//if rr.Body.String() != expected {
	//	t.Errorf("handler returned unexpected body: got %v want %v",
	//		rr.Body.String(), expected)
	//}
}

func TestReadyz(t *testing.T) {
	tables := []struct {
		token          string
		endpoint       string
		expectedStatus int
	}{
		{"sometoken", "http://ecrendpoint", 200},
		{"", "", 500},
	}

	for _, table := range tables {
		webData := WebData{&ecr_token.EcrFetcher{Token: table.token, Endpoint: table.endpoint}, nil}

		req, err := http.NewRequest("GET", "/readyz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webData.Readyz)
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != table.expectedStatus {
			t.Errorf("readyz handler returned wrong status code with token \"%v\": got %v want %v",
				table.token, status, table.expectedStatus)
		}

	}
}

func TestVersion(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Version)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("version handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	jsonResult := make(map[string]string)

	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResult); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"build_date", "sha", "version"} {
		if _, exists := jsonResult[key]; !exists {
			t.Errorf("version endpoint JSON missing \"%s\" field", key)
		}
	}
}
