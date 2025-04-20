package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type progressWriter struct {
	Reader io.Reader
	Total  int64
	Bytes  int64
}

var (
	urlPtr = flag.String("url", "", "URL to download")
)

func downloadFile(url string) error {
	fileName := path.Base(url)

	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	response, err := http.Get(url)

	if err != nil {
		log.Fatal()
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Download Failed: %s (status: %d)", url, response.StatusCode)
	}
	contentLength, _ := strconv.ParseInt(response.Header.Get("Content-Length"), 10, 64)

	progressReader := &progressWriter{Reader: response.Body, Total: contentLength}

	_, err = io.Copy(outFile, progressReader)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	fmt.Printf("\n Download completed: %s\n", fileName)
	return nil
}

func (pw *progressWriter) Read(p []byte) (int, error) {
	n, err := pw.Reader.Read(p)
	pw.Bytes += int64(n)
	pw.printProgress()
	return n, err
}

func (pw *progressWriter) printProgress() {
	if pw.Total > 0 {
		percentage := float64(pw.Bytes) / float64(pw.Total) * 100
		fmt.Printf("\rDownloading: %.2f%% [%d/%d bytes]", percentage, pw.Bytes, pw.Total)
	}
}

func main() {

	flag.Parse()

	fmt.Print("Welcome to uwuGet!\n")

	if strings.HasPrefix(*urlPtr, "") {
		fmt.Print("Make sure to follow the intstructions\n 1. Either Go run uwuGet.go -url 'some http url' ")
		os.Exit(1)

	} else if !strings.HasPrefix(*urlPtr, "http") {
		fmt.Println("Invalid URL. Must start with http or https.")
		os.Exit(1)
	}

	start := time.Now()
	if err := downloadFile(*urlPtr); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("\nTime elapsed: %.2fs\n", time.Since(start).Seconds())
}
