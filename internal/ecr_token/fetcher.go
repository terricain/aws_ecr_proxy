package ecr_token

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/rs/zerolog/log"
)

type EcrClient interface {
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

type EcrFetcher struct {
	EcrClient EcrClient
	Token     string
	ExpiresAt time.Time
	Endpoint  string

	closeChannel chan bool
	closed       bool
}

func New(ecrClient EcrClient) *EcrFetcher {
	return &EcrFetcher{EcrClient: ecrClient, closeChannel: make(chan bool)}
}

func (e *EcrFetcher) Close() {
	if e.closed {
		return
	}

	e.closeChannel <- true
	select {
	case <-e.closeChannel:
		break
	case <-time.After(2 * time.Second):
		log.Error().Msg("Failed to stop token fetcher in under 2 seconds, giving up")
		break
	}
	e.closed = true
}

func (e *EcrFetcher) Run() {
	if e.closed {
		return
	}

	// Logic here is un
runLoop:
	for {
		expireTime := e.ExpiresAt.Add(-15 * time.Minute)
		needsRenew := time.Now().After(expireTime)

		select {
		case <-e.closeChannel:
			break runLoop

		default:
			if needsRenew {
				break
			}

			select {
			case <-e.closeChannel:
				break runLoop
			case <-time.After(5 * time.Minute):
				continue runLoop
			}
		}

		log.Info().Msg("Getting ECR credential")

		result, err := getECRToken(e.EcrClient)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get ECR token")

			// Sleeping for a few seconds to prevent spamming the ECR API
			<-time.After(15 * time.Second)
			continue
		}

		e.ExpiresAt = *result.AuthorizationData[0].ExpiresAt
		e.Token = *result.AuthorizationData[0].AuthorizationToken
		e.Endpoint = *result.AuthorizationData[0].ProxyEndpoint

		log.Info().Msgf("Successfully retrieved ECR token for %s which expires on %s", e.Endpoint, e.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	e.closeChannel <- true
}

func getECRToken(ecrClient EcrClient) (*ecr.GetAuthorizationTokenOutput, error) {
	// Split into a function to scope context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
}
