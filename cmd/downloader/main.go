package main

import (
	"fmt"
	"log"

	"github.com/vigasin/toredo"
	"github.com/vigasin/toredo/torrent_manager"
	"github.com/vigasin/toredo/sqschan"

	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/satori/go.uuid"
	"os"
	"github.com/jessevdk/go-flags"
	"runtime"
	"bytes"
	"io"
)

const (
	InQueueUrl  = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-downloader-in"
	OutQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-downloader-out"
	Region      = "us-west-1"
)

type Config struct {
	DownloadPath string
	PublicPath   string
	PublicUrl    string
}

func deleteMessage(svc *sqs.SQS, queue string, message *sqs.Message) {
	// Delete message
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

func processMessage(config Config, manager *torrent_manager.TorrentManager, svc *sqs.SQS, msg toredo.DownloaderInMessage) toredo.DownloaderOutMessage {
	var outMessage toredo.DownloaderOutMessage

	switch msg.MessageType {
	case toredo.MsgDownloaderInDownload:
		{
			tarFile := manager.DownloadTorrent(msg.RequestId, msg.Url)
			tarUrl := fmt.Sprintf("%s/%s", config.PublicUrl, tarFile)
			fmt.Println(tarFile)

			outMessage = toredo.DownloaderOutMessage{
				MessageType: toredo.MsgDownloaderOutDownloaded,
				Url:         tarUrl,
			}
		}

	case toredo.MsgDownloaderInInfo:
		{
			buf := bytes.NewBufferString("")
			writer := io.Writer(buf)
			manager.WriteStatus(writer)

			requestId, _ := uuid.NewV4()

			outMessage = toredo.DownloaderOutMessage{
				RequestId:   requestId.String(),
				MessageType: toredo.MsgDownloaderOutGotInfo,
				Message: buf.String(),
			}
		}
	}

	return outMessage

}

func main() {
	// Use up to 20 OS threads
	runtime.GOMAXPROCS(20)

	var opts struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

		ConfigFile string `short:"c" long:"config" description:"Path to config file" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()

	if err != nil {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(opts.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	manager := torrent_manager.New(config.DownloadPath, config.PublicPath)

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewEnvCredentials(),
		MaxRetries:  aws.Int(5),
	}))

	svc := sqs.New(sess)

	client := sqschan.NewClient(svc, InQueueUrl)
	ch, errch, err := sqschan.Incoming(client)
	if err != nil {
		log.Printf("Can't create incoming channel. Error: %v", err)
	}

	var message *sqs.Message

	for {
		select {
		case err = <-errch:
			log.Printf("Error reading from SQS channel: %v", err)
		case message = <-ch:
			msg := toredo.DownloaderInMessage{}
			err := json.Unmarshal(([]byte)(*message.Body), &msg)

			if err != nil {
				log.Printf("Can't unmarshal message. Error: %s\n", err)
				continue
			}

			fmt.Printf("[Receive message] \n%v \n\n", msg)

			client.DeleteMessage(message)
			go func() {
				outMessage := processMessage(config, manager, svc, msg)

				requestId, err := uuid.NewV4()
				outMessage.RequestId = requestId.String()

				messageJson, err := json.Marshal(outMessage)
				if err != nil {
					fmt.Println(err)
					return
				}

				sendMessage(svc, OutQueueUrl, string(messageJson))
			}()
		}
	}

}
