package station

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"go.uber.org/zap"
)

var leftTicketURL string

func InitLeftTickerURL() (err error) {
	const (
		url     = "https://%s/otn/leftTicket/init"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN()), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	body, ok, _, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取余票查询 URL 错误", zap.Error(err))

		return
	} else if !ok {
		logger.Error("获取余票查询 URL 失败", zap.ByteString("body", body))

		return errors.New("get left ticket url failure")
	}

	var re *regexp.Regexp
	if re, err = regexp.Compile("var CLeftTicketUrl = '([^']+?)'"); err != nil {
		logger.Error("生成正则表达式失败", zap.Error(err))

		return
	}

	body2 := re.FindSubmatch(body)
	if body2 == nil || len(body2) != 2 {
		logger.Error("匹配正则表达式失败", zap.ByteString("body", body), zap.String("re", re.String()))

		return errors.New("regexp failure")
	}

	leftTicketURL = string(body2[1])
	return
}

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js
// 32 - 商务座/特等座
// 31 - 一等座
// 30 - 二等座/二等包座
// 21 - 高级软卧
// 23 - 软卧/一等卧
// 33 - 动卧
// 28 - 硬卧/二等卧
// 24 - 软座
// 29 - 硬座
// 26 - 无座
// 22 - 其他

type LeftTicketInfo struct {
	TrainCode        string // 车次
	Start            string // 始发站
	End              string // 终到站
	From             string // 出发站
	To               string // 到达站
	StartTime        string // 出发时间
	ArriveTime       string // 到达时间
	Duration         string // 历时
	ShangWuZuo       string // 商务座/特等座
	YiDengZuo        string // 一等座
	ErDengZuo        string // 二等座/二等包座
	GaoJiRuanWo      string // 高级软卧
	RuanWo           string // 软卧/一等卧
	DongWo           string // 动卧
	YingWo           string // 硬卧/二等卧
	RuanZuo          string // 软座
	YingZuo          string // 硬座
	WuZuo            string // 无座
	QiTa             string // 其他
	LeftTicketsCount []int  // 各类型座位的剩余票数
}

func ParseLeftTicketInfo(row string) (info *LeftTicketInfo, err error) {
	parts := strings.Split(row, "|")
	if len(parts) < 34 {
		return nil, errors.New("parts len errors")
	}

	info = &LeftTicketInfo{}
	info.TrainCode = parts[3]

	start := StationTelegramCodeToStationInfo(parts[4])
	if start == nil {
		return nil, errors.New("start station error")
	}
	info.Start = start.StationName

	end := StationTelegramCodeToStationInfo(parts[5])
	if end == nil {
		return nil, errors.New("end station error")
	}
	info.End = end.StationName

	from := StationTelegramCodeToStationInfo(parts[6])
	if from == nil {
		return nil, errors.New("from station error")
	}
	info.From = from.StationName

	to := StationTelegramCodeToStationInfo(parts[7])
	if to == nil {
		return nil, errors.New("to station error")
	}
	info.To = to.StationName

	info.StartTime = parts[8]
	info.ArriveTime = parts[9]
	info.Duration = parts[10]

	if parts[32] == "" || parts[32] == "无" || parts[32] == "--" {
		info.ShangWuZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.ShangWuZuo = parts[32]

		if parts[32] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[32], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 32), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[31] == "" || parts[31] == "无" || parts[31] == "--" {
		info.YiDengZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.YiDengZuo = parts[31]

		if parts[31] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[31], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 31), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[30] == "" || parts[30] == "无" || parts[30] == "--" {
		info.ErDengZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.ErDengZuo = parts[30]

		if parts[30] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[30], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 30), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[21] == "" || parts[21] == "无" || parts[21] == "--" {
		info.GaoJiRuanWo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.GaoJiRuanWo = parts[21]

		if parts[21] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[21], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 21), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[23] == "" || parts[23] == "无" || parts[23] == "--" {
		info.RuanWo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.RuanWo = parts[23]

		if parts[23] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[23], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 23), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[33] == "" || parts[33] == "无" || parts[33] == "--" {
		info.DongWo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.DongWo = parts[33]

		if parts[33] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[33], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 33), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[28] == "" || parts[28] == "无" || parts[28] == "--" {
		info.YingWo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.YingWo = parts[28]

		if parts[28] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[28], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 28), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[24] == "" || parts[24] == "无" || parts[24] == "--" {
		info.RuanZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.RuanZuo = parts[24]

		if parts[24] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[24], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 24), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[29] == "" || parts[29] == "无" || parts[29] == "--" {
		info.YingZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.YingZuo = parts[29]

		if parts[29] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[29], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 29), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[26] == "" || parts[26] == "无" || parts[26] == "--" {
		info.WuZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.WuZuo = parts[26]

		if parts[26] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[26], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 26), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	if parts[22] == "" || parts[22] == "无" || parts[22] == "--" {
		info.QiTa = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.QiTa = parts[22]

		if parts[22] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[22], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 22), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	return
}

