package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/gokrazy/rsync/rsyncclient"
)

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

func syncFile(ctx context.Context, wg *sync.WaitGroup, serverAddr string, file string, dest string) {
	defer wg.Done()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Printf("[ERROR] Cannot connect to %s: %v\n", serverAddr, err)
		return
	}
	defer conn.Close()

	rsyncArgs := []string{"rsync", file, dest}
	client, err := rsyncclient.New(rsyncArgs)
	if err != nil {
		log.Printf("[ERROR] Creating rsync client: %v\n", err)
		return
	}

	if _, err := client.Run(ctx, conn, []string{file}); err != nil {
		log.Printf("[ERROR] Sync failed for %s: %v\n", file, err)
		return
	}

	md5sum, err := computeMD5(file)
	if err != nil {
		log.Printf("[ERROR] Computing MD5 for %s: %v\n", file, err)
	} else {
		log.Printf("[OK] Synced %s with MD5 %s\n", file, md5sum)
	}
}

func main() {
	var (
		dir        string
		serverAddr string
		destPath   string
	)

	flag.StringVar(&dir, "src", ".", "Directory to backup")
	flag.StringVar(&serverAddr, "host", "localhost:873", "Remote rsync server address")
	flag.StringVar(&destPath, "dest", "/backup", "Destination path on remote")
	flag.Parse()

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
		log.Fatalf("[FATAL] Failed to walk source directory: %v\n", err)
	}

	wg.Wait()
	fmt.Println("✅ femsync complete.")
}
