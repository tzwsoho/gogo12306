package candidate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type SubmitOrderRequest struct {
	SecretStr string
}

// SubmitOrder 提交候补订单请求
func SubmitOrder(jar *cookiejar.Jar, request *SubmitOrderRequest) (err error) {
	const (
		url0    = "https://%s/otn/afterNate/submitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := &url.Values{}
	payload.Add("secretList", request.SecretStr)
	payload.Add("_json_att", "")

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("提交候补订单请求错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("提交候补订单请求失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("submit candidate order request failure")
	}

	logger.Debug("提交候补订单请求", zap.ByteString("body", body))

	type SubmitOrderData struct {
		Flag bool   `json:"flag"`
		Msg  string `json:"msg,omitempty"`
	}

	type SubmitOrderResponse struct {
		Status   bool            `json:"status"`
		Messages []string        `json:"messages"`
		Data     SubmitOrderData `json:"data"`
	}
	response := SubmitOrderResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析提交候补订单请求返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("提交候补订单请求失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.Flag {
		logger.Error("提交候补订单请求失败", zap.ByteString("body", body))

		return errors.New(response.Data.Msg)
	}

	return
}
