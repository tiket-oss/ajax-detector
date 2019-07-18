// Command click is a chromedp example demonstrating how to use a selector to
// click on an element.
package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(v interface{}) {
		log.Printf("Event triggered: %T\n", v)

		if event, ok := v.(*network.EventRequestWillBeSent); ok {
			requestInfo, err := event.Request.MarshalJSON()
			if err != nil {
				log.Fatal(err)
			}

			log.Println(string(requestInfo))
		}
	})

	// navigate to a page, wait for an element, click
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(`https://www.tiket.com/kereta-api`),
	)
	if err != nil {
		log.Fatal(err)
	}
}
