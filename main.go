package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/hawari17/page-profiler/network"
)

func createOutputFile(filePath string) (io.Writer, error) {
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	return os.Create(filePath)
}

func main() {

	pageURL := flag.String("page-url", "", "URL of the page to be profiled")
	filePath := flag.String("file-path", "out/result.txt", "Path of the file to be written to")

	flag.Parse()

	outFile, err := createOutputFile(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	network.MonitorPageNetwork(ctx, outFile, *pageURL)
}
