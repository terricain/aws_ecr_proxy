package ecr_token

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"time"
	"github.com/rs/zerolog/log"
)

type EcrClient interface {
	GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error)
}

type EcrFetcher struct {
	EcrClient EcrClient
	Token string
	ExpiresAt time.Time
	Endpoint string

	closeChannel chan bool
	closed bool
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
		expireTime := e.ExpiresAt.Add(- 15*time.Minute)
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
			case <-time.After(5*time.Minute):
				continue runLoop
			}
		}

		log.Info().Msg("Getting ECR credential")

		input := &ecr.GetAuthorizationTokenInput{}

		result, err := e.EcrClient.GetAuthorizationToken(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case ecr.ErrCodeServerException:
					log.Error().Err(aerr).Msg("AWS SKD ServerException")
				case ecr.ErrCodeInvalidParameterException:
					log.Error().Err(aerr).Msg("AWS SKD InvalidParameterException")
				default:
					log.Error().Err(aerr).Msg("Failed to get ECR token")
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.Error().Err(err).Msg("Failed to get ECR token")
			}

			// Sleeping for a few seconds to prevent spamming the ECR API
			<-time.After(15*time.Second)
			continue
		}

		e.ExpiresAt = *result.AuthorizationData[0].ExpiresAt
		e.Token = *result.AuthorizationData[0].AuthorizationToken
		e.Endpoint = *result.AuthorizationData[0].ProxyEndpoint

		log.Info().Msgf("Successfully retrieved ECR token for %s which expires on %s", e.Endpoint, e.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	e.closeChannel <- true
}