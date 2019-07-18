package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {

	pageURL := flag.String("page-url", "", "URL of the page to be profiled")
	filePath := flag.String("file-path", "output.txt", "Path of the file to be written to")

	flag.Parse()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	logFile, err := os.Create(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	MonitorPageNetwork(ctx, logFile, *pageURL)
}
