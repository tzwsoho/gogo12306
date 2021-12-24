package serverchan

import (
	"bytes"
	"errors"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

func Notify(msg string) (err error) {
	if !config.Cfg.Notifier.ServerChan.On {
		return
	}

	const title = "GOGO12306 购票消息"
	const scurl = "https://sctapi.ftqq.com/"
	msgBody := []byte("text=" + url.QueryEscape(title) + "&desp=" + url.QueryEscape(msg))
	buf := bytes.NewBuffer(msgBody)

	req, _ := http.NewRequest("POST", scurl+config.Cfg.Notifier.ServerChan.SKey+".send", buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, nil); err != nil {
		logger.Error("Server 酱消息发送错误", zap.ByteString("body", msgBody), zap.Error(err))

		return err
	} else if !ok {
		logger.Error("Server 酱消息发送失败", zap.ByteString("body", msgBody), zap.ByteString("res", body))

		return errors.New("serverchan msg send failure")
	}

	logger.Debug("Server 酱消息发送成功！")
	return
}
