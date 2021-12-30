package station

import (
	"errors"
	"gogo12306/config"
	"gogo12306/login"
	"gogo12306/worker"
	"sort"
	"strings"
	"time"
)

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js
// 0 - 32 - 9 - 商务座
// 1 - 25 - TZ - 特等座
// 2 - 31 - M - 一等座
// 3 - 30 - O - 二等座/二等包座
// 4 - 21 - 6 - 高级软卧
// 5 - 23 - 4 - 软卧/一等卧
// 6 - 33 - F - 动卧
// 7 - 28 - 3 - 硬卧/二等卧
// 8 - 24 - 2 - 软座
// 9 - 29 - 1 - 硬座
// 10 - 26 - WZ - 无座
// 11 - 22 - 其他
// 全部: 硬座 -> 二等座 -> 硬卧 -> 一等座 -> 软座 -> 软卧 -> 特等座 -> 动卧 -> 高级软卧 -> 商务座 -> 无座 -> 其他
func seatNamesToSeatIndices(seatNames []string) (types, indices []int, err error) {
	for _, seatName := range seatNames {
		if seatName == "全部" {
			indices = append(indices, 9, 3, 7, 2, 8, 5, 1, 6, 4, 0, 10, 11)
			types = append(types, 29, 30, 28, 31, 24, 23, 25, 33, 21, 32, 26, 22)
			return
		}
	}

	for _, seatName := range seatNames {
		switch seatName {
		case "商务座":
			types = append(types, 32)
			indices = append(indices, 0)

		case "特等座":
			types = append(types, 25)
			indices = append(indices, 1)

		case "一等座":
			types = append(types, 31)
			indices = append(indices, 2)

		case "二等座", "二等包座", "二等座/二等包座":
			types = append(types, 30)
			indices = append(indices, 3)

		case "高级软卧":
			types = append(types, 21)
			indices = append(indices, 4)

		case "软卧", "一等卧", "软卧/一等卧":
			types = append(types, 23)
			indices = append(indices, 5)

		case "动卧":
			types = append(types, 33)
			indices = append(indices, 6)

		case "硬卧", "二等卧", "硬卧/二等卧":
			types = append(types, 28)
			indices = append(indices, 7)

		case "软座":
			types = append(types, 24)
			indices = append(indices, 8)

		case "硬座":
			types = append(types, 29)
			indices = append(indices, 9)

		case "无座":
			types = append(types, 26)
			indices = append(indices, 10)

		case "其他":
			types = append(types, 22)
			indices = append(indices, 11)

		default:
			return nil, nil, errors.New("unknown seat name")
		}
	}

	return
}

func ParseTask(taskCfg *config.TaskConfig) (task *worker.Task, err error) {
	task = &worker.Task{
		QueryOnly:      taskCfg.QueryOnly,
		Done:           make(chan struct{}, 1),
		OrderType:      taskCfg.OrderType,
		OrderCandidate: taskCfg.OrderCandidate,
		NextQueryTime:  time.Now(),
		CB:             QueryLeftTicket,
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

	if len(taskCfg.Passengers) <= 0 && len(taskCfg.UUIDs) <= 0 {
		return nil, errors.New("passengers/uuid error")
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
	if task.SeatTypes, task.SeatIndices, err = seatNamesToSeatIndices(taskCfg.Seats); err != nil {
		return nil, err
	}

	// 是否接受提交无座
	task.AllowNoSeat = taskCfg.AllowNoSeat

	// 乘客
	if !taskCfg.QueryOnly { // 只查询的任务将忽略乘客信息
		if len(taskCfg.Passengers) > 0 { // 使用乘客姓名做索引
			for _, passengerName := range taskCfg.Passengers {
				passengerName = strings.TrimSpace(passengerName)
				passenger := login.GetPassenger(passengerName)
				if passenger == nil {
					return nil, errors.New("passenger name not in passenger list")
				}

				task.Passengers = append(task.Passengers, passenger)
			}
		} else if len(taskCfg.UUIDs) > 0 { // 使用 UUID 做索引
			for _, uuid := range taskCfg.UUIDs {
				uuid = strings.TrimSpace(uuid)
				passenger := login.GetPassengerByUUID(uuid)
				if passenger == nil {
					return nil, errors.New("uuid not in passenger list")
				}

				task.Passengers = append(task.Passengers, passenger)
			}
		}
	}

	// 是否接受提交部分乘客
	task.AllowPartly = taskCfg.AllowPartly

	return
}
