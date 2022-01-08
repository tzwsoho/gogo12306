package ticket

import (
	"encoding/json"
	"errors"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"path"
	"regexp"
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

func setHeaders(req *http.Request) {
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 Edg/96.0.1054.62"
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "www.12306.cn")
	req.URL.Host = "www.12306.cn"
	req.Host = "www.12306.cn"
}

func InitStations() (err error) {
	const (
		urlHomepage = "https://www.12306.cn/index/index.html"
	)
	req, _ := http.NewRequest("GET", urlHomepage, nil)
	setHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取主页错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取主页失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get homepage failure")
	}

	var re1, re2 *regexp.Regexp
	if re1, err = regexp.Compile(`\s{2,}<script src\s*=\s*"(.+?station_name.*\.js)">`); err != nil {
		logger.Error("生成正则表达式 1 失败", zap.Error(err))

		return
	}

	if re2, err = regexp.Compile(`\s{2,}<script src\s*=\s*"(.+?qss.*\.js)">`); err != nil {
		logger.Error("生成正则表达式 2 失败", zap.Error(err))

		return
	}

	body1 := re1.FindSubmatch(body)
	if len(body1) != 2 {
		logger.Error("匹配正则表达式 1 失败", zap.ByteString("body", body), zap.String("re", re1.String()))

		return errors.New("regexp 1 failure")
	}

	body2 := re2.FindSubmatch(body)
	if len(body2) != 2 {
		logger.Error("匹配正则表达式 2 失败", zap.ByteString("body", body), zap.String("re", re2.String()))

		return errors.New("regexp 2 failure")
	}

	urlStationName := path.Join("www.12306.cn/index", string(body1[1]))
	logger.Debug("获取站点列表 URL", zap.String("url", urlStationName))

	urlQSS := path.Join("www.12306.cn/index", string(body2[1]))
	logger.Debug("获取站点开售时间列表 URL", zap.String("url", urlQSS))

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 站点列表
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	req, _ = http.NewRequest("GET", "https://"+urlStationName, nil)
	setHeaders(req)
	req.Header.Set("Referer", urlHomepage)

	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点列表错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取站点列表失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get stations failure")
	}

	const (
		prefix1 = "var station_names ='@"
		suffix1 = "';"
	)
	stationList := string(body)
	if !strings.HasPrefix(stationList, prefix1) ||
		!strings.HasSuffix(stationList, suffix1) {
		logger.Error("获取到的不是站点列表", zap.ByteString("body", body))

		return errors.New("not station list")
	}
	stationList = strings.TrimSuffix(strings.TrimPrefix(stationList, prefix1), suffix1)

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

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 开售时间
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	req, _ = http.NewRequest("GET", "https://"+urlQSS, nil)
	setHeaders(req)
	req.Header.Set("Referer", urlHomepage)

	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点开售时间列表错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取站点开售时间列表失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get stations sale list failure")
	}

	const (
		prefix2 = "var citys = "
	)
	var saleList string
	if body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF { // 前 3 字节 UTF8 编码头部：EF BB BF
		saleList = string(body[3:])
	} else {
		saleList = string(body)
	}

	if !strings.HasPrefix(saleList, prefix2) {
		logger.Error("获取到的不是站点开售时间列表", zap.ByteString("body", body))

		return errors.New("not sale list")
	}
	saleList = strings.TrimPrefix(saleList, prefix2)

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
