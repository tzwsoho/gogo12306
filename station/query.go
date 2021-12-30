package station

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/notifier"
	"gogo12306/order"
	"gogo12306/worker"

	"go.uber.org/zap"
)

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

// 0 - 9 - 商务座
// 1 - TZ - 特等座
// 2 - M - 一等座
// 3 - O - 二等座/二等包座
// 4 - 6 - 高级软卧
// 5 - 4 - 软卧/一等卧
// 6 - F - 动卧
// 7 - 3 - 硬卧/二等卧
// 8 - 2 - 软座
// 9 - 1 - 硬座
// 10 - WZ - 无座
// 11 - 其他
func seatIndexToSeatType(seatIndex int) string {
	var seatTypes []string = []string{"9", "TZ", "M", "O", "6", "4", "F", "3", "2", "1", "WZ", ""}
	if seatIndex < 0 || seatIndex >= len(seatTypes) {
		return ""
	}

	return seatTypes[seatIndex]
}

func seatIndexToSeatName(seatIndex int) string {
	var seatNames []string = []string{
		"商务座", "特等座", "一等座", "二等座/二等包座",
		"高级软卧", "软卧/一等卧", "动卧", "硬卧/二等卧",
		"软座", "硬座", "无座", "其他"}
	if seatIndex < 0 || seatIndex >= len(seatNames) {
		return strconv.Itoa(seatIndex)
	}

	return seatNames[seatIndex]
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
			logger.Info("未到开售时间，略过此日期...",
				zap.String("出发站", task.From),
				zap.String("到达站", task.To),
				zap.String("出发日期", startDate),
				zap.Time("开售时间", task.SaleTimes[i]),
			)

			continue
		}

		logger.Info("开始查询余票信息",
			zap.String("出发站", task.From),
			zap.String("到达站", task.To),
			zap.String("出发日期", startDate),
		)

		req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN(), leftTicketURL, startDate, task.FromTelegramCode, task.ToTelegramCode), nil)
		req.Header.Set("Referer", referer)
		httpcli.DefaultHeaders(req)

		var (
			body       []byte
			statusCode int
		)
		body, statusCode, err = httpcli.DoHttp(req, jar)
		if err != nil {
			logger.Error("获取余票查询 URL 错误", zap.Error(err))

			return
		} else if statusCode != http.StatusOK {
			logger.Error("获取余票查询 URL 失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

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

			return
		}

		// 仅查询
		if task.QueryOnly {
			fmt.Printf("出发站: %s, 到达站: %s, 出发日期: %s（标 * 车次为待购买车次）\n", task.From, task.To, startDate)
			fmt.Printf("%-5s%-8s%-6s%-8s%-6s%-7s%-8s%-8s%-6s%-6s%-6s%-6s%-7s%-7s%-7s%-7s%-7s%-7s%-7s%-7s\n",
				"车次", "出发站", "出发时间", "到达站", "到达时间", "历时", "始发站", "终到站",
				"商务座", "特等座", "一等座", "二等座", "高级软卧", "软卧", "动卧", "硬卧", "软座", "硬座", "无座", "其他",
			)
		}

		for _, row := range result.Data.Result {
			var leftTicketInfo *LeftTicketInfo
			if leftTicketInfo, err = parseLeftTicketInfo(row); err != nil || leftTicketInfo == nil {
				logger.Error("解析余票行信息错误", zap.String("行信息", row), zap.Error(err))

				continue
			}

			trainCode := strings.ToUpper(leftTicketInfo.TrainCode)

			// 仅查询
			if task.QueryOnly {

				// 筛选车次
				if inStringArray(trainCode, task.TrainCodes) {
					trainCode = "*" + trainCode
				} else {
					trainCode = " " + trainCode
				}

				// 每个汉字宽度约等于 2 个数字或字母，站点名最长五个汉字
				f := fmt.Sprintf("%%-7s%%-%ds%%-9s%%-%ds%%-9s%%-9s%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds\n",
					11-utf8.RuneCountInString(leftTicketInfo.From),
					11-utf8.RuneCountInString(leftTicketInfo.To),
					11-utf8.RuneCountInString(leftTicketInfo.Start),
					11-utf8.RuneCountInString(leftTicketInfo.End),
					9-countHan(leftTicketInfo.ShangWuZuo),
					9-countHan(leftTicketInfo.TeDengZuo),
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
					leftTicketInfo.TeDengZuo,
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

				logger.Debug("列车余票查询结果",
					zap.String("出发日期", startDate),
					zap.String("车次", leftTicketInfo.TrainCode),
					zap.String("始发站", leftTicketInfo.Start),
					zap.String("终到站", leftTicketInfo.End),
					zap.String("出发站", leftTicketInfo.From),
					zap.String("到达站", leftTicketInfo.To),
					zap.String("出发时间", leftTicketInfo.StartTime),
					zap.String("到达时间", leftTicketInfo.ArriveTime),
					zap.String("历时", leftTicketInfo.Duration),
					zap.String("商务座", leftTicketInfo.ShangWuZuo),
					zap.String("特等座", leftTicketInfo.TeDengZuo),
					zap.String("一等座", leftTicketInfo.YiDengZuo),
					zap.String("二等座/二等包座", leftTicketInfo.ErDengZuo),
					zap.String("高级软卧", leftTicketInfo.GaoJiRuanWo),
					zap.String("软卧/一等卧", leftTicketInfo.RuanWo),
					zap.String("动卧", leftTicketInfo.DongWo),
					zap.String("硬卧/二等卧", leftTicketInfo.YingWo),
					zap.String("软座", leftTicketInfo.RuanZuo),
					zap.String("硬座", leftTicketInfo.YingZuo),
					zap.String("无座", leftTicketInfo.WuZuo),
					zap.String("其他", leftTicketInfo.QiTa),
				)

				// 仅查询不下单
				continue
			}

			////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 以下为下单
			////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

			// 当前无法预订
			if !leftTicketInfo.CanOrder {
				continue
			}

			// 筛选车次
			if !inStringArray(trainCode, task.TrainCodes) {
				continue
			}

			// 筛选座位
			for _, seatIndex := range task.SeatIndices {
				var passengers worker.PassengerTicketInfos
				leftTickets := leftTicketInfo.LeftTicketsCount[seatIndex]
				if len(task.Passengers) <= leftTickets { // 剩余票数比乘客多，可以下单
					logger.Info("发现有足够的余票，准备尝试下单...",
						zap.String("车次", trainCode),
						zap.String("坐席类型", seatIndexToSeatName(seatIndex)),
						zap.String("出发站", leftTicketInfo.From),
						zap.String("到达站", leftTicketInfo.To),
						zap.String("出发时间", leftTicketInfo.StartTime),
						zap.String("到达时间", leftTicketInfo.ArriveTime),
						zap.Array("乘客", task.Passengers),
					)

					for _, passenger := range task.Passengers {
						passengers = append(passengers, &worker.PassengerTicketInfo{
							PassengerInfo: *passenger,
							SeatType:      seatIndexToSeatType(seatIndex),
							BedPos:        0,
						})
					}
				} else if task.AllowPartly { // 允许提交部分乘客
					somePassengers := task.Passengers[:leftTickets]

					logger.Info("乘车人数比余票数量多，只提交部分乘客...",
						zap.String("车次", trainCode),
						zap.String("坐席类型", seatIndexToSeatName(seatIndex)),
						zap.String("出发站", leftTicketInfo.From),
						zap.String("到达站", leftTicketInfo.To),
						zap.String("出发时间", leftTicketInfo.StartTime),
						zap.String("到达时间", leftTicketInfo.ArriveTime),
						zap.Array("乘客", somePassengers),
					)

					for _, passenger := range somePassengers {
						passengers = append(passengers, &worker.PassengerTicketInfo{
							PassengerInfo: *passenger,
							SeatType:      seatIndexToSeatType(seatIndex),
							BedPos:        0,
						})
					}
				} else {
					logger.Debug("乘车人数比余票数量多，忽略此车次和坐席...",
						zap.String("车次", trainCode),
						zap.String("坐席类型", seatIndexToSeatName(seatIndex)),
					)

					continue
				}

				// TODO 判断小黑屋

				// TODO 候补算法逻辑：
				// ①本车次无余票但可候补，马上进行候补
				// ②本车次无余票但可候补，不进行候补，待所有车次都查询完均无余票时，再遍历选择的车次并进行候补

				// TODO 有部分坐席类型不能进行候补，需要判断

				if task.OrderType == 1 { // 普通购票
					if err = order.SubmitOrder(jar, &order.SubmitOrderRequest{
						SecretStr:            leftTicketInfo.SecretStr,
						TrainDate:            startDate,
						QueryFromStationName: task.From, // 注意使用中文站名
						QueryToStationName:   task.To,   // 注意使用中文站名
					}); err != nil {
						// TODO 加入小黑屋

						continue
					}

					if err = order.InitToken(jar); err != nil {
						// TODO 加入小黑屋

						continue
					}

					var (
						ifShowPassCode        bool
						ifShowPassCodeTime    int
						passengerTicketStr    string = order.GetPassengerTickets(passengers)
						oldPassengerTicketStr string = order.GetOldPassengers(passengers)
					)
					if ifShowPassCode, ifShowPassCodeTime, err = order.CheckOrder(jar, &order.CheckOrderRequest{
						// Passengers: passengers,
						PassengerTicketStr:    passengerTicketStr,
						OldPassengerTicketStr: oldPassengerTicketStr,
					}); err != nil {
						// TODO 加入小黑屋

						continue
					}

					logger.Debug("是否需要验证码", zap.Bool("ifShowPassCode", ifShowPassCode), zap.Int("ifShowPassCodeTime", ifShowPassCodeTime))

					if err = order.GetQueueCountResult(jar, &order.GetQueueCountRequest{
						TrainDate:            startDate,
						TrainNumber:          leftTicketInfo.TrainNumber,
						TrainCode:            leftTicketInfo.TrainCode,
						SeatType:             seatIndexToSeatType(seatIndex),
						QueryFromStationName: task.FromTelegramCode,
						QueryToStationName:   task.ToTelegramCode,
						LeftTicketStr:        leftTicketInfo.LeftTicketStr,
					}); err != nil {
						// TODO 加入小黑屋

						continue
					}

					if ifShowPassCode {
						// TODO 验证码识别
					}

					if ifShowPassCodeTime > 0 {
						time.Sleep(time.Millisecond * time.Duration(ifShowPassCodeTime))
					}

					if err = order.ConfirmSingleForQueue(jar, &order.ConfirmSingleForQueueRequest{
						// Passengers: passengers,
						PassengerTicketStr:    passengerTicketStr,
						OldPassengerTicketStr: oldPassengerTicketStr,
					}); err != nil {
						// TODO 加入小黑屋

						continue
					}

					// 每三秒查询一次下单情况
					var orderID string
					for {
						if orderID, err = order.QueryOrderWaitTime(jar); err != nil || orderID != "" {
							break
						}

						time.Sleep(time.Second * 3)
					}

					if orderID == "" {
						// TODO 加入小黑屋

						continue
					}

					notifier.Broadcast(fmt.Sprintf("GOGO12306 已成功帮您抢到 %s 至 %s，出发时间 %s %s，车次 %s，乘客: %s 的车票，订单号为 %s，请尽快登陆 12306 网站完成购票支付",
						task.From, task.To, startDate, leftTicketInfo.StartTime, leftTicketInfo.TrainCode, passengers.Names(), orderID,
					))

					task.Done <- struct{}{}
					return
				} else if task.OrderType == 2 { // 自动捡漏下单
					var (
						passengerTicketStr    string = order.GetPassengerTicketsForAutoSubmit(passengers)
						oldPassengerTicketStr string = order.GetOldPassengersForAutoSubmit(passengers)
					)
					if err = order.AutoSubmitOrder(jar, &order.AutoSubmitOrderRequest{
						SecretStr:            leftTicketInfo.SecretStr,
						TrainDate:            startDate,
						QueryFromStationName: task.FromTelegramCode, // 注意使用电报码
						QueryToStationName:   task.ToTelegramCode,   // 注意使用电报码
						// Passengers: passengers,
						PassengerTicketStr:    passengerTicketStr,
						OldPassengerTicketStr: oldPassengerTicketStr,
					}); err != nil {
						logger.Error("自动下单失败", zap.Error(err))

						continue
					}
				} else if task.OrderCandidate { // 候补票
					// TODO
				} else { // 不接受候补
					logger.Debug("由于设置不接受候补，忽略此车次和坐席...",
						zap.String("车次", trainCode),
						zap.String("坐席类型", seatIndexToSeatName(seatIndex)),
					)

					continue
				}
			}
		}

		logger.Debug("查询列车余票",
			zap.String("出发站", task.From),
			zap.String("到达站", task.To),
			zap.String("出发日期", startDate),
			zap.Strings("result", result.Data.Result),
		)

		fmt.Println(strings.Repeat("-", 100))

		time.Sleep(time.Second)
	}

	return
}
