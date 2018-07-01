package main

import (
	"fmt"
	"log"

	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
	"github.com/vigasin/toredo/tar"
	"github.com/vigasin/toredo"
	"os/user"
)

const (
	QueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-downloader-out"
	Region   = "us-west-1"
)

type TransferMessage struct {
	RequestId string
	Url       string
}

func main() {
	usr, _ := user.Current()
	homedir := usr.HomeDir

	credPath := fmt.Sprintf("%s/.aws/credentials", homedir)
	credProfile := "transferer"

	for {
		sess := session.New(&aws.Config{
			Region:      aws.String(Region),
			Credentials: credentials.NewSharedCredentials(credPath, credProfile),
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

		// Delete message
		for _, message := range receiveResp.Messages {
			msg := toredo.DownloaderOutMessage{}
			err := json.Unmarshal(([]byte)(*message.Body), &msg)

			fmt.Printf("[Receive message] \n%v \n\n", msg)

			deleteParams := &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(QueueUrl),  // Required
				ReceiptHandle: message.ReceiptHandle, // Required
			}

			_, err = svc.DeleteMessage(deleteParams) // No response returned when succeeded.
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)

			file := downloadFile(msg.Url)
			fmt.Println(file)

			tar.UntarFolder(file, ".")

			os.Remove(file)
		}
	}
}
