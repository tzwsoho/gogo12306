package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/common"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"go.uber.org/zap"
)

var passengers map[string]*common.PassengerInfo

func init() {
	passengers = make(map[string]*common.PassengerInfo)
}

func GetPassengerList(jar *cookiejar.Jar) (err error) {
	const (
		url     = "https://%s/otn/confirmPassenger/getPassengerDTOs"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN()), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取乘客列表请求错误", zap.Error(err))

		return err
	} else if statusCode != http.StatusOK {
		logger.Error("获取乘客列表请求失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

		return errors.New("get passenger list failure")
	}

	type PassengerData struct {
		NormalPassengers common.PassengerInfos `json:"normal_passengers"`
		DJPassengers     common.PassengerInfos `json:"dj_passengers"`
	}

	type PassengerList struct {
		Data PassengerData `json:"data"`
	}

	result := PassengerList{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析乘客列表信息错误", zap.ByteString("res", body), zap.Error(err))

		return err
	}

	logger.Debug("联系人列表",
		zap.Array("passengers", result.Data.NormalPassengers),
	//  zap.ByteString("body", body),
	)

	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("联系人列表，若有重名联系人，请在配置中使用 UUID 作为乘车人:")
	for _, passenger := range result.Data.NormalPassengers {
		passengers[passenger.UUID] = passenger
		fmt.Println(passenger.String())
	}
	fmt.Println(strings.Repeat("-", 100))

	return
}

func GetPassenger(passengerName string) *common.PassengerInfo {
	for _, passengerInfo := range passengers {
		if passengerInfo.PassengerName == passengerName {
			return passengerInfo
		}
	}

	return nil
}

func GetPassengerByUUID(uuid string) *common.PassengerInfo {
	return passengers[uuid]
}
