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
	"time"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/order/common"

	"go.uber.org/zap"
)

type SubmitOrderRequest struct {
	SecretStr            string // 下单用的密钥
	TrainDate            string // 出发日期
	QueryFromStationName string // 出发站中文站名
	QueryToStationName   string // 到达站中文站名
}

// SubmitOrder 一般下单请求，用于普通购票
func SubmitOrder(jar *cookiejar.Jar, request *SubmitOrderRequest) (err error) {
	const (
		url0    = "https://%s/otn/leftTicket/submitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	payload := url.Values{}
	payload.Add("secretStr", request.SecretStr)
	payload.Add("train_date", request.TrainDate)
	payload.Add("back_train_date", time.Now().Format("2006-01-02")) // 返程日期，貌似可以是任意日期
	payload.Add("tour_flag", "dc")                                  // dc: 单程
	payload.Add("purpose_codes", common.PassengerTypeToPurposeCodes())
	payload.Add("query_from_station_name", request.QueryFromStationName) // 出发站中文站名
	payload.Add("query_to_station_name", request.QueryToStationName)     // 到达站中文站名
	payload.Add("undefined", "")

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
		logger.Error("提交订单错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("提交订单失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("submit order failure")
	}

	logger.Debug("提交订单", zap.ByteString("body", body))

	type OrderTicketResponse struct {
		Status   bool     `json:"status"`
		Messages []string `json:"messages"`
		Data     string   `json:"data"`
	}
	response := OrderTicketResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析提交订单返回信息错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("提交订单失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	}

	switch response.Data {
	case "N", "0": // 下单成功
		return

	default:
		logger.Error("下单失败", zap.ByteString("body", body))
		return errors.New("order ticket failure")
	}
}
