package notifier

import (
	"gogo12306/notifier/serverchan"
	"gogo12306/notifier/wxpusher"
)

// Broadcast 广播刷票成功的消息
func Broadcast(msg string) (err error) {
	if err = serverchan.Notify(msg); err != nil {
		return
	}

	if err = wxpusher.Notify(msg); err != nil {
		return
	}

	return
}
