package worker

import (
	"gogo12306/httpcli"
	"net/http/cookiejar"
)

var (
	ch chan *Item
)

func init() {
	ch = make(chan *Item)

	for i := 0; i < 100; i++ {
		go func(n int) {
			jar, _ := cookiejar.New(nil)

			for item := range ch {
				if item.Callback != nil {
					item.Callback(httpcli.DoHttp(item.HttpReq, jar))
				} else {
					httpcli.DoHttp(item.HttpReq, jar)
				}
			}
		}(i)
	}
}

func Do(item *Item) {
	ch <- item
}
