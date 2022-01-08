package normal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

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
		QueryOrderWaitTimeStatus bool `json:"queryOrderWaitTimeStatus"`
		Count                    int  `json:"count"`
		RequestID                int  `json:"requestId"`
		WaitCount                int  `json:"waitCount"`
		WaitTime                 int  `json:"waitTime"`

		ErrorCode string `json:"errorcode,omitempty"`
		Msg       string `json:"msg,omitempty"`

		OrderID string `json:"orderId,omitempty"`
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
	} else if response.Data.ErrorCode != "" && response.Data.Msg != "" {
		logger.Error("查询订单排队等待时间错误",
			zap.String("errorCode", response.Data.ErrorCode),
			zap.String("msg", response.Data.Msg),
		)

		return "", errors.New(response.Data.Msg)
	}

	logger.Info("查询订单排队等待情况...",
		zap.Int("请求 ID", response.Data.RequestID),
		zap.Int("count", response.Data.Count),
		zap.Int("waitCount", response.Data.WaitCount),
		zap.Int("waitTime", response.Data.WaitTime),
		zap.String("订单号", response.Data.OrderID),
	)

	return response.Data.OrderID, nil
}
