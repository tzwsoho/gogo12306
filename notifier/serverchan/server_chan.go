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

func Notify(cfg *config.ServerChan, msg string) (err error) {
	if !cfg.On {
		return
	}

	const title = "GOGO12306 购票消息"
	const scurl = "https://sctapi.ftqq.com/"

	payload := url.Values{}
	payload.Add("text", title)
	payload.Add("desp", msg)

	reqMsg := []byte(payload.Encode())
	buf := bytes.NewBuffer(reqMsg)
	req, _ := http.NewRequest("POST", scurl+cfg.SKey+".send", buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, nil); err != nil {
		logger.Error("Server 酱消息发送错误", zap.ByteString("body", reqMsg), zap.Error(err))

		return err
	} else if statusCode != http.StatusOK {
		logger.Error("Server 酱消息发送失败",
			zap.Int("statusCode", statusCode),
			zap.ByteString("req", reqMsg),
			zap.ByteString("res", body),
		)

		return errors.New("serverchan msg send failure")
	}

	logger.Info("Server 酱消息发送成功！")
	return
}
