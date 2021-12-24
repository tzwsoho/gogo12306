package wxpusher

import (
	"bytes"
	"encoding/json"
	"errors"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"

	"go.uber.org/zap"
)

func Notify(msg string) (err error) {
	if !config.Cfg.Notifier.WXPusher.On {
		return
	}

	type wxPusherBody struct {
		AppToken    string   `json:"appToken"`
		Content     string   `json:"content"`
		Summary     string   `json:"summary"`
		ContentType int64    `json:"contentType"` // 1 - 文字，2 - HTML，3 - MARKDOWN
		TopicIDs    []int64  `json:"topicIds"`
		UIDs        []string `json:"uids"`
	}

	wxBody := wxPusherBody{
		AppToken:    config.Cfg.Notifier.WXPusher.AppToken,
		Content:     msg,
		Summary:     "GOGO12306 购票消息",
		ContentType: 1,
		TopicIDs:    config.Cfg.Notifier.WXPusher.TopicIDs,
		UIDs:        config.Cfg.Notifier.WXPusher.UIDs,
	}
	var msgBody []byte
	if msgBody, err = json.Marshal(&wxBody); err != nil {
		logger.Error("序列化消息体错误", zap.Any("body", wxBody), zap.Error(err))
	}

	const (
		url = "http://wxpusher.zjiecode.com/api/send/message"
	)
	buf := bytes.NewBuffer(msgBody)
	req, _ := http.NewRequest("POST", url, buf)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, nil); err != nil {
		logger.Error("WXPusher 消息发送错误", zap.ByteString("body", msgBody), zap.Error(err))

		return err
	} else if !ok {
		logger.Error("WXPusher 消息发送失败", zap.ByteString("body", msgBody), zap.ByteString("res", body))

		return errors.New("wxPusher msg send failure")
	}

	logger.Debug("WXPusher 消息发送成功！")
	return
}
