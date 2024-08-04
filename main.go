package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	aws "oncall-notify/pkg/config/aws"
	slack "oncall-notify/pkg/config/slack"
	"oncall-notify/pkg/env"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	jsoniter "github.com/json-iterator/go"
)

var config *Config

type Config struct {
	AwsConfig   *aws.AwsConfig
	SlackConfig *slack.SlackConfig
}

func NewConfig(awsConfig *aws.AwsConfig, slackConfig *slack.SlackConfig) *Config {
	return &Config{
		AwsConfig:   awsConfig,
		SlackConfig: slackConfig,
	}
}

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

	config = NewConfig(
		awsConfig,
		slackConfig,
	)

	lambda.Start(handleRequest)
}

func printSupportEngineerForWeek() string {
	contactIds, err := config.AwsConfig.LoadAndPrintAllRotationShifts(context.Background())
	if err != nil {
		log.Fatalf("Error loading rotation shifts: %v\n", err)
		return ""
	}

	for _, contactId := range contactIds {
		contactName := strings.Split(contactId, "/")[1]
		if userID, ok := config.SlackConfig.UserIDs[contactName]; ok {
			return fmt.Sprintf("<!subteam^%s|%s> this week is <@%s> \n", config.SlackConfig.SubTeamID, config.SlackConfig.SubTeamName, userID)
		}
	}

	return ""
}

func updateUserGroupUsers(userGroupID string, userIDs []string) error {
	data := url.Values{}
	data.Set("token", config.SlackConfig.ApiToken)
	data.Set("usergroup", userGroupID)
	data.Set("users", strings.Join(userIDs, ","))

	if _, err := http.PostForm(config.SlackConfig.ApiUrl, data); err != nil {
		return err
	}

	return nil
}

func handleRequest(ctx context.Context) (string, error) {
	message := printSupportEngineerForWeek()
	fmt.Println(message)

	if message != "" {
		// Extract user IDs from the userIDs map
		var userIdList []string
		for _, userId := range config.SlackConfig.UserIDs {
			userIdList = append(userIdList, userId)
		}

		// Update user group users
		if err := updateUserGroupUsers(config.SlackConfig.SubTeamID, userIdList); err != nil {
			log.Fatalf("Error updating user group users: %v\n", err)
		}

		slackBody, _ := jsoniter.Marshal(map[string]string{"text": message})
		response, err := http.Post(config.SlackConfig.WebhookUrl, "application/json", bytes.NewBuffer(slackBody))
		if err != nil {
			log.Fatalf("Error sending message to Slack: %v\n", err)
		}

		if response.StatusCode != 200 {
			log.Fatalf("Error sending message to Slack: received non-200 response: %d\n", response.StatusCode)
		}
	}

	return message, nil
}
