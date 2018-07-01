package sqschan

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
)

func Incoming(client *Client) (<-chan *sqs.Message, <-chan error, error) {
	ch := make(chan *sqs.Message)
	errch := make(chan error)

	go func() {
		for {
			// Receive message
			receiveParams := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(client.queue),
				MaxNumberOfMessages: aws.Int64(3),
				VisibilityTimeout:   aws.Int64(30),
				WaitTimeSeconds:     aws.Int64(20),
			}
			receiveResp, err := client.svc.ReceiveMessage(receiveParams)
			if err != nil {
				errch <- err
				continue
			}

			for _, message := range receiveResp.Messages {
				ch <- message
			}
		}
	}()

	return ch, errch, nil
}

func Outgoing(client *Client) (chan <- string, <-chan error, error) {
	ch := make(chan string)
	errch := make(chan error)

	go func() {
		for msg := range ch {
			client.SendMessage(msg)
		}
	}()

	return ch, errch, nil
}