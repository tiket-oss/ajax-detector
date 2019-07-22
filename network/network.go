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

func formatRequestLog(request *network.Request) ([]byte, error) {
	requestLog, err := request.MarshalJSON()
	if err == nil {
		newline := "\n"
		requestLog = append(requestLog, newline...)
	}

	return requestLog, err
}

// MonitorPageNetwork runs NavigateAction towards pageURL against a chromedp context
// while listening for network.EventRequestWillBeSent event.
// All of those event's Request object will be written as the result.
func MonitorPageNetwork(ctx context.Context, writer io.Writer, pageURL string) {
	writeChan := make(chan []byte, 4)
	signalFinish := make(chan int)
	var wg sync.WaitGroup

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*network.EventRequestWillBeSent); ok && (ev.Type == network.ResourceTypeFetch || ev.Type == network.ResourceTypeXHR) {
			wg.Add(1)
			go func(event *network.EventRequestWillBeSent) {
				requestLog, err := formatRequestLog(event.Request)
				if err != nil {
					log.Fatal(err)
				}

				writeChan <- requestLog
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
	go func() {
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

		wg.Wait()
		signalFinish <- 0
	}()

	var request []byte
	var done bool
	for done == false {
		select {
		case requestLog := <-writeChan:
			request = append(request, requestLog...)
			wg.Done()
		case <-signalFinish:
			done = true
		}
	}

	if _, err := writer.Write(request); err != nil {
		log.Fatal(err)
	}
}
