package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
)

func connectSSH(user, host, keyPlace string) (*ssh.Client, error) {

	priKey, err := os.ReadFile(keyPlace)

	if err != nil {
		log.Fatal("There's soemthing wrong with the private key")
	}

	_, err = ssh.ParsePrivateKey(priKey)

	if err != nil {
		log.Fatal("There's soemthing wrong with the signature.")
	}

	clientKey, err := auth.PrivateKey(user, keyPlace, ssh.InsecureIgnoreHostKey())

	if err != nil {
		log.Fatal("There's soemthing wrong with the signature.")
	}

	client, err := ssh.Dial("tcp", host, &clientKey)

	if err != nil {
		log.Fatal("There's soemthing wrong with the signature.")
	}

	return client, err

}

func main() {

	clientSSH := flag.String("clientSSH", "root", "Client")
	server := flag.String("server", "localhost:22", "SSH server")
	keyPath := flag.String("key", os.Getenv("HOME")+"/.ssh/id_ed25519.pub", "Path to private key")
	source := flag.String("source", "", "Path to local file to copy")
	destination := flag.String("destination", "", "Destination path on remote server")
	flag.Parse()

	if clientSSH == nil {
		log.Fatal("The client SSH sesion doesn't work")
	}

	clientconnect, err := connectSSH(*clientSSH, *server, *keyPath)

	if err != nil {

		os.Exit(1)
	}

	defer clientconnect.Close()

	scpStart, err := scp.NewClientBySSH(clientconnect)
	if err != nil {
		log.Fatal()
	}

	defer scpStart.Close()

	file, err := os.Open(*source)

	if err != nil {
		log.Fatal("There's soemthing wrong with the signature.")
		os.Exit(1)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = scpStart.CopyFile(ctx, file, *destination, "0655")

	if err != nil {
		if err != nil {
			log.Fatalf("❌ SCP file transfer failed: %v", err)
		}
	}

	fmt.Println("It worked")

}
