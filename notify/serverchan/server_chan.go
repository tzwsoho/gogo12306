package serverchan

import (
	"bytes"
	"errors"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/zap"
)

func Notify(msg string) (err error) {
	if !config.Cfg.Notifier.ServerChan.On {
		return
	}

	const title = "GOGO12306 购票消息"
	const scurl = "https://sctapi.ftqq.com/"
	buf := bytes.NewBuffer([]byte("text=" + url.QueryEscape(title) +
		"&desp=" + url.QueryEscape(msg)))

	req, _ := http.NewRequest("POST", scurl+config.Cfg.Notifier.ServerChan.SKey+".send", buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var (
		body []byte
		ok   bool
	)
	jar, _ := cookiejar.New(nil)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("发送 Server 酱消息错误", zap.Error(err))

		return err
	} else if !ok {
		logger.Error("发送 Server 酱消息失败", zap.ByteString("res", body))

		return errors.New("serverchan msg send failure")
	}

	return
}
