package main

import (
	"context"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

var credentials aws.CredentialsProviderFunc = func(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{AccessKeyID: "ABC", SecretAccessKey: "EFG", CanExpire: false}, nil
}

func lambdaEndpointResolver(cfg *settings.Config) aws.EndpointResolverWithOptionsFunc {
	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               cfg.LambdaEndpoint,
			HostnameImmutable: true,
		}, nil
	}
}

func NewLambdaClient(cfg *settings.Config) *lambda.Client {
	config := aws.Config{
		Region:                      "us-west-2",
		Credentials:                 credentials,
		EndpointResolverWithOptions: lambdaEndpointResolver(cfg),
		ClientLogMode:               0,
		DefaultsMode:                "",
		RuntimeEnvironment:          aws.RuntimeEnvironment{},
	}

	return lambda.NewFromConfig(config)
}
