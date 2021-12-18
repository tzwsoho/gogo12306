package worker

import (
	"net/http"
	"time"
)

type CB func(body []byte, ok bool, duration time.Duration, err error)

type Item struct {
	HttpReq  *http.Request
	Callback CB
}
