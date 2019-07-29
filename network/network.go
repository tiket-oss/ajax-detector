package network

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	ajaxdetector "github.com/tiket-libre/ajax-detector"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/sync/errgroup"
)

// getResourceType is a makeshift function to retrieve the ResourceType value
// from various event. Think about an interface, but since these events don't share any,
// this function serve a similar purpose.
func getResourceType(event interface{}) network.ResourceType {
	switch ev := event.(type) {
	case *network.EventRequestWillBeSent:
		return ev.Type
	case *network.EventResponseReceived:
		return ev.Type
	default:
		return ""
	}
}

// getResourceType is a makeshift function to retrieve the RequestID value
// from various event. Think about an interface, but since these events don't share any,
// this function serve a similar purpose.
func getRequestID(event interface{}) network.RequestID {
	switch ev := event.(type) {
	case *network.EventRequestWillBeSent:
		return ev.RequestID
	case *network.EventResponseReceived:
		return ev.RequestID
	default:
		return ""
	}
}

func pairRequestEvent(event interface{}, group map[string]networkRoundTrip) {
	requestID := string(getRequestID(event))

	relEvent, ok := group[requestID]
	if !ok {
		relEvent = networkRoundTrip{}
	}

	switch ev := event.(type) {
	case *network.EventRequestWillBeSent:
		relEvent.requestEvent = ev
	case *network.EventResponseReceived:
		relEvent.responseEvent = ev
	}

	group[requestID] = relEvent
}

// LogAjaxRequest will call monitorPageNetwork on every pages, logging the result using writer
func LogAjaxRequest(ctx context.Context, writer io.Writer, pages []ajaxdetector.PageInfo) error {
	eventLogs := [][]string{csvHeader}

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
		eventGroup := make(map[string]networkRoundTrip)
		for _, event := range events {
			pairRequestEvent(event, eventGroup)
		}

		for _, relatedEvent := range eventGroup {
			if relatedEvent.requestEvent != nil && relatedEvent.responseEvent != nil {
				eventLogs = append(eventLogs, relatedEvent.formatLog())
			}
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
	eventChan := make(chan interface{}, 16)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent, *network.EventResponseReceived:
			resourceType := getResourceType(event)
			if resourceType == network.ResourceTypeFetch || resourceType == network.ResourceTypeXHR {
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

		select {
		case <-ctx.Done():
			log.Printf("Sleep canceled by context")
		case <-time.After(1 * time.Second):
			log.Printf("Waited 1 second to ensure all network call has been made...")
		}

		group.Wait()
		close(eventChan)
	}()

	for event := range eventChan {
		events = append(events, event)
		group.Done()
	}

	return events, nil
}
