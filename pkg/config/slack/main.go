package slack

import (
	"os"
	"strings"
)

type SlackConfig struct {
	ApiUrl      string
	WebhookUrl  string
	ApiToken    string
	SubTeamID   string
	SubTeamName string
	UserIDs     map[string]string
}

func NewConfig(
	slackApiUrl,
	slackWebhookUrl,
	slackApiToken,
	slackSubTeamID,
	slackSubTeamName string,
	slackUserIDs map[string]string,
) *SlackConfig {

	return &SlackConfig{
		ApiUrl:      slackApiUrl,
		WebhookUrl:  slackWebhookUrl,
		ApiToken:    slackApiToken,
		SubTeamID:   slackSubTeamID,
		SubTeamName: slackSubTeamName,
		UserIDs:     slackUserIDs,
	}
}

func (config *SlackConfig) ConfigureUserIDs() {
	// Iterate over all environment variables
	for _, envVar := range os.Environ() {
		// Split the environment variable into key and value
		keyValuePair := strings.SplitN(envVar, "=", 2)
		key := keyValuePair[0]
		value := keyValuePair[1]

		// Check if the key has the prefix "SSM_"
		if strings.HasPrefix(key, "SSM_") {
			// Convert the key to lowercase and store it in the UserIDs map
			config.UserIDs[strings.ToLower(key)] = value
		}
	}
}
