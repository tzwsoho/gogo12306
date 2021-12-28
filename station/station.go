package station

import (
	"encoding/json"
	"errors"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"sort"
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
		body []byte
		ok   bool
	)
	body, ok, _, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点列表错误", zap.Error(err))

		return
	} else if !ok {
		logger.Error("获取站点列表失败", zap.ByteString("body", body))

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
		body []byte
		ok   bool
	)
	body, ok, _, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取站点开售时间列表错误", zap.Error(err))

		return
	} else if !ok {
		logger.Error("获取站点开售时间列表失败", zap.ByteString("body", body))

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

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js
// 0 - 32 - 商务座/特等座
// 1 - 31 - 一等座
// 2 - 30 - 二等座/二等包座
// 3 - 21 - 高级软卧
// 4 - 23 - 软卧/一等卧
// 5 - 33 - 动卧
// 6 - 28 - 硬卧/二等卧
// 7 - 24 - 软座
// 8 - 29 - 硬座
// 9 - 26 - 无座
// 10 - 22 - 其他
func SeatNamesToSeatIndices(seatNames []string) (types, indices []int, err error) {
	for _, seatName := range seatNames {
		if seatName == "全部" {
			indices = append(indices, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
			types = append(types, 32, 31, 30, 21, 23, 33, 28, 24, 29, 26, 22)
			return
		}
	}

	for _, seatName := range seatNames {
		switch seatName {
		case "商务座", "特等座", "商务座/特等座":
			types = append(types, 32)
			indices = append(indices, 0)

		case "一等座":
			types = append(types, 31)
			indices = append(indices, 1)

		case "二等座", "二等包座", "二等座/二等包座":
			types = append(types, 30)
			indices = append(indices, 2)

		case "高级软卧":
			types = append(types, 21)
			indices = append(indices, 3)

		case "软卧", "一等卧", "软卧/一等卧":
			types = append(types, 23)
			indices = append(indices, 4)

		case "动卧":
			types = append(types, 33)
			indices = append(indices, 5)

		case "硬卧", "二等卧", "硬卧/二等卧":
			types = append(types, 28)
			indices = append(indices, 6)

		case "软座":
			types = append(types, 24)
			indices = append(indices, 7)

		case "硬座":
			types = append(types, 29)
			indices = append(indices, 8)

		case "无座":
			types = append(types, 26)
			indices = append(indices, 9)

		case "其他":
			types = append(types, 22)
			indices = append(indices, 10)

		default:
			return nil, nil, errors.New("unknown seat name")
		}
	}

	return
}

func ParseTask(taskCfg *config.TaskConfig) (task *worker.Task, err error) {
	task = &worker.Task{
		QueryOnly: taskCfg.QueryOnly,
	}

	from := StationNameToStationInfo(taskCfg.From)
	if from == nil {
		return nil, errors.New("from error")
	} else {
		task.From = taskCfg.From
		task.FromTelegramCode = from.TelegramCode
	}

	to := StationNameToStationInfo(taskCfg.To)
	if to == nil {
		return nil, errors.New("to error")
	} else {
		task.To = taskCfg.To
		task.ToTelegramCode = to.TelegramCode
	}

	if len(taskCfg.StartDates) <= 0 {
		return nil, errors.New("start_dates error")
	}

	if len(taskCfg.Seats) <= 0 {
		return nil, errors.New("seats error")
	}

	if len(taskCfg.Passengers) <= 0 {
		return nil, errors.New("passengers error")
	}

	// 计算开售时间
	for _, date := range taskCfg.StartDates {
		var saleTime time.Time
		if saleTime, err = time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("start_dates error")
		}

		// 需要减去（预售天数 - 1）
		saleTime = saleTime.Add(from.SaleTime - time.Hour*24*time.Duration(config.Cfg.OtherPresellDays-1))

		task.StartDates = append(task.StartDates, date)
		task.SaleTimes = append(task.SaleTimes, saleTime)
	}

	// 开售时间排序，从最快开售到最迟开售
	sort.Slice(task.SaleTimes, func(i, j int) bool {
		return task.SaleTimes[i].Before(task.SaleTimes[j])
	})

	// 车次
	for _, trainCode := range taskCfg.TrainCodes {
		task.TrainCodes = append(task.TrainCodes, strings.TrimSpace(strings.ToUpper(trainCode)))
	}

	// 座位
	task.Seats = append(task.Seats, taskCfg.Seats...)
	if task.SeatTypes, task.SeatIndices, err = SeatNamesToSeatIndices(taskCfg.Seats); err != nil {
		return nil, err
	}

	// 是否接受提交无座
	task.AllowNoSeat = taskCfg.AllowNoSeat

	// 乘客
	for _, passenger := range taskCfg.Passengers {
		task.Passengers = append(task.Passengers, strings.TrimSpace(passenger))
	}

	// 是否接受提交部分乘客
	task.AllowPartly = taskCfg.AllowPartly

	task.NextQueryTime = time.Now()
	task.CB = QueryLeftTicket
	return
}
