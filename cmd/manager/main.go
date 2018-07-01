package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/vigasin/toredo/sqschan"
	"github.com/vigasin/toredo"
	"github.com/satori/go.uuid"
	"encoding/json"
)

const (
	DownloaderInQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-downloader-in"
	TransfererInQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-transferer-in"
	Region               = "us-west-1"
)

type ApiRequest struct {
	Url string
}

type ApiResponse struct {
	RequestId string
}

func HandleApiEvent(event ApiRequest) (ApiResponse, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	client := sqschan.NewClient(svc, DownloaderInQueueUrl)

	requestId, _ := uuid.NewV4()

	message := &toredo.DownloaderInMessage{Url: event.Url, RequestId: requestId.String(), MessageType: toredo.MsgDownloaderInDownload}

	messageJson, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Couldn't marshal message. Error: %s\n", err)
	}

	fmt.Print(string(messageJson))
	err = client.SendMessage(string(messageJson))
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}

	response := ApiResponse{RequestId:requestId.String()}
	return response, nil
}

func main() {
	lambda.Start(HandleApiEvent)
}
