package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
)

type AwsConfig struct {
	AwsRegion         string
	SSMContactsClient *ssmcontacts.Client
}

func NewConfig(awsRegion string) (*AwsConfig, error) {
	// Load the default AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	// Create the SSM Contacts client with additional options
	ssmClient := ssmcontacts.NewFromConfig(cfg, func(options *ssmcontacts.Options) {
		options.Region = awsRegion
	})

	if ssmClient == nil {
		log.Fatalf("failed to create SSM Contacts client")
	}

	return &AwsConfig{
		AwsRegion:         awsRegion,
		SSMContactsClient: ssmClient,
	}, nil
}

func (cfg *AwsConfig) LoadAndPrintAllRotationShifts(ctx context.Context) ([]string, error) {
	weekDuration := time.Hour * 24 * 7
	var contactIds []string

	rotationList, err := cfg.SSMContactsClient.ListRotations(ctx, &ssmcontacts.ListRotationsInput{})
	if err != nil {
		return nil, fmt.Errorf("error loading rotations: %w", err)
	}

	for _, rotation := range rotationList.Rotations {
		shifts, err := cfg.SSMContactsClient.ListRotationShifts(ctx, &ssmcontacts.ListRotationShiftsInput{
			EndTime:    aws.Time(time.Now().Add(1 * weekDuration)),
			RotationId: rotation.RotationArn,
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)),
		})

		if err != nil {
			return nil, fmt.Errorf("error loading rotation shifts for rotation %s: %w", *rotation.RotationArn, err)
		}

		for _, shift := range shifts.RotationShifts {
			contactIds = append(contactIds, shift.ContactIds...)
		}
	}

	return contactIds, nil
}