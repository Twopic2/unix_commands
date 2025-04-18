package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {

	fmt.Println("Welcome to UWURL!")

	GET := flag.String("GET", "", "GET protocol")
	flag.Parse()

	if *GET == "" {
		fmt.Print("Make sure to follow the intstructions")
		os.Exit(1)
	}

	client := &http.Client{}

	fmt.Printf("Making GET request to: %s\n", *GET)

	resp, err := client.Get(*GET)

	if err != nil {
		fmt.Print(nil)
		os.Exit(1)
	}

	defer resp.Body.Close()

	fmt.Printf("Status: %s\n\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(nil)
		os.Exit(1)
	}

	fmt.Println(string(body))
}
