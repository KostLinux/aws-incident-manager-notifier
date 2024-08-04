package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	aws "oncall-notify/pkg/config/aws"
	slack "oncall-notify/pkg/config/slack"

	jsoniter "github.com/json-iterator/go"
)

var LambdaConfig *Config

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

func PrintSupportEngineerForWeek() string {
	contactIds, err := LambdaConfig.AwsConfig.LoadAndPrintAllRotationShifts(context.Background())
	if err != nil {
		log.Fatalf("Error loading rotation shifts: %v\n", err)
		return ""
	}

	for _, contactId := range contactIds {
		contactName := strings.Split(contactId, "/")[1]
		if userID, ok := LambdaConfig.SlackConfig.UserIDs[contactName]; ok {
			return fmt.Sprintf("<!subteam^%s|%s> this week is <@%s> \n", LambdaConfig.SlackConfig.SubTeamID, LambdaConfig.SlackConfig.SubTeamName, userID)
		}
	}

	return ""
}

func UpdateUserGroupUsers(userGroupID string, userIDs []string) error {
	data := url.Values{}
	data.Set("token", LambdaConfig.SlackConfig.ApiToken)
	data.Set("usergroup", userGroupID)
	data.Set("users", strings.Join(userIDs, ","))

	if _, err := http.PostForm(LambdaConfig.SlackConfig.ApiUrl, data); err != nil {
		return err
	}

	return nil
}

func HandleRequest(ctx context.Context) (string, error) {
	message := PrintSupportEngineerForWeek()
	fmt.Println(message)

	if message != "" {
		// Extract user IDs from the userIDs map
		var userIdList []string
		for _, userId := range LambdaConfig.SlackConfig.UserIDs {
			userIdList = append(userIdList, userId)
		}

		// Update user group users
		if err := UpdateUserGroupUsers(LambdaConfig.SlackConfig.SubTeamID, userIdList); err != nil {
			log.Fatalf("Error updating user group users: %v\n", err)
		}

		slackBody, _ := jsoniter.Marshal(map[string]string{"text": message})
		response, err := http.Post(LambdaConfig.SlackConfig.WebhookUrl, "application/json", bytes.NewBuffer(slackBody))
		if err != nil {
			log.Fatalf("Error sending message to Slack: %v\n", err)
		}

		if response.StatusCode != 200 {
			log.Fatalf("Error sending message to Slack: received non-200 response: %d\n", response.StatusCode)
		}
	}

	return message, nil
}
