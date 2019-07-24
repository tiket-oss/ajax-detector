package network

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"strings"
	"sync"

	ajaxdetector "github.com/tiket-libre/ajax-detector"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/sync/errgroup"
)

func formatEventLog(ev interface{}) []string {
	switch event := ev.(type) {
	case *network.EventRequestWillBeSent:
		referer := event.Request.Headers["Referer"].(string)
		return []string{referer, event.Request.URL, event.Request.Method}
	}

	return nil
}

// LogAjaxRequest will call monitorPageNetwork on every pages, logging the result using writer
func LogAjaxRequest(ctx context.Context, writer io.Writer, pages []ajaxdetector.PageInfo) error {
	eventLogs := [][]string{
		{"Referer", "URL", "Method"},
	}

	eventsChan := make(chan []interface{}, len(pages))
	group, ctx := errgroup.WithContext(ctx)

	// Create a new browser
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	for _, page := range pages {
		pageURL := page.URL // https://golang.org/doc/faq#closures_and_goroutines
		group.Go(func() error {
			// Create new tab for each page
			ctxt, cancel := chromedp.NewContext(ctx)
			defer cancel()

			events, err := MonitorPageNetwork(ctxt, pageURL)
			if err == nil {
				eventsChan <- events
			}

			return err
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	close(eventsChan)
	for events := range eventsChan {
		for _, event := range events {
			eventLogs = append(eventLogs, formatEventLog(event))
		}
	}

	csvWriter := csv.NewWriter(writer)
	return csvWriter.WriteAll(eventLogs)
}

// MonitorPageNetwork runs NavigateAction towards pageURL against a chromedp context
// while listening for network.EventRequestWillBeSent and network.EventResponseReceived event.
// All of those events object will be returned as the result.
func MonitorPageNetwork(ctx context.Context, pageURL string) ([]interface{}, error) {
	events := make([]interface{}, 0)

	var group sync.WaitGroup
	eventChan := make(chan interface{}, 8)
	signalFinish := make(chan int)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent:
			if event.Type == network.ResourceTypeFetch || event.Type == network.ResourceTypeXHR {
				group.Add(1)
				go func() {
					eventChan <- event
				}()
			}
		}
	})

	if err := chromedp.Run(ctx, network.Enable(), chromedp.Navigate(pageURL)); err != nil {
		return events, err
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

		group.Wait()
		signalFinish <- 0
	}()

Loop:
	for {
		select {
		case <-signalFinish:
			break Loop
		case event := <-eventChan:
			events = append(events, event)
			group.Done()
		}
	}

	return events, nil
}
