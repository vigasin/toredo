package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/anacrolix/torrent"
)

const (
	QueueUrl    = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-download-queue"
	Region      = "us-west-1"
	CredPath    = "/Users/ivigasin/.aws/credentials"
	CredProfile = "default"
)

func main() {

	for ; ;  {
		sess := session.New(&aws.Config{
			Region:      aws.String(Region),
			Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
			MaxRetries:  aws.Int(5),
		})

		svc := sqs.New(sess)

		// Receive message
		receiveParams := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(QueueUrl),
			MaxNumberOfMessages: aws.Int64(3),
			VisibilityTimeout:   aws.Int64(30),
			WaitTimeSeconds:     aws.Int64(20),
		}
		receiveResp, err := svc.ReceiveMessage(receiveParams)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("[Receive message] \n%v \n\n", receiveResp)

		// Delete message
		for _, message := range receiveResp.Messages {
			clientConfig := &torrent.ClientConfig{
				DataDir: "penis",
			}
			client, err := torrent.NewClient(clientConfig)
			if err != nil {
				log.Println(err)
			}

			client.AddMagnet(message.String())

			deleteParams := &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(QueueUrl),  // Required
				ReceiptHandle: message.ReceiptHandle, // Required

			}
			_, err = svc.DeleteMessage(deleteParams) // No response returned when succeeded.
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)
		}
	}

}