func inStringArray(s string, arr []string) bool {
	for _, ss := range arr {
		if s == ss {
			return true
		}
	}

	return false
}

func countHan(s string) (count int) {
	for _, c := range s {
		if unicode.Is(unicode.Han, c) {
			count++
		}
	}

	return
}

func QueryLeftTicket(jar *cookiejar.Jar, task *worker.Task) (err error) {
	if len(task.StartDates) != len(task.SaleTimes) {
		return errors.New("len of start_dates/saletimes not match")
	}

	if leftTicketURL == "" {
		leftTicketURL = "leftTicket/query"
	}

	const (
		url     = "https://%s/otn/%s?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	now := time.Now()
	for i, startDate := range task.StartDates {
		if i >= len(task.SaleTimes) || now.Before(task.SaleTimes[i]) {
			logger.Debug("未到开售时间，略过此日期...",
				zap.String("出发站", task.From),
				zap.String("到达站", task.To),
				zap.String("出发日期", startDate),
				zap.Time("开售时间", task.SaleTimes[i]),
			)

			continue
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN(), leftTicketURL, startDate, task.FromTelegramCode, task.ToTelegramCode), nil)
		req.Header.Set("Referer", referer)
		httpcli.DefaultHeaders(req)

		var (
			body []byte
			ok   bool
		)
		body, ok, _, err = httpcli.DoHttp(req, jar)
		if err != nil {
			logger.Error("获取余票查询 URL 错误", zap.Error(err))

			return
		} else if !ok {
			logger.Error("获取余票查询 URL 失败", zap.ByteString("body", body))

			return errors.New("get left ticket url failure")
		}

		type LeftTicketData struct {
			Result []string `json:"result"`
		}

		type LeftTicketResult struct {
			Data LeftTicketData `json:"data"`
		}
		result := LeftTicketResult{}
		if err = json.Unmarshal(body, &result); err != nil {
			logger.Error("解析余票信息错误", zap.ByteString("res", body), zap.Error(err))

			return err
		}

		fmt.Printf("出发站: %s, 到达站: %s, 出发日期: %s（标 * 车次为待购买车次）\n", task.From, task.To, startDate)
		fmt.Printf("%-5s%-8s%-6s%-8s%-6s%-7s%-8s%-8s%-6s%-6s%-6s%-7s%-7s%-7s%-7s%-7s%-7s%-7s%-7s\n",
			"车次", "出发站", "出发时间", "到达站", "到达时间", "历时", "始发站", "终到站",
			"商务座", "一等座", "二等座", "高级软卧", "软卧", "动卧", "硬卧", "软座", "硬座", "无座", "其他",
		)

		for _, row := range result.Data.Result {
			var leftTicketInfo *LeftTicketInfo
			if leftTicketInfo, err = ParseLeftTicketInfo(row); err != nil || leftTicketInfo == nil {
				logger.Error("解析余票行信息错误", zap.String("行信息", row), zap.Error(err))

				continue
			}

			// 筛选车次
			trainCode := strings.ToUpper(leftTicketInfo.TrainCode)
			if inStringArray(trainCode, task.TrainCodes) {
				trainCode = "*" + trainCode
			} else {
				trainCode = " " + trainCode
			}

			// 每个汉字宽度约等于 2 个数字或字母，站点名最长五个汉字
			f := fmt.Sprintf("%%-7s%%-%ds%%-9s%%-%ds%%-9s%%-9s%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds\n",
				11-utf8.RuneCountInString(leftTicketInfo.From),
				11-utf8.RuneCountInString(leftTicketInfo.To),
				11-utf8.RuneCountInString(leftTicketInfo.Start),
				11-utf8.RuneCountInString(leftTicketInfo.End),
				9-countHan(leftTicketInfo.ShangWuZuo),
				9-countHan(leftTicketInfo.YiDengZuo),
				9-countHan(leftTicketInfo.ErDengZuo),
				9-countHan(leftTicketInfo.GaoJiRuanWo),
				9-countHan(leftTicketInfo.RuanWo),
				9-countHan(leftTicketInfo.DongWo),
				9-countHan(leftTicketInfo.YingWo),
				9-countHan(leftTicketInfo.RuanZuo),
				9-countHan(leftTicketInfo.YingZuo),
				9-countHan(leftTicketInfo.WuZuo),
				9-countHan(leftTicketInfo.QiTa),
			)
			fmt.Printf(f,
				trainCode,
				leftTicketInfo.From,
				leftTicketInfo.StartTime,
				leftTicketInfo.To,
				leftTicketInfo.ArriveTime,
				leftTicketInfo.Duration,
				leftTicketInfo.Start,
				leftTicketInfo.End,
				leftTicketInfo.ShangWuZuo,
				leftTicketInfo.YiDengZuo,
				leftTicketInfo.ErDengZuo,
				leftTicketInfo.GaoJiRuanWo,
				leftTicketInfo.RuanWo,
				leftTicketInfo.DongWo,
				leftTicketInfo.YingWo,
				leftTicketInfo.RuanZuo,
				leftTicketInfo.YingZuo,
				leftTicketInfo.WuZuo,
				leftTicketInfo.QiTa,
			)

			// logger.Debug("列车余票查询结果",
			// 	zap.String("出发日期", trainDate),
			// 	zap.String("车次", leftTicketInfo.TrainCode),
			// 	zap.String("始发站", leftTicketInfo.Start),
			// 	zap.String("终到站", leftTicketInfo.End),
			// 	zap.String("出发站", leftTicketInfo.From),
			// 	zap.String("到达站", leftTicketInfo.To),
			// 	zap.String("出发时间", leftTicketInfo.StartTime),
			// 	zap.String("到达时间", leftTicketInfo.ArriveTime),
			// 	zap.String("历时", leftTicketInfo.Duration),
			// 	zap.String("商务座/特等座", leftTicketInfo.ShangWuZuo),
			// 	zap.String("一等座", leftTicketInfo.YiDengZuo),
			// 	zap.String("二等座/二等包座", leftTicketInfo.ErDengZuo),
			// 	zap.String("高级软卧", leftTicketInfo.GaoJiRuanWo),
			// 	zap.String("软卧/一等卧", leftTicketInfo.RuanWo),
			// 	zap.String("动卧", leftTicketInfo.DongWo),
			// 	zap.String("硬卧/二等卧", leftTicketInfo.YingWo),
			// 	zap.String("软座", leftTicketInfo.RuanZuo),
			// 	zap.String("硬座", leftTicketInfo.YingZuo),
			// 	zap.String("无座", leftTicketInfo.WuZuo),
			// 	zap.String("其他", leftTicketInfo.QiTa),
			// )

			// 仅查询
			if task.QueryOnly {
				continue
			}

			// 筛选车次
			if !inStringArray(leftTicketInfo.TrainCode, task.TrainCodes) {
				continue
			}

			// 筛选座位
			for _, seatIndex := range task.SeatIndices {
				leftTickets := leftTicketInfo.LeftTicketsCount[seatIndex]
				if len(task.Passengers) <= leftTickets { // 剩余票数比乘客多，可以下单
					// TODO 下单
				} else if task.AllowPartly { // 允许提交部分乘客
					passengers := task.Passengers[:leftTickets]
					_ = passengers

					// TODO 下单
				}
			}
		}

		// logger.Debug("查询列车余票",
		// 	zap.String("出发站", task.From),
		// 	zap.String("到达站", task.To),
		// 	zap.String("出发日期", trainDate),
		// 	zap.Strings("result", result.Data.Result),
		// )

		fmt.Println(strings.Repeat("-", 100))

		time.Sleep(time.Second)
	}

	return
}
