package sqschan

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

type Client struct {
	svc   *sqs.SQS
	queue string
}

func NewClient(svc *sqs.SQS, queue string) (*Client) {
	return &Client{svc: svc, queue: queue}
}

func (client *Client) DeleteMessage(message *sqs.Message) error {
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(client.queue),
		ReceiptHandle: message.ReceiptHandle,
	}

	_, err := client.svc.DeleteMessage(deleteParams) // No response returned when succeeded.
	if err != nil {
		return err
	}

	fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)

	return nil
}

func (client *Client) SendMessage(message string) error {
	sendParams := &sqs.SendMessageInput{
		QueueUrl:    aws.String(client.queue),
		MessageBody: aws.String(message),
	}

	sendResp, err := client.svc.SendMessage(sendParams)
	if err != nil {
		return err
	}

	fmt.Printf("[Send message] \n%v \n\n", sendResp)

	return nil
}
