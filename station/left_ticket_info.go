package station

import (
	"errors"
	"gogo12306/logger"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js
// 32 - 商务座
// 25 - 特等座
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
	SecretStr        string // 下单用的密钥
	CanOrder         bool   // 是否接受预订
	TrainCode        string // 车次
	Start            string // 始发站
	End              string // 终到站
	From             string // 出发站
	To               string // 到达站
	StartTime        string // 出发时间
	ArriveTime       string // 到达时间
	Duration         string // 历时
	ShangWuZuo       string // 商务座
	TeDengZuo        string // 特等座
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

func parseLeftTicketInfo(row string) (info *LeftTicketInfo, err error) {
	parts := strings.Split(row, "|")
	if len(parts) < 34 {
		return nil, errors.New("parts len errors")
	}

	info = &LeftTicketInfo{}
	if info.SecretStr, err = url.QueryUnescape(parts[0]); err != nil {
		return nil, err
	}

	info.CanOrder = (parts[1] == "预订")
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

	// 商务座
	if parts[32] == "" || parts[32] == "*" || parts[32] == "无" || parts[32] == "--" {
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

	// 特等座
	if parts[25] == "" || parts[25] == "*" || parts[25] == "无" || parts[25] == "--" {
		info.TeDengZuo = "0"
		info.LeftTicketsCount = append(info.LeftTicketsCount, 0)
	} else {
		info.TeDengZuo = parts[25]

		if parts[25] == "有" {
			info.LeftTicketsCount = append(info.LeftTicketsCount, 99)
		} else if LeftTicketsCount, e := strconv.ParseInt(parts[25], 10, 64); e != nil {
			logger.Error("转换数值错误", zap.Int("索引", 25), zap.Error(e))
			return nil, e
		} else {
			info.LeftTicketsCount = append(info.LeftTicketsCount, int(LeftTicketsCount))
		}
	}

	// 一等座
	if parts[31] == "" || parts[31] == "*" || parts[31] == "无" || parts[31] == "--" {
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

	// 二等座
	if parts[30] == "" || parts[30] == "*" || parts[30] == "无" || parts[30] == "--" {
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

	// 高级软卧
	if parts[21] == "" || parts[21] == "*" || parts[21] == "无" || parts[21] == "--" {
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

	// 软卧
	if parts[23] == "" || parts[23] == "*" || parts[23] == "无" || parts[23] == "--" {
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

	// 动卧
	if parts[33] == "" || parts[33] == "*" || parts[33] == "无" || parts[33] == "--" {
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

	// 硬卧
	if parts[28] == "" || parts[28] == "*" || parts[28] == "无" || parts[28] == "--" {
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

	// 软座
	if parts[24] == "" || parts[24] == "*" || parts[24] == "无" || parts[24] == "--" {
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

	// 硬座
	if parts[29] == "" || parts[29] == "*" || parts[29] == "无" || parts[29] == "--" {
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

	// 无座
	if parts[26] == "" || parts[26] == "*" || parts[26] == "无" || parts[26] == "--" {
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

	// 其他
	if parts[22] == "" || parts[22] == "*" || parts[22] == "无" || parts[22] == "--" {
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
