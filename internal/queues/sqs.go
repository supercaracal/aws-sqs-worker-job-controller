package queues

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	dequeueSize = 1
	waitTimeout = 5
)

// MySQS is
type MySQS struct {
	cli *sqs.SQS
}

// NewMySQS is
func NewMySQS(region, endpointURL string) (*MySQS, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create AWS client's session: %w", err)
	}

	config := aws.NewConfig().WithRegion(region)
	if endpointURL != "" {
		config = config.WithEndpoint(endpointURL)
	}

	return &MySQS{cli: sqs.New(sess, config)}, nil
}

// Dequeue is
func (s *MySQS) Dequeue(queueURL string) (string, error) {
	input := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: aws.Int64(dequeueSize),
		WaitTimeSeconds:     aws.Int64(waitTimeout),
	}

	output, err = s.cli.ReceiveMessage(&input)
	if err != nil {
		return fmt.Errorf("Failed to receive message from AWS SQS: %w", err)
	}
	if len(output.Messages) != dequeueSize {
		return fmt.Errorf("The queue is empty or something is wrong")
	}

	return output.Messages[0]
}
