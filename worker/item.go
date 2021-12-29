package worker

import (
	"net/http"
)

type CB func(body []byte, statusCode int, err error)

type Item struct {
	HttpReq  *http.Request
	Callback CB
}
