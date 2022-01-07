package order

import (
	"errors"
	"fmt"
	"net/http/cookiejar"
	"time"

	"gogo12306/common"
	"gogo12306/logger"
	"gogo12306/notifier"
	"gogo12306/order/auto"
	"gogo12306/order/candidate"
	orderCommon "gogo12306/order/common"
	"gogo12306/order/normal"
	"gogo12306/worker"

	"go.uber.org/zap"
)

func DoOrder(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *common.LeftTicketInfo,
	startDate, trainCode string, seatIndex int, passengers common.PassengerTicketInfos) (err error) {
	// if err = login.CheckAndRelogin(jar); err != nil {
	// 	return
	// }

	// 候补算法可以有以下逻辑，目前实现了第①种：
	// ①本车次无余票但可候补，马上进行候补
	// ②本车次无余票但可候补，不进行候补，待所有车次都查询完均无余票时，再遍历选择的车次并进行候补

	// 正规的候补规则异常复杂，目前只实现了第⑤种情况，应该足够使用：
	// ①开车时间在当前时间 6 小时以内不能候补
	// ②超过开售时间的列车不能候补
	// ③余票查询时 secretStr 为空的列车不能候补
	// ④返程(fc)列车票不能候补，改签(gc)列车票不能候补
	// ⑤余票查询结果第 11 列(canWebBuy)是 Y，或第 37 列(houbu_train_flag)不是 1 不能候补

	var orderID string
	if !leftTicketInfo.CanWebBuy && leftTicketInfo.CandidateFlag { // 可以候补
		if task.AllowCandidate { // 抢候补票
			if err = candidate.DoCandidate(jar, task, leftTicketInfo, seatIndex, passengers); err != nil {
				return
			}

			// TODO 候补成功消息广播

			// 候补完成后继续尝试抢其他车次的票
			return
		} else { // 不接受候补
			logger.Debug("由于设置不接受候补，忽略此车次和座席...",
				zap.String("车次", trainCode),
				zap.String("座席类型", orderCommon.SeatIndexToSeatName(seatIndex)),
			)

			return errors.New("candidate not allow")
		}
	} else { // 直接购票
		if task.OrderType == 1 { // 普通购票，流程复杂，耗时较长，但成功率高
			if orderID, err = normal.DoNormalOrder(jar, task, leftTicketInfo, startDate, seatIndex, passengers); err != nil {
				return
			}
		} else if task.OrderType == 2 { // 自动捡漏下单，流程简单，但成功率不高不稳定
			if orderID, err = auto.DoAutoOrder(jar, task, leftTicketInfo, startDate, passengers); err != nil {
				return
			}
		}
	}

	notifier.Broadcast(fmt.Sprintf("GOGO12306 于 %s 成功帮您抢到 %s 至 %s，出发时间 %s %s，车次 %s，乘客: %s 的车票，订单号为 %s，请尽快登陆 12306 网站完成购票支付",
		time.Now().Format(time.RFC3339), task.From, task.To, startDate, leftTicketInfo.StartTime, leftTicketInfo.TrainCode, passengers.Names(), orderID,
	))

	task.Done <- struct{}{}
	return
}
