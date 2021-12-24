package worker

import (
	"gogo12306/httpcli"
	"gogo12306/logger"
)

const GOROUTINE_MAX = 50

var (
	ch         chan *Item
	goroutines int
)

func init() {
	ch = make(chan *Item, GOROUTINE_MAX)
}

func Do(item *Item) {
	ch <- item

	if goroutines >= GOROUTINE_MAX {
		return
	}

	go func(n int) {
		for item := range ch {
			if item.Callback != nil {
				item.Callback(httpcli.DoHttp(item.HttpReq, nil))
			} else {
				httpcli.DoHttp(item.HttpReq, nil)
			}
		}

		logger.Debug("Exit")
	}(goroutines)

	goroutines++
}
