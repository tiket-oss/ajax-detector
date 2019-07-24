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

func formatRequestLog(request network.Request) []string {
	referer := request.Headers["Referer"].(string)
	return []string{referer, request.URL, request.Method}
}

// LogAjaxRequest will call monitorPageNetwork on every pages, logging the result using writer
func LogAjaxRequest(ctx context.Context, writer io.Writer, pages []ajaxdetector.PageInfo) error {
	requestLogs := [][]string{
		{"Referer", "URL", "Method"},
	}

	requestsChan := make(chan []network.Request, len(pages))
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

			requests, err := MonitorPageNetwork(ctxt, pageURL)
			if err == nil {
				requestsChan <- requests
			}

			return err
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	close(requestsChan)
	for requests := range requestsChan {
		for _, request := range requests {
			requestLogs = append(requestLogs, formatRequestLog(request))
		}
	}

	csvWriter := csv.NewWriter(writer)
	return csvWriter.WriteAll(requestLogs)
}

// MonitorPageNetwork runs NavigateAction towards pageURL against a chromedp context
// while listening for network.EventRequestWillBeSent event.
// All of those event's Request object will be returned as the result.
func MonitorPageNetwork(ctx context.Context, pageURL string) ([]network.Request, error) {
	requests := make([]network.Request, 0)

	var group sync.WaitGroup
	reqChan := make(chan network.Request, 4)
	signalFinish := make(chan int)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent:
			if event.Type == network.ResourceTypeFetch || event.Type == network.ResourceTypeXHR {
				group.Add(1)

				go func() {
					request := *event.Request // https://golang.org/doc/faq#closures_and_goroutines
					reqChan <- request
				}()
			}
		}
	})

	if err := chromedp.Run(ctx, network.Enable(), chromedp.Navigate(pageURL)); err != nil {
		return requests, err
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
		case requestLog := <-reqChan:
			requests = append(requests, requestLog)
			group.Done()
		}
	}

	return requests, nil
}
