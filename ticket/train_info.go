package ticket

import (
	"encoding/json"
	"errors"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type StationInfo struct {
	ID           int
	TelegramCode string        // 电报码
	StationName  string        // 站点名
	PinYin       string        // 拼音
	PY           string        // 拼音首字母
	PYCode       string        // 拼音码
	SaleTime     time.Duration // 站点开售时间
}

var stations map[int]*StationInfo

func init() {
	stations = make(map[int]*StationInfo)
}

func InitStations() (err error) {
	const (
		url = "https://kyfw.12306.cn/otn/resources/js/framework/station_name.js"
	)
	req, _ := http.NewRequest("GET", url, nil)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点列表错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取站点列表失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get stations failure")
	}

	const (
		prefix = "var station_names ='@"
		suffix = "';"
	)
	stationList := string(body)
	if !strings.HasPrefix(stationList, prefix) ||
		!strings.HasSuffix(stationList, suffix) {
		logger.Error("获取到的不是站点列表", zap.ByteString("body", body))

		return errors.New("not station list")
	}
	stationList = strings.TrimSuffix(strings.TrimPrefix(stationList, prefix), suffix)

	for _, row := range strings.Split(stationList, "@") {
		// 拼音码|站点名|电报码|拼音|拼音首字母|ID
		// gzh|广州|GZQ|guangzhou|gz|33
		// gzn|广州南|IZQ|guangzhounan|gzn|5
		fields := strings.Split(row, "|")

		var id int
		if id, err = strconv.Atoi(fields[5]); err != nil {
			logger.Error("解析站点信息错误", zap.Strings("fields", fields), zap.Error(err))

			continue
		}

		if stations[id] != nil {
			logger.Error("站点已存在", zap.Int("id", id), zap.Strings("fields", fields))

			continue
		}

		stations[id] = &StationInfo{
			ID:           id,
			TelegramCode: fields[2],
			StationName:  fields[1],
			PinYin:       fields[3],
			PY:           fields[4],
			PYCode:       fields[0],
		}
	}

	return
}

func StationNameToStationInfo(stationName string) (stationInfo *StationInfo) {
	for _, station := range stations {
		if station.StationName == stationName {
			return station
		}
	}

	return nil
}

func StationTelegramCodeToStationInfo(telegramCode string) (stationInfo *StationInfo) {
	for _, station := range stations {
		if station.TelegramCode == telegramCode {
			return station
		}
	}

	return nil
}

func InitSaleTime() (err error) {
	const (
		url = "https://kyfw.12306.cn/otn/resources/js/query/qss.js"
	)
	req, _ := http.NewRequest("GET", url, nil)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点开售时间列表错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取站点开售时间列表失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get sale list failure")
	}

	const (
		prefix = "var citys = "
	)
	saleList := string(body[3:]) // 前 3 字节 UTF8 编码头部：EF BB BF
	if !strings.HasPrefix(saleList, prefix) {
		logger.Error("获取到的不是站点开售时间列表", zap.ByteString("body", body))

		return errors.New("not sale list")
	}
	saleList = strings.TrimPrefix(saleList, prefix)

	saleMap := make(map[string]string)
	if err = json.Unmarshal([]byte(saleList), &saleMap); err != nil {
		logger.Error("解析站点开售时间列表错误", zap.ByteString("body", body))

		return
	}

	for stationName, saleTime := range saleMap {
		if stationInfo := StationNameToStationInfo(stationName); stationInfo == nil {
			// logger.Error("没有找到站点信息", zap.String("站点", stationName))

			continue
		} else {
			if stationInfo.SaleTime, err = time.ParseDuration(strings.ReplaceAll(saleTime, ":", "h") + "m"); err != nil {
				logger.Error("解析开售时间错误", zap.String("站点", stationName), zap.String("开售时间", saleTime))

				continue
			}
		}
	}

	return
}
