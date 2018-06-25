package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"encoding/json"
	"os"
	"strings"
	"net/http"
	"time"
	"strconv"
	"io"
	"runtime"
)

const (
	QueueUrl    = "https://sqs.us-west-1.amazonaws.com/009544449203/toredo-transfer-queue"
	Region      = "us-west-1"
	CredPath    = "/Users/ivigasin/.aws/credentials"
	CredProfile = "default"
)

type download struct {
	url         string
	dlStatus    chan string
	fileName    string
	fileStatus  string
	connections int
	size        int
	client      *http.Client
}

func (dl *download) get(chunkSize int, i int) {
	var start, end int
	req, _ := http.NewRequest("GET", dl.url, nil)
	if i == 0 {
		start = 0
	} else {
		start = (i * chunkSize)
	}

	if i == (dl.connections - 1) {
		end = dl.size
	} else {

		end = start + chunkSize - 1
	}
	reqRange := "bytes=" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
	//fmt.Println("Request :", reqRange)
	req.Header.Set("Range", reqRange)

	filename := dl.fileName + "." + strconv.Itoa(i)
	out, _ := os.Create(filename)

	client := dl.client
	resp, _ := client.Do(req)

	writer := io.Writer(out)

	_, err := io.Copy(writer, resp.Body)

	if err != nil {
		fmt.Println("Error occured", err)
		dl.dlStatus <- "failed"
		return
	} else {
		//fmt.Println("Success", filename, " written", written)
		dl.dlStatus <- filename + " done"
		return

	}

}

func (dl *download) join() string {
	i := dl.connections - 1
	out, _ := os.Create(dl.fileName)
	defer out.Close()
	for j := 0; j <= i; j++ {

		infile := dl.fileName + "." + strconv.Itoa(j)
		in, _ := os.Open(infile)
		defer os.Remove(infile)
		writer := io.Writer(out)

		_, err := io.Copy(writer, in)

		if err != nil {
			fmt.Println("Error occured", err)

		} else {
			//fmt.Println("file ", infile, "joined")

		}

	}

	fmt.Println("file join complete")
	return "done"
}

func downloadFile(url string) string {
	cpuCount := runtime.NumCPU()
	i := runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Was using ", i, " CPUs changing to ", cpuCount)
	startTime := time.Now()

	dl := download{
		url:         url,
		connections: 30,
		dlStatus:    make(chan string, 1),
		client:      &http.Client{},
	}

	i = strings.LastIndex(dl.url, "/")
	dl.fileName = dl.url[i+1:]

	fmt.Println(dl.url)

	resp, err := dl.client.Get(dl.url)

	if err != nil {
		fmt.Println(err)
	}
	dl.size = int(resp.ContentLength)
	isPartial := resp.Header.Get("Accept-Ranges")
	if isPartial == "" {
		fmt.Println("Partial downloads not supported", resp.Body)
		dl.connections = 1
	} else {
		fmt.Println("Partials supported: ", isPartial)
	}
	chunkSize := int(dl.size) / (dl.connections)
	fmt.Println("Content size:", dl.size, "\n Chunk Size ", chunkSize)

	for i := 0; i < dl.connections; i++ {
		go dl.get(chunkSize, i)
	}

	defer resp.Body.Close()

	for j := 0; j < dl.connections; j++ {

		<-dl.dlStatus

	}
	close(dl.dlStatus)
	timeTaken := time.Since(startTime).Seconds()
	dl.join()

	speed := (float64(dl.size / 1024)) / timeTaken
	fmt.Printf("done time taken:  %f , %.4f Speed:kB/s\n", timeTaken, speed)

	return dl.fileName
}

type TransferMessage struct {
	RequestId string
	Url       string
}

func main() {
	if true {
		downloadFile("https://la.hdw.mx/~ivigasin/from_logs.tar.gz")
	} else {
		for ; ; {
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
				msg := TransferMessage{}
				err := json.Unmarshal(([]byte)(*message.Body), &msg)

				file := downloadFile(msg.Url)
				fmt.Println(file)

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
}
