package httpcli

import (
	"compress/gzip"
	"crypto/tls"
	"gogo12306/logger"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

func DefaultHeaders(req *http.Request) {
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 Edg/96.0.1054.62"
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Origin", "https://kyfw.12306.cn")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Host", "kyfw.12306.cn")
	req.Host = "kyfw.12306.cn"
	// req.URL.Host = "kyfw.12306.cn"
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

func DoHttp(req *http.Request, jar *cookiejar.Jar) (body []byte, statusCode int, err error) {
	j := http.DefaultClient.Jar
	if jar != nil {
		j = jar
	}

	cli := http.Client{
		Jar:     j,
		Timeout: time.Second * 10,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	t0 := time.Now()

	var res *http.Response
	res, err = cli.Do(req)

	logger.Debug("HTTP 耗时",
		zap.String("url", req.URL.String()),
		zap.Duration("耗时(秒)", time.Now().Sub(t0)),
	)

	if j != nil && len(res.Cookies()) > 0 {
		u, _ := url.Parse("https://kyfw.12306.cn" + req.URL.Path)
		j.SetCookies(u, res.Cookies())
	}

	if err != nil {
		// logger.Error("HttpDo err",
		// 	zap.String("method", string(req.Header.Method())),
		// 	zap.String("url", req.URI().String()),
		// 	zap.Error(err))

		return nil, http.StatusInternalServerError, err
	}

	body, err = GetBody(res)
	return body, res.StatusCode, err
}
