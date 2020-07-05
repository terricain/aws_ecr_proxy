package ecr_token

import (
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/rs/zerolog"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

type MockEcr struct {
	req     *ecr.GetAuthorizationTokenInput
	resp    *ecr.GetAuthorizationTokenOutput
	respErr error
}

func (m *MockEcr) GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error) {
	m.req = input
	return m.resp, m.respErr
}

func TestNew(t *testing.T) {
	mockEcr := &MockEcr{}
	newFetcher := New(mockEcr)

	if newFetcher.EcrClient != mockEcr {
		t.Error("Expected ECR client provided to be stored in struct")
	}
}

func TestFetch(t *testing.T) {
	token := "test1234"
	expire := time.Now().Add(5 * time.Hour)
	endpoint := "https://ecrendpoint"
	mockEcr := &MockEcr{resp: &ecr.GetAuthorizationTokenOutput{AuthorizationData: []*ecr.AuthorizationData{
		{
			AuthorizationToken: &token,
			ExpiresAt:          &expire,
			ProxyEndpoint:      &endpoint,
		},
	}}}
	newFetcher := New(mockEcr)

	go newFetcher.Run()

	// Should move the main processing logic in the for loop to a separate function
	<-time.After(500 * time.Millisecond)

	newFetcher.Close()
	if !newFetcher.closed {
		t.Error("Fetcher failed to close")
	}

	if newFetcher.Token != token {
		t.Errorf("Invalid token, got: %s expected %s", newFetcher.Token, token)
	}
	if newFetcher.Endpoint != endpoint {
		t.Errorf("Invalid endpoint, got: %s expected %s", newFetcher.Endpoint, endpoint)
	}
	if newFetcher.ExpiresAt != expire {
		t.Errorf("Invalid token expiry, got: %v expected %v", newFetcher.ExpiresAt, expire)
	}

}
