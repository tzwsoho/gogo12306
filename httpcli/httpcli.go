package httpcli

import (
	"compress/gzip"
	"crypto/tls"
	"gogo12306/logger"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"go.uber.org/zap"
)

func DefaultHeaders(req *http.Request) {
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 Edg/96.0.1054.62"
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", "https://kyfw.12306.cn")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "kyfw.12306.cn")
	req.Host = "kyfw.12306.cn"
	req.URL.Host = "kyfw.12306.cn"
}

func GetBody(res *http.Response) (body []byte, err error) {
	resBody := res.Body
	defer resBody.Close()

	if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
		if resBody, err = gzip.NewReader(res.Body); err != nil {
			logger.Error("解压响应体错误", zap.Error(err))
			return nil, err
		}
	}

	body, err = ioutil.ReadAll(resBody)
	if err != nil {
		logger.Error("获取响应体错误", zap.Error(err))
		return
	}

	return
}

func DoHttp(req *http.Request, jar *cookiejar.Jar) (body []byte, ok bool, duration time.Duration, err error) {
	j := http.DefaultClient.Jar
	if jar != nil {
		j = jar
	}

	cli := http.Client{
		Jar:           j,
		Timeout:       time.Second * 10,
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

		body, err = GetBody(res)
		return
	}

	body, err = GetBody(res)
	return body, true, duration, err
}
