package http

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/tzwsoho/gogo12306/pkg/logger"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func Get(url string, cookies cookiejar.Jar) (response string, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	cli := fasthttp.HostClient{
		IsTLS: true,
		Addr:  string(req.URI().Host()) + ":443",
		TLSConfig: &tls.Config{
			ServerName:         string(req.URI().Host()),
			InsecureSkipVerify: true,
		},
	}

	err = cli.DoTimeout(req, res, time.Second*20)
	if err != nil {
		logger.Error("HttpGet err", zap.String("url", url), zap.Error(err))
		return "", err
	} else if res.StatusCode() != http.StatusOK {
		logger.Error("HttpGet errCode", zap.String("url", url), zap.Int("statusCode", res.StatusCode()))
		return "", err
	}

	return string(res.Body()), nil
}
