package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/terrycain/aws_ecr_proxy/internal/ecr_token"
	"github.com/terrycain/aws_ecr_proxy/internal/proxy_server"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting ECR Proxy")
	awsSession, err := session.NewSession()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create an AWS session, check credentials")
		panic(err)
	}

	// Create an ECR client
	svc := ecr.New(awsSession)

	// Instantiate our ecs token getter
	tokenFetcher := ecr_token.New(svc)
	go tokenFetcher.Run()
	defer tokenFetcher.Close()

	// Pass reference to our refreshing token to the HTTP server
	proxy_server.Run(tokenFetcher)
}
