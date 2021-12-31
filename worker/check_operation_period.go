package worker

import (
	"gogo12306/logger"
	"time"

	"go.uber.org/zap"
)

// CheckOperationPeriod 判断当前时间是否在 12306 开放的 6~23 点之间，不在的话定时到明天 5 点 59 分开始
func CheckOperationPeriod() {
	return
	now := time.Now()
	if now.Hour() < 6 || now.Hour() >= 23 {
		_, zone := now.Zone()
		delta := now.Truncate(time.Hour*24).AddDate(0, 0, 1).Add(time.Hour*6 - time.Hour*time.Duration(zone/3600)).Sub(now)

		logger.Warn("不在 12306 营业时间内，程序将进入等待状态...",
			zap.String("再次运行时间", now.Add(delta).String()),
		)

		time.Sleep(delta)
	}
}
