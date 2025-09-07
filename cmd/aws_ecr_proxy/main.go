package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/terricain/aws_ecr_proxy/internal/ecr_token"
	"github.com/terricain/aws_ecr_proxy/internal/proxy_server"
	"github.com/terricain/aws_ecr_proxy/internal/utils"
	"github.com/terricain/aws_ecr_proxy/internal/version"
)

func main() {
	envLevel := utils.GetEnv("LOG_LEVEL", "INFO")
	zerolog.SetGlobalLevel(utils.LogNameToLevel(envLevel))

	disableProxyHeaders := utils.GetEnv("DISABLE_PROXY_HEADERS", "false") == "true"
	host := utils.GetEnv("LISTEN_HOST", "0.0.0.0")
	port := utils.GetEnv("LISTEN_PORT", "8080")
	addr := host + ":" + port

	log.Info().Str("version", version.VERSION).Str("build_date", version.BUILDDATE).Str("sha", version.SHA).Msg("Starting ECR Proxy")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create an AWS session, check credentials")
		panic(err)
	}

	svc := ecr.NewFromConfig(cfg)

	// Instantiate our ecs token getter
	tokenFetcher := ecr_token.New(svc)
	go tokenFetcher.Run()
	defer tokenFetcher.Close()

	// Pass reference to our refreshing token to the HTTP server
	proxy_server.Run(addr, disableProxyHeaders, tokenFetcher)
}
