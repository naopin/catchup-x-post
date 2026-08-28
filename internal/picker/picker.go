package picker

import (
	"fmt"

	"catchup-x-post/internal/history"
	"catchup-x-post/internal/newsclient"
)

func Pick(items []newsclient.NewsItem, store *history.Store) (*newsclient.NewsItem, error) {
	for i := range items {
		item := &items[i]
		if item.URL == "" || item.Title == "" || item.Summary == "" {
			continue
		}
		if store.IsSimilar(item.URL, item.Title, item.Topic) {
			continue
		}
		return item, nil
	}
	return nil, fmt.Errorf("no eligible news item (history=%d candidates=%d)", store.Count(), len(items))
}
