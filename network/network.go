package network

import (
	"context"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// MonitorPageNetwork runs NavigateAction towards pageURL against a chromedp context
// while listening for network.EventRequestWillBeSent event.
// All of those event's Request object will be written as the result.
func MonitorPageNetwork(ctx context.Context, writer io.Writer, pageURL string) {
	writeChan := make(chan []byte, 4)
	signalFinish := make(chan int)
	var wg sync.WaitGroup

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*network.EventResponseReceived); ok {
			wg.Add(1)
			go func(event *network.EventResponseReceived) {
				responseInfo, err := event.Response.MarshalJSON()
				if err != nil {
					log.Fatal(err)
				}

				newline := "\n"
				responseInfo = append(responseInfo, newline...)

				writeChan <- responseInfo
			}(ev)
		}
	})

	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(pageURL),
	); err != nil {
		log.Fatal(err)
	}

	/*
		Since chromedp.Navigate does not wait for the page to be fully loaded
		we wait manually, there may be a better and more reliable way to do this
	*/
	state := "notloaded"
	for {
		script := `document.readyState`
		err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(script, &state))
		if err != nil {
			log.Println(err)
		}
		if strings.Compare(state, "complete") == 0 {
			break
		}
	}

	go func() {
		wg.Wait()
		signalFinish <- 0
	}()

	var response []byte
	var done bool
	for done == false {
		select {
		case responseLog := <-writeChan:
			response = append(response, responseLog...)
			wg.Done()
		case <-signalFinish:
			done = true
		}
	}

	if _, err := writer.Write(response); err != nil {
		log.Fatal(err)
	}
}
