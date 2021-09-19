package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	dequeueSize    = 1
	waitTimeout    = 0 // Don't block the loop
	requestTimeout = 10 * time.Second
)

// SQSClient is
type SQSClient struct {
	cli *sqs.Client
}

// NewSQSClient is
func NewSQSClient(region, endpointURL string) (*SQSClient, error) {
	cfg, err := loadAWSConfig(region, endpointURL)
	if err != nil {
		return nil, err
	}

	return &SQSClient{cli: sqs.NewFromConfig(cfg)}, nil
}

func loadAWSConfig(region, endpointURL string) (aws.Config, error) {
	if endpointURL == "" {
		return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	}

	// for localstack
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           endpointURL,
						SigningRegion: region,
					}, nil
				},
			),
		),
	)
}

// Dequeue is
func (s *SQSClient) Dequeue(queueURL string) (string, error) {
	input := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: dequeueSize,
		WaitTimeSeconds:     waitTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	output, err := s.cli.ReceiveMessage(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("Failed to receive message from AWS SQS: %w", err)
	}

	size := len(output.Messages)
	if size == 0 {
		return "", nil
	}
	if size > dequeueSize {
		return "", fmt.Errorf("Failed to receive a message from AWS SQS")
	}

	err = s.deleteMessage(ctx, aws.String(queueURL), output.Messages[0].ReceiptHandle)
	if err != nil {
		return "", err
	}

	return *output.Messages[0].Body, nil
}

func (s *SQSClient) deleteMessage(ctx context.Context, queueURL, identifier *string) error {
	input := sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: identifier,
	}

	_, err := s.cli.DeleteMessage(ctx, &input)
	if err != nil {
		return fmt.Errorf("Failed to delete message from AWS SQS: %w", err)
	}

	return nil
}
