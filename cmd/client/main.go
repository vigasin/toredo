package main

import (
	"encoding/json"
	"fmt"
	"github.com/vigasin/toredo"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/satori/go.uuid"
	"log"
	"os"
)

const (
	QueueUrl    = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-download-queue"
	Region      = "us-west-1"
	CredPath    = "/Users/ivigasin/.aws/credentials"
	CredProfile = "default"
)

func main() {
	var message *toredo.DownloaderMessage

	url := "magnet:?xt=urn:btih:144D0B866D954E4663B7797D0023493C44EF0F4D&tr=http%3A%2F%2Fbt2.t-ru.org%2Fann%3Fmagnet&dn=Robert%20McKee%20%2F%20Роберт%20Макки%20-%20Story%20%2F%20История%20на%20миллион%20долларов%20%5B2008%2C%20EPUB%2FFB2%2FMOBI%2C%20RUS%5D"

	requestId, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Couldn't generate request id. Error: %s\n", err)
		return
	}

	if len(os.Args) > 1 {
		message = &toredo.DownloaderMessage{Url: url, RequestId: requestId.String(), MessageType: toredo.MsgInfo}
	} else {
		message = &toredo.DownloaderMessage{Url: url, RequestId: requestId.String(), MessageType: toredo.MsgDownload}
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Couldn't marshal message. Error: %s\n", err)
		return
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
		MaxRetries:  aws.Int(5),
	}))

	svc := sqs.New(sess)

	// Send message
	sendParams := &sqs.SendMessageInput{
		QueueUrl:    aws.String(QueueUrl),
		MessageBody: aws.String(string(messageJson)),
	}

	sendResp, err := svc.SendMessage(sendParams)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("[Send message] \n%v \n\n", sendResp)
}
