package queue

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	testRegion      = "ap-northeast-1"
	testEndpointURL = "http://127.0.0.1:4566"
)

func TestDequeue(t *testing.T) {
	cli, err := NewSQSClient(testRegion, testEndpointURL)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc     string
		prepare  func() error
		queueURL string
		want     string
		err      error
		clean    func() error
	}{
		{
			desc: "dequeue from existent queue",
			prepare: func() error {
				qURL, err := createQueueForTest(t, cli, "test-queue1.fifo")
				if err != nil {
					return err
				}
				return enqueueForTest(t, cli, qURL, "test-queue1.fifo", `{"foo":"bar"}`)
			},
			queueURL: "http://127.0.0.1:4566/000000000000/test-queue1.fifo",
			want:     `{"foo":"bar"}`,
			err:      nil,
			clean:    func() error { return nil },
		},
		{
			desc:     "dequeue from non-existent queue",
			prepare:  func() error { return nil },
			queueURL: "http://127.0.0.1:4566/000000000000/test-queue2.fifo",
			want:     "",
			err:      fmt.Errorf("Failed to receive message from AWS SQS"),
			clean:    func() error { return nil },
		},
		{
			desc: "dequeue from empty queue",
			prepare: func() error {
				_, err := createQueueForTest(t, cli, "test-queue3.fifo")
				return err
			},
			queueURL: "http://127.0.0.1:4566/000000000000/test-queue3.fifo",
			want:     "",
			err:      nil,
			clean:    func() error { return nil },
		},
	}

	for n, c := range cases {
		if err := c.prepare(); err != nil {
			t.Error(err)
			continue
		}

		got, err := cli.Dequeue(c.queueURL)
		if c.err != nil || err != nil {
			if (c.err != nil && err == nil) || (c.err == nil && err != nil) || !strings.Contains(err.Error(), c.err.Error()) {
				t.Error(fmt.Errorf("%d: %w", n, err))
			}
		}

		if got != c.want {
			t.Errorf("%d: want=%s, got=%s", n, c.want, got)
		}

		if err := c.clean(); err != nil {
			t.Error(err)
		}
	}
}

func createQueueForTest(t *testing.T, s *SQSClient, key string) (string, error) {
	t.Helper()

	input := sqs.CreateQueueInput{
		QueueName: aws.String(key),
		Attributes: map[string]string{
			"FifoQueue":                 "true",
			"ContentBasedDeduplication": "true",
		},
	}

	output, err := s.cli.CreateQueue(context.TODO(), &input)
	if err != nil {
		return "", fmt.Errorf("Failed to create AWS SQS queue URL: %w", err)
	}

	return *output.QueueUrl, nil
}

func enqueueForTest(t *testing.T, s *SQSClient, queueURL, key, msg string) error {
	t.Helper()

	input := sqs.SendMessageInput{
		MessageBody:            aws.String(msg),
		MessageGroupId:         aws.String(key),
		MessageDeduplicationId: aws.String(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s%d", msg, time.Now().UnixMicro()))))),
		QueueUrl:               aws.String(queueURL),
	}
	if _, err := s.cli.SendMessage(context.TODO(), &input); err != nil {
		return fmt.Errorf("Failed to send message to AWS SQS: %w", err)
	}

	return nil
}
