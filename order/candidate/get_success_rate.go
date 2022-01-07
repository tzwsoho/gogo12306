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

type GetSuccessRateRequest struct {
	SecretStr string
}

// GetSuccessRate 获取人脸识别核验后的成功信息
func GetSuccessRate(jar *cookiejar.Jar, request *GetSuccessRateRequest) (trainNos []string, info string, err error) {
	const (
		url0    = "https://%s/otn/afterNate/getSuccessRate"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := &url.Values{}
	payload.Add("successSecret", request.SecretStr)
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
		logger.Error("获取人脸识别核验后的成功信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取人脸识别核验后的成功信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return nil, "", errors.New("get success rate failure")
	}

	logger.Debug("获取人脸识别核验后的成功信息", zap.ByteString("body", body))

	type GetSuccessRateFlag struct {
		TrainNo string `json:"train_no"`
		Info    string `json:"info"`
	}

	type GetSuccessRateData struct {
		Flag []GetSuccessRateFlag `json:"flag"`
	}

	type GetSuccessRateResponse struct {
		Status   bool               `json:"status"`
		Messages []string           `json:"messages"`
		Data     GetSuccessRateData `json:"data"`
	}
	response := GetSuccessRateResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取人脸识别核验后的成功信息返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("获取人脸识别核验后的成功信息失败", zap.Strings("错误消息", response.Messages))

		return nil, "", errors.New(strings.Join(response.Messages, ""))
	} else if len(response.Data.Flag) < 1 {
		logger.Error("获取人脸识别核验后的成功信息结果: Flag 数量有误",
			zap.ByteString("body", body),
		)

		return nil, "", errors.New(strings.Join(response.Messages, ""))
	}

	logger.Info("获取人脸识别核验后的成功信息结果",
		zap.String("候补情况", response.Data.Flag[0].Info),
		zap.String("列车号", response.Data.Flag[0].TrainNo),
	)

	info = response.Data.Flag[0].Info
	for _, flag := range response.Data.Flag {
		trainNos = append(trainNos, flag.TrainNo)
	}

	return
}
