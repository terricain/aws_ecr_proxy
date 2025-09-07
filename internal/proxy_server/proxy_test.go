package proxy_server

import (
	"bytes"
	"github.com/terricain/aws_ecr_proxy/internal/ecr_token"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFixLinkHeader(t *testing.T) {
	header := `<https://000000000000.dkr.ecr.eu-west-2.amazonaws.com/v2/test/tags/list?last=somekey>; rel="next"`
	expectedHdr := `<http://localhost:8080/v2/test/tags/list?last=somekey>; rel="next"`
	resultantHdr, err := FixLinkHeader("http", "localhost:8080", header)
	if err != nil {
		t.Fatalf("Failed to fix link header: %s", err.Error())
	}

	if resultantHdr != expectedHdr {
		t.Errorf("Expected link header: \"%s\" got \"%s\"", expectedHdr, resultantHdr)
	}
}

func TestCopyHeaders(t *testing.T) {
	webData := WebData{fetcher: &ecr_token.EcrFetcher{Token: "test1234"}, httpClient: nil}
	expectedAuthHeader := "Basic test1234"

	sourceReq := http.Request{Header: map[string][]string{
		"Data": {"test1"},
	}}
	destReq := http.Request{Header: map[string][]string{}}

	webData.CopyHeaders(&sourceReq, &destReq)

	if value := destReq.Header.Get("Data"); value != "test1" {
		t.Errorf("destination request header Data not copied from source")
	}

	if value := destReq.Header.Get("Authorization"); value != expectedAuthHeader {
		t.Errorf("destination request header Authorization not correct, got \"%s\" wanted \"%s\"", value, expectedAuthHeader)
	}

}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestHandler(t *testing.T) {
	requestPath := "/v2/test/tags/list"

	webData := WebData{
		fetcher: &ecr_token.EcrFetcher{Token: "test1234", Endpoint: "https://ecrendpoint"},
		httpClient: NewTestClient(func(req *http.Request) *http.Response {

			if req.URL.Scheme != "https" {
				t.Fatalf("Incorrect proxied request scheme, got: \"%s\" expected: \"%s\"", req.URL.Scheme, "https")
			}
			if req.URL.Host != "ecrendpoint" {
				t.Fatalf("Incorrect proxied request host, got: \"%s\" expected: \"%s\"", req.URL.Host, "ecrendpoint")
			}
			if req.URL.Path != requestPath {
				t.Fatalf("Incorrect proxied request path, got: \"%s\" expected: \"%s\"", req.URL.Path, requestPath)
			}

			return &http.Response{
				StatusCode: 234,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				Header: http.Header{
					"Data": {"test1"},
					"Link": {"<https://000000000000.dkr.ecr.eu-west-2.amazonaws.com/v2/test/tags/list?last=somekey>; rel=\"next\""},
				},
			}
		}),
	}

	req, err := http.NewRequest("GET", requestPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = requestPath
	req.Host = "localhost:8080"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webData.Handler)
	handler.ServeHTTP(rr, req)

	if rr.Code != 234 {
		t.Errorf("Incorrect status code, got: %d expected %d", rr.Code, 234)
	}
	if body := rr.Body.String(); body != "OK" {
		t.Errorf("Incorrect response body, got: %s expected %s", body, "OK")
	}
	linkHdr := rr.Header().Get("Link")
	expected := `<http://localhost:8080/v2/test/tags/list?last=somekey>; rel="next"`
	if linkHdr != expected {
		t.Errorf("Incorrect response body, got: %s expected %s", linkHdr, expected)
	}
}
