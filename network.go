package main

import (
	"context"
	"io"
	"log"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func MonitorPageNetwork(ctx context.Context, writer io.Writer, pageURL string) {

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if event, ok := v.(*network.EventRequestWillBeSent); ok {
			requestInfo, err := event.Request.MarshalJSON()
			if err != nil {
				log.Fatal(err)
			}

			newline := "\n"
			requestInfo = append(requestInfo, newline...)
			if _, err := writer.Write(requestInfo); err != nil {
				log.Fatal(err)
			}
		}
	})

	// navigate to a page, wait for an element, click
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(pageURL),
	)
	if err != nil {
		log.Fatal(err)
	}
}
