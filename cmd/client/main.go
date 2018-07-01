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
	"github.com/jessevdk/go-flags"
	"os"
)

const (
	QueueUrl    = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-downloader-in"
	Region      = "us-west-1"
	CredPath    = "/Users/ivigasin/.aws/credentials"
	CredProfile = "default"
)

func downloadByUrl(svc *sqs.SQS, url string) {
	var message *toredo.DownloaderInMessage

	requestId, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Couldn't generate request id. Error: %s\n", err)
		return
	}

	message = &toredo.DownloaderInMessage{Url: url, RequestId: requestId.String(), MessageType: toredo.MsgDownloaderInDownload}

	messageJson, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Couldn't marshal message. Error: %s\n", err)
		return
	}

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

func main() {
	var opts struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

		Positional struct {
			Url []string `required:"yes" positional-arg-name:"URL"`
		} `positional-args:"yes" description:"Magnet URLs"`
	}

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()

	if err != nil {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
		MaxRetries:  aws.Int(5),
	}))

	svc := sqs.New(sess)

	for _, url := range opts.Positional.Url {
		downloadByUrl(svc, url)
	}
}
