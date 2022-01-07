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

type GetQueueNumRequest struct {
}

// GetQueueNum 获取候补人数信息
func GetQueueNum(jar *cookiejar.Jar, info *GetQueueNumRequest) (err error) {
	const (
		url0    = "https://%s/otn/afterNate/getQueueNum"
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
		logger.Error("获取候补人数信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取候补人数信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("passenger init api failure")
	}

	logger.Debug("获取候补人数信息", zap.ByteString("body", body))

	type GetQueueNumInfo struct {
		QueueInfo        string `json:"queue_info"`
		QueueLevel       int    `json:"queue_level,string"`
		SeatTypeCode     string `json:"seat_type_code"`
		StationTrainCode string `json:"station_train_code"`
		TrainDate        string `json:"train_date"`
		TrainNo          string `json:"train_no"`
	}

	type GetQueueNumData struct {
		Flag     bool              `json:"flag"`
		Msg      string            `json:"msg,omitempty"`
		QueueNum []GetQueueNumInfo `json:"queueNum,omitempty"`
	}

	type GetQueueNumResponse struct {
		Status   bool            `json:"status"`
		Messages []string        `json:"messages"`
		Data     GetQueueNumData `json:"data"`
	}
	response := GetQueueNumResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取候补人数信息返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("获取候补人数信息失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.Flag {
		logger.Error("获取候补人数信息失败", zap.ByteString("body", body))

		return errors.New(response.Data.Msg)
	}

	logger.Info("获取候补人数信息",
		zap.Any("候补信息", response.Data.QueueNum),
	)

	return
}
