package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"encoding/json"
	"github.com/anacrolix/torrent"
	"os"
	"archive/tar"
	"io"
	"path/filepath"
	"github.com/satori/go.uuid"
)

const (
	DownloadQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-download-queue"
	TransferQueueUrl = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-transfer-queue"
	Region           = "us-west-1"
	CredPath         = "/Users/ivigasin/.aws/credentials"
	CredProfile      = "default"
)

type Message struct {
	Url       string
	RequestId string
}

func addFile(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = path
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}

func tarFolder(archiveName string, src string) {
	file, err := os.Create(archiveName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = file // strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		defer f.Close()
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}

func downloadTorrent(requestId string, url string) string {
	client, err := torrent.NewClient(nil)
	if err != nil {
		log.Println(err)
	}

	defer client.Close()

	t, _ := client.AddMagnet(url)

	<-t.GotInfo()
	t.DownloadAll()
	client.WaitAll()

	tarName := fmt.Sprintf("%s.tar", requestId)
	tarFolder(tarName, t.Info().Name)

	os.RemoveAll(t.Info().Name)

	log.Printf("Finished %s", t.Info().Name)

	return tarName
}

type TransferMessage struct {
	RequestId string
	Url string
}

func main() {

	for ; ; {
		sess := session.New(&aws.Config{
			Region:      aws.String(Region),
			Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
			MaxRetries:  aws.Int(5),
		})

		svc := sqs.New(sess)

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

		// Delete message
		for _, message := range receiveResp.Messages {
			msg := Message{}
			err := json.Unmarshal(([]byte)(*message.Body), &msg)

			tarFile := downloadTorrent(msg.RequestId, msg.Url)
			fmt.Println(tarFile)

			deleteParams := &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(DownloadQueueUrl), // Required
				ReceiptHandle: message.ReceiptHandle,        // Required
			}
			_, err = svc.DeleteMessage(deleteParams) // No response returned when succeeded.
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)

			baseUrl := "http://enk.me/vivo"
			requestId, err := uuid.NewV4()
			transferMessage := TransferMessage{
				RequestId: requestId.String(),
				Url: fmt.Sprintf("%s/%s", baseUrl, tarFile),
			}

			messageJson, err := json.Marshal(transferMessage)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Send message
			sendParams := &sqs.SendMessageInput{
				QueueUrl:    aws.String(TransferQueueUrl),
				MessageBody: aws.String(string(messageJson)),
			}

			sendResp, err := svc.SendMessage(sendParams)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("[Send message] \n%v \n\n", sendResp)
		}
	}

}
