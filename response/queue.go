package response

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

// QueueIten is returned as a part of a response headers when the build is invoked
type QueueItem struct {
	URL *url.URL
	ID  uint
}

// NewQueueItemFromURL returns struct with URL and parsed location
func NewQueueItemFromURL(URL *url.URL) (*QueueItem, error) {
	// TODO: use precompiled regex
	pattern := regexp.MustCompile("queue/item/(?P<id>[0-9]+)/")
	if !pattern.MatchString(URL.Path) {
		return nil, fmt.Errorf("Returned URL (%v) doesn't match expected pattern", URL)
	} else {
		raw := pattern.FindStringSubmatch(URL.Path)[1]
		if u64, err := strconv.ParseUint(raw, 10, 0); err != nil {
			return nil, err
		} else {
			return &QueueItem{URL, uint(u64)}, nil
		}
	}
}
