package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"sync"
)

// This isn't horrible but I'm tired
func computeMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func firstLogin() {

	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(home, ".femsync_first_run")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("It looks like this is your first time using the tool.")
		fmt.Println("Make sure the rsync/SSH daemon is configured and accessible via SSH on the remote server.")
		fmt.Println("You can use the -src, -host, and -dest flags to specify your backup source and destination.")

		file, err := os.Create(configPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

}

func syncFile(ctx context.Context, wg *sync.WaitGroup, remoteHost string, file string, dest string) {
	defer wg.Done()

	destPath := fmt.Sprintf("%s:%s", remoteHost, dest)
	cmd := exec.CommandContext(ctx, "rsync", "-az", "-e", "ssh", file, destPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] rsync failed for %s: %v\nOutput: %s", file, err, output)
		return
	}

	md5sum, err := computeMD5(file)
	if err != nil {
		log.Printf("[ERROR] Computing MD5 for %s: %v\n", file, err)
	} else {
		log.Printf("[OK] Synced %s → %s with MD5 %s\n", file, destPath, md5sum)
	}
}

func main() {
	var (
		dir        string
		serverAddr string
		destPath   string
	)

	fmt.Println("Welcome to femSync!")

	flag.StringVar(&dir, "src", ".", "Directory to backup")
	flag.StringVar(&serverAddr, "server", "user@remote", "Remote SSH server (user@host)")
	flag.StringVar(&destPath, "destination", "/backup", "Destination path on remote")
	flag.Parse()

	firstLogin()

	var wg sync.WaitGroup
	ctx := context.Background()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		wg.Add(1)
		go syncFile(ctx, &wg, serverAddr, path, destPath)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

}
