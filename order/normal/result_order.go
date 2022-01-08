package normal

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

type ResultOrderForDcQueueRequest struct {
	OrderID string
}

// ResultOrderForDcQueue 获取下单最后的结果
func ResultOrderForDcQueue(jar *cookiejar.Jar, request *ResultOrderForDcQueueRequest) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/resultOrderForDcQueue"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)
	payload := url.Values{}
	payload.Add("orderSequence_no", request.OrderID)
	payload.Add("_json_att", "")
	payload.Add("REPEAT_SUBMIT_TOKEN", globalRepeatSubmitToken)

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
		logger.Error("获取下单最后的结果错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取下单最后的结果失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("auto submit order failure")
	}

	type ResultOrderForDcQueueData struct {
		SubmitStatus bool `json:"submitStatus"`
	}

	type ResultOrderForDcQueueResponse struct {
		Status   bool                      `json:"status"`
		Messages []string                  `json:"messages"`
		Data     ResultOrderForDcQueueData `json:"data"`
	}
	response := ResultOrderForDcQueueResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取下单最后的结果返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("获取下单最后的结果失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	}

	logger.Info("获取下单最后的结果", zap.Bool("结果", response.Data.SubmitStatus))
	return
}
