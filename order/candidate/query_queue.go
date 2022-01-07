package candidate

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type QueryQueueRequest struct {
}

// QueryQueue 查询候补结果
func QueryQueue(jar *cookiejar.Jar, info *QueryQueueRequest) (err error) {
	const (
		url0    = "https://%s/otn/afterNate/queryQueue"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("查询候补结果错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("查询候补结果失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("query queue failure")
	}

	logger.Debug("查询候补结果", zap.ByteString("body", body))

	type QueryQueueData struct {
		JZDHDateS      string   `json:"jzdhDateS"` // jzdh = 截止兑换
		JZDHHourS      string   `json:"jzdhHourS"`
		JZDHDateE      string   `json:"jzdhDateE"`
		JZDHHourE      string   `json:"jzdhHourE"`
		JZDHDiffSelect []string `json:"jzdhDiffSelect"` // 可选的截止兑换时间（单位：分钟，默认：360）：120/180/360/720
	}

	type QueryQueueResponse struct {
		Status   bool           `json:"status"`
		Messages []string       `json:"messages"`
		Data     QueryQueueData `json:"data"`
	}
	response := QueryQueueResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析查询候补结果返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("查询候补结果失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	}

	logger.Info("查询候补结果",
		zap.String("候补开始日期", response.Data.JZDHDateS+" "+response.Data.JZDHHourS),
		zap.String("截止兑换日期", response.Data.JZDHDateE+" "+response.Data.JZDHHourE),
		zap.Strings("可选的截止兑换时间", response.Data.JZDHDiffSelect),
	)

	return
}
