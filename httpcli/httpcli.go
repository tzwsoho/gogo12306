package httpcli

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"
)

func DoHttp(req *http.Request) (body []byte, ok bool, duration time.Duration, err error) {
	cli := http.Client{
		Timeout:       time.Second * 3,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	var res *http.Response
	t0 := time.Now()
	res, err = cli.Do(req)
	duration = time.Now().Sub(t0)
	if err != nil {
		// logger.Error("HttpDo err",
		// 	zap.String("method", string(req.Header.Method())),
		// 	zap.String("url", req.URI().String()),
		// 	zap.Error(err))

		return nil, false, duration, err
	} else if res.StatusCode != http.StatusOK {
		// logger.Error("HttpDo statusCode err",
		// 	zap.String("method", string(req.Header.Method())),
		// 	zap.String("url", req.URI().String()),
		// 	zap.Int("statusCode", res.StatusCode()))

		body, err = ioutil.ReadAll(res.Body)
		return
	}

	body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return body, true, duration, nil
}
