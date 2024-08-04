package main

import (
	"log"
	"oncall-notify/pkg/config/aws"
	"oncall-notify/pkg/config/slack"
	"oncall-notify/pkg/env"
	handler "oncall-notify/src"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	awsRegion := env.DefaultEnv("AWS_REGION", "eu-west-1")
	slackApiUrl := env.DefaultEnv("SLACK_API_URL", "https://slack.com/api/usergroups.users.update")
	slackWebhookUrl := env.DefaultEnv("SLACK_WEBHOOK_URL", "")
	slackApiToken := env.DefaultEnv("SLACK_API_TOKEN", "")
	slackSubTeamID := env.DefaultEnv("SLACK_SUBTEAM_ID", "")
	slackSubTeamName := env.DefaultEnv("SLACK_SUBTEAM_NAME", "support")
	slackUserIDs := make(map[string]string)

	if slackApiToken == "" {
		log.Fatalf("Error: Slack API token is not set")
	}

	awsConfig, err := aws.NewConfig(awsRegion)
	if err != nil {
		log.Fatalf("Error loading AWS config: %v\n", err)
	}

	// Create a new Slack config
	slackConfig := slack.NewConfig(
		slackApiUrl,
		slackWebhookUrl,
		slackApiToken,
		slackSubTeamID,
		slackSubTeamName,
		slackUserIDs,
	)

	// Configure the UserIDs map
	slackConfig.ConfigureUserIDs()
	if len(slackConfig.UserIDs) == 0 {
		log.Fatalf("Error: No user IDs found in Slack config")
	}

	handler.LambdaConfig = handler.NewConfig(
		awsConfig,
		slackConfig,
	)

	lambda.Start(handler.HandleRequest)
}
