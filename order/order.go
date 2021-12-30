package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type OrderInfo struct {
	SecretStr            string // 下单用的密钥
	TrainDate            string // 出发日期
	QueryFromStationName string // 出发站中文站名
	QueryToStationName   string // 到达站中文站名
}

// OrderTicket 一般下单请求，用于普通购票
func OrderTicket(jar *cookiejar.Jar, info *OrderInfo) (err error) {
	const (
		url0    = "https://%s/otn/leftTicket/submitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	secretStr := url.QueryEscape(info.SecretStr)
	trainDate := url.QueryEscape(info.TrainDate)
	purposeCodes := url.QueryEscape(passengerTypeToPurposeCodes())

	checkusermdId := "_json_att="
	if len(CheckUserMDID) > 0 {
		checkusermdId += url.QueryEscape(CheckUserMDID)
	}

	payload := "secretStr=" + secretStr +
		"&train_date=" + trainDate +
		"&back_train_date=" + url.QueryEscape(time.Now().Format("2006-01-02")) + // 返程日期，貌似可以是任意日期
		"&tour_flag=dc" + // dc: 单程
		"&purpose_codes=" + purposeCodes +
		"&query_from_station_name=" + info.QueryFromStationName + // 出发站中文站名
		"&query_to_station_name=" + info.QueryToStationName + // 到达站中文站名
		"&" + checkusermdId
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("提交订单错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("提交订单失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("submit order failure")
	}

	type OrderTicketResult struct {
		Status   bool     `json:"status"`
		Messages []string `json:"messages"`
		Data     string   `json:"data"`
	}
	result := OrderTicketResult{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析提交订单返回信息错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !result.Status {
		logger.Error("提交订单失败", zap.Strings("错误消息", result.Messages))

		return errors.New(strings.Join(result.Messages, ""))
	}

	logger.Info("提交订单", zap.ByteString("body", body))

	switch result.Data {
	case "N", "0": // 下单成功
		return

	default:
		logger.Error("下单失败", zap.ByteString("body", body))
		return errors.New("order ticket failure")
	}
}
