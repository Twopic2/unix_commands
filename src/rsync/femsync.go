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

func syncFile(ctx context.Context, wg *sync.WaitGroup, serverAddr string, file string, dest string) {
	defer wg.Done()

	remote := fmt.Sprintf("%s:%s", serverAddr, dest)
	cmd := exec.CommandContext(ctx, "rsync", "-avz", "-e", "ssh", file, remote)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] rsync failed for %s: %v\nOutput: %s\n", file, err, string(out))
		return
	}

	md5sum, err := computeMD5(file)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("[OK] Synced %s → %s with MD5 %s\n", file, remote, md5sum)
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
