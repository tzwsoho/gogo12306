package normal

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type ResultOrderForDcQueueRequest struct {
	OrderID string
}

// ResultOrderForDcQueue 获取下单最后的结果
func ResultOrderForDcQueue(jar *cookiejar.Jar, info *ResultOrderForDcQueueRequest) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/resultOrderForDcQueue"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)
	payload := url.Values{}
	payload.Add("orderSequence_no", info.OrderID)
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

	logger.Debug("获取下单最后的结果", zap.ByteString("body", body))
	return
}
