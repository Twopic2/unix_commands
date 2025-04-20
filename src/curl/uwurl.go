package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	data = flag.String("data", "", "POST request")
	POST = flag.String("POST", "", "POST protocol")
	O    = flag.String("O", "output", "O is the output file")
	GET  = flag.String("GET", "", "GET protocol")
)

func main() {

	fmt.Println("Welcome to UWURL!")
	flag.Parse()

	if *POST == "" && *GET == "" {
		fmt.Println(" uwurl -POST https://example.com -data Some info\n uwurl -GET https://example.com -O example.txt")
		os.Exit(1)
	}

	switch {
	case *GET != "":

		err := download(*GET, *O)

		if err != nil {
			fmt.Print("doesn't work")
			os.Exit(1)
		}

		go download(*GET, *O)

		fmt.Printf("Downloaded to: %s\n", *O)

	case *POST != "":

		resBody, err := httpPost(*POST, *data)

		if err != nil {
			fmt.Print("doesn't work")
			os.Exit(1)
		}

		fmt.Printf("Reponse to %s", resBody)

	default:
		fmt.Println("Usage:")
		fmt.Println("  -GET <url> -O <output_file>")
		fmt.Println("  -POST <url> -data <body>")
		os.Exit(1)

	}

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

func httpPost(url, body string) (string, error) {

	fmt.Printf("Making POST request to: %s\n", url)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))

	if err != nil {
		fmt.Print(nil)
		os.Exit(1)

	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Print(nil)

	}

	stringToBytes := string(bytes)

	return stringToBytes, nil
}
