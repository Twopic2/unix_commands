package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	O   = flag.String("O", "output", "O is the output file")
	GET = flag.String("GET", "", "GET protocol")
)

func main() {

	fmt.Println("Welcome to UWURL!")
	flag.Parse()

	if *GET == "" {
		fmt.Print("Make sure to follow the intstructions")
		os.Exit(1)
	}

	err := download(*GET, *O)

	if err != nil {
		fmt.Print("doesn't work")
		os.Exit(1)

	}

	fmt.Printf("Download completed! Saved to %s\n", *O)

}

func download(url, filename string) error {

	fmt.Printf("Making GET request to: %s\n", url)

	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Print(nil)
		os.Exit(1)

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("bad status: %s", resp.Status)

	}

	output, err := os.Create(filename)

	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, resp.Body)

	if err != nil {
		log.Fatal()
	}

	return nil

}
