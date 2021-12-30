package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QueryOrderWaitTime 查询订单排队等待时间
func QueryOrderWaitTime(jar *cookiejar.Jar) (orderID string, err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/queryOrderWaitTime?random=%d&tourFlag=dc&_json_att=&REPEAT_SUBMIT_TOKEN=%s"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url0, cdn.GetCDN(), time.Now().UnixMilli(), globalRepeatSubmitToken), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("查询订单排队等待时间错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("查询订单排队等待时间失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return "", errors.New("query order wait time failure")
	}

	logger.Debug("查询订单排队等待时间", zap.ByteString("body", body))

	type QueryOrderWaitTimeData struct {
		Count                    int    `json:"count"`
		OrderID                  string `json:"orderId,omitempty"`
		QueryOrderWaitTimeStatus bool   `json:"queryOrderWaitTimeStatus"`
		RequestID                int    `json:"requestId"`
		WaitCount                int    `json:"waitCount"`
		WaitTime                 int    `json:"waitTime"`
	}

	type QueryOrderWaitTimeResponse struct {
		Status   bool                   `json:"status"`
		Messages []string               `json:"messages"`
		Data     QueryOrderWaitTimeData `json:"data"`
	}
	response := QueryOrderWaitTimeResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析查询订单排队等待时间返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("查询订单排队等待时间失败", zap.Strings("错误消息", response.Messages))

		return "", errors.New(strings.Join(response.Messages, ""))
	}

	return response.Data.OrderID, nil
}
