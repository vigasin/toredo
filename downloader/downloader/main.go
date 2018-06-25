package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/satori/go.uuid"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"./torrent_manager"
)

const (
	DownloadQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-download-queue"
	TransferQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-transfer-queue"
	Region           = "us-west-1"
	CredPath         = "/Users/ivigasin/.aws/credentials"
	CredProfile      = "default"
)

const (
	MsgDownload = "Download"
	MsgRemove   = "Remove"
	MsgInfo     = "Info"
)

type Config struct {
	DownloadPath string
	PublicUrl    string
}

type Message struct {
	RequestId   string
	MessageType string // Download, Remove, Info
	Url         string
}

type TransferMessage struct {
	RequestId string
	Url       string
}

func deleteMessage(svc *sqs.SQS, queue string, message *sqs.Message) {
	// Delete message
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queue),     // Required
		ReceiptHandle: message.ReceiptHandle, // Required
	}
	_, err := svc.DeleteMessage(deleteParams) // No response returned when succeeded.
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)
}

func sendMessage(svc *sqs.SQS, queue string, body string) {
	// Send message
	sendParams := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(body),
	}

	sendResp, err := svc.SendMessage(sendParams)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("[Send message] \n%v \n\n", sendResp)
}

func processMessage(config Config, manager *torrent_manager.TorrentManager, svc *sqs.SQS, message *sqs.Message) {
	msg := Message{}
	err := json.Unmarshal(([]byte)(*message.Body), &msg)

	if err != nil {
		log.Printf("Can't unmarshal message. Error: %s\n", err)
		return
	}

	switch msg.MessageType {
	case MsgDownload:
		{
			tarFile := manager.DownloadTorrent(msg.RequestId, msg.Url)
			fmt.Println(tarFile)
		}

	case MsgInfo:
		{

		}
	}

	deleteMessage(svc, DownloadQueueUrl, message)

	requestId, err := uuid.NewV4()
	transferMessage := TransferMessage{
		RequestId: requestId.String(),
		Url:       fmt.Sprintf("%s/%s", config.PublicUrl, tarFile),
	}

	messageJson, err := json.Marshal(transferMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	sendMessage(svc, TransferQueueUrl, string(messageJson))
}

func main() {
	content, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", config)

	if true {
		return
	}

	manager := torrent_manager.New()

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
		MaxRetries:  aws.Int(5),
	}))

	svc := sqs.New(sess)

	for ; ; {
		// Receive message
		receiveParams := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(DownloadQueueUrl),
			MaxNumberOfMessages: aws.Int64(3),
			VisibilityTimeout:   aws.Int64(30),
			WaitTimeSeconds:     aws.Int64(20),
		}
		receiveResp, err := svc.ReceiveMessage(receiveParams)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("[Receive message] \n%v \n\n", receiveResp)

		for _, message := range receiveResp.Messages {
			go processMessage(config, manager, svc, message)
		}
	}

}
