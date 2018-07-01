package main

import (
	"runtime"
	"fmt"
	"time"
	"net/http"
	"strings"
	"strconv"
	"os"
	"io"
)

type download struct {
	url         string
	dlStatus    chan string
	fileName    string
	fileStatus  string
	connections int64
	size        int64
	client      *http.Client
}

func (dl *download) get(chunkSize int64, i int64) {
	var start, end int64
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
	reqRange := "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end, 10)
	//fmt.Println("Request :", reqRange)
	req.Header.Set("Range", reqRange)

	filename := dl.fileName + "." + strconv.FormatInt(i, 10)
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
	i := dl.connections - int64(1)
	out, _ := os.Create(dl.fileName)
	defer out.Close()
	for j := int64(0); j <= i; j++ {
		infile := dl.fileName + "." + strconv.FormatInt(j, 10)
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
	dl.size = int64(resp.ContentLength)
	isPartial := resp.Header.Get("Accept-Ranges")
	if isPartial == "" {
		fmt.Println("Partial downloads not supported", resp.Body)
		dl.connections = 1
	} else {
		fmt.Println("Partials supported: ", isPartial)
	}
	chunkSize := int64(dl.size) / (dl.connections)
	fmt.Println("Content size:", dl.size, "\n Chunk Size ", chunkSize)

	for i := int64(0); i < dl.connections; i++ {
		go dl.get(chunkSize, i)
	}

	defer resp.Body.Close()

	for j := int64(0); j < dl.connections; j++ {

		<-dl.dlStatus

	}
	close(dl.dlStatus)
	timeTaken := time.Since(startTime).Seconds()
	dl.join()

	speed := (float64(dl.size / 1024)) / timeTaken
	fmt.Printf("done time taken:  %f , %.4f Speed:kB/s\n", timeTaken, speed)

	return dl.fileName
}
