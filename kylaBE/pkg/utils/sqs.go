package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var sqsClient *sqs.Client

func NewSQSClient(region, accessKey, secretKey string) *sqs.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
	)

	if err != nil {
		panic(fmt.Errorf("error loading aws configs %w", err))
	}

	sqsClient = sqs.NewFromConfig(cfg)
	return sqsClient
}

type SQSActions struct {
	client *sqs.Client
}

func NewSQSActions(client *sqs.Client) *SQSActions {
	return &SQSActions{
		client: client,
	}
}

type MessageAttribute struct {
	Name  string
	Type  string
	Value string
}

func formatMessageAttributes(attributes []MessageAttribute) map[string]types.MessageAttributeValue {
	messageAttributes := map[string]types.MessageAttributeValue{}
	for _, attribute := range attributes {
		messageAttributes[attribute.Name] = types.MessageAttributeValue{
			DataType:    &attribute.Type,
			StringValue: &attribute.Value,
		}
	}

	return messageAttributes
}

func (s SQSActions) SendMessage(ctx context.Context, queueUrl string, messageBody string, attributes []MessageAttribute) (string, error) {
	out, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          &queueUrl,
		MessageBody:       &messageBody,
		MessageAttributes: formatMessageAttributes(attributes),
	})

	if err != nil {
		log.Printf("error sending message to queue: %v", err)
	}

	return *out.MessageId, err
}

func (s SQSActions) GetQueueUrl(ctx context.Context, queueName string) (string, error) {
	var queueUrl string
	result, err := s.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		log.Printf("error getting queue url: %v", err)
	} else {
		queueUrl = *result.QueueUrl
	}

	return queueUrl, err
}

func (s SQSActions) GetMessage(ctx context.Context, queueUrl string, maxMessages int32, waitTime int32) ([]types.Message, error) {
	var messages []types.Message
	result, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              &queueUrl,
		MaxNumberOfMessages:   maxMessages,
		WaitTimeSeconds:       waitTime,
		MessageAttributeNames: []string{"All"},
	})

	if err != nil {
		log.Printf("error receiving message from queue: %v", err)
	} else {
		messages = result.Messages
	}

	return messages, err
}

func (s SQSActions) DeleteMessage(ctx context.Context, queueUrl string, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: &receiptHandle,
	})

	if err != nil {
		log.Printf("error deleting message from queue: %v", err)
	}

	return err
}
