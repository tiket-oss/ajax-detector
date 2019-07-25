package network

import (
	"strconv"

	"github.com/chromedp/cdproto/network"
	"github.com/dustin/go-humanize"
)

var csvHeader = []string{"Page", "URL", "Status", "Resource Type", "Method", "Size", "Time"}

type networkRoundTrip struct {
	requestEvent  *network.EventRequestWillBeSent
	responseEvent *network.EventResponseReceived
}

func (rt networkRoundTrip) sourcePage() string {
	return rt.requestEvent.Request.Headers["Referer"].(string)
}

func (rt networkRoundTrip) url() string {
	return rt.requestEvent.Request.URL
}

func (rt networkRoundTrip) status() string {
	return strconv.FormatInt(rt.responseEvent.Response.Status, 10)
}

func (rt networkRoundTrip) resourceType() string {
	return rt.requestEvent.Type.String()
}

func (rt networkRoundTrip) method() string {
	return rt.requestEvent.Request.Method
}

func (rt networkRoundTrip) size() string {
	return humanize.Bytes(
		uint64(rt.responseEvent.Response.EncodedDataLength))
}

func (rt networkRoundTrip) time() string {
	responseTime := rt.responseEvent.Timestamp.Time()
	requestTime := rt.requestEvent.Timestamp.Time()

	timeDiff := responseTime.Sub(requestTime)

	return timeDiff.String()
}

func (rt networkRoundTrip) formatLog() []string {
	return []string{
		rt.sourcePage(),   // Page
		rt.url(),          // URL
		rt.status(),       // Status
		rt.resourceType(), // Resource Type
		rt.method(),       // Method
		rt.size(),         // Size
		rt.time(),         // Time
	}
}
