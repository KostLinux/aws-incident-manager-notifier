package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
)

const weekDuration = time.Hour * 24 * 7

var (
	client           *ssmcontacts.Client
	slackWebhookUrl  string
	slackApiUrl      string
	slackApiToken    string
	slackSubTeamID   string
	slackSubTeamName string
	slackUserIDs     map[string]string
)

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	awsRegion := getEnvOrDefault("AWS_REGION", "eu-west-1")
	slackApiUrl = getEnvOrDefault("SLACK_API_URL", "https://slack.com/api/usergroups.users.update")
	slackWebhookUrl = getEnvOrDefault("SLACK_WEBHOOK_URL", "")
	slackApiToken = getEnvOrDefault("SLACK_API_TOKEN", "")
	slackSubTeamID = getEnvOrDefault("SLACK_SUBTEAM_ID", "")
	slackSubTeamName = getEnvOrDefault("SLACK_SUBTEAM_NAME", "support")

	if slackApiToken == "" {
		log.Fatalf("error: Slack API token is not set")
	}

	// Load all environment variables starting with LHV_ into a map
	slackUserIDs = make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "SSM_") {
			slackUserIDs[strings.ToLower(pair[0])] = pair[1]
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("error loading AWS configuration: %v\n", err)
	}

	client = ssmcontacts.NewFromConfig(cfg, func(options *ssmcontacts.Options) {
		options.Region = awsRegion
	})
}

func loadAndPrintAllRotationShifts(ctx context.Context) ([]string, error) {
	outputRotations, err := client.ListRotations(ctx, &ssmcontacts.ListRotationsInput{})
	if err != nil {
		return nil, fmt.Errorf("error loading rotations: %w", err)
	}

	var contactIds []string
	for _, rotation := range outputRotations.Rotations {
		outputShifts, err := client.ListRotationShifts(ctx, &ssmcontacts.ListRotationShiftsInput{
			EndTime:    aws.Time(time.Now().Add(1 * weekDuration)),
			RotationId: rotation.RotationArn,
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)),
		})
		if err != nil {
			return nil, fmt.Errorf("error loading rotation shifts for rotation %s: %w", *rotation.RotationArn, err)
		}

		for _, shift := range outputShifts.RotationShifts {
			contactIds = append(contactIds, shift.ContactIds...)
		}
	}

	return contactIds, nil
}

func printSupportEngineerForWeek() string {
	contactIds, err := loadAndPrintAllRotationShifts(context.Background())
	if err != nil {
		log.Fatalf("Error loading rotation shifts: %v\n", err)
		return ""
	}

	for _, contactId := range contactIds {
		contactName := strings.Split(contactId, "/")[1]
		if userID, ok := slackUserIDs[contactName]; ok {
			return fmt.Sprintf("<!subteam^%s|%s> this week is <@%s> \n", slackSubTeamID, slackSubTeamName, userID)
		}
	}
	return ""
}

func updateUserGroupUsers(userGroupID string, userIDs []string) error {

	data := url.Values{}
	data.Set("token", slackApiToken)
	data.Set("usergroup", userGroupID)
	data.Set("users", strings.Join(userIDs, ","))

	_, err := http.PostForm(slackApiUrl, data)
	if err != nil {
		return err
	}

	return nil
}

func HandleRequest(ctx context.Context) (string, error) {
	message := printSupportEngineerForWeek()
	fmt.Println(message)
	if message != "" {
		// Extract user IDs from the userIDs map
		var ids []string
		for _, id := range slackUserIDs {
			ids = append(ids, id)
		}

		// Update user group users
		err := updateUserGroupUsers(slackSubTeamID, ids)
		if err != nil {
			log.Fatalf("Error updating user group users: %v\n", err)
		}

		slackBody, _ := json.Marshal(map[string]string{"text": message})
		resp, err := http.Post(slackWebhookUrl, "application/json", bytes.NewBuffer(slackBody))
		if err != nil {
			log.Fatalf("Error sending message to Slack: %v\n", err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("Error sending message to Slack: received non-200 response: %d\n", resp.StatusCode)
		}
	}
	return message, nil
}

func main() {
	lambda.Start(HandleRequest)
}
