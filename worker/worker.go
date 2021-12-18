package worker

import (
	"gogo12306/httpcli"
)

var (
	ch chan *Item
)

func init() {
	ch = make(chan *Item)

	for i := 0; i < 100; i++ {
		go func(n int) {
			for item := range ch {
				if item.Callback != nil {
					item.Callback(httpcli.DoHttp(item.HttpReq))
				} else {
					httpcli.DoHttp(item.HttpReq)
				}
			}
		}(i)
	}
}

func Do(item *Item) {
	ch <- item
}
