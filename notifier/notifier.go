package notifier

import (
	"gogo12306/notifier/serverchan"
	"gogo12306/notifier/wxpusher"
)

func Broadcast(msg string) (err error) {
	if err = serverchan.Notify(msg); err != nil {
		return
	}

	if err = wxpusher.Notify(msg); err != nil {
		return
	}

	return
}
