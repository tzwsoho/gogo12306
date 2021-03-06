package common

import (
	"gogo12306/logger"
	"time"

	"go.uber.org/zap"
)

// CheckOperationPeriod 判断当前时间是否在 12306 开放的 6~23 点之间，有时不在这个时间段也可以做登录和购票操作
func CheckOperationPeriod() {
	// FOR TESTING
	return

	now := time.Now()
	if now.Hour() < 6 || now.Hour() >= 23 {
		_, zone := now.Zone()
		delta := now.Truncate(time.Hour*24).AddDate(0, 0, 1).Add(time.Hour*6 - time.Hour*time.Duration(zone/3600)).Sub(now)

		logger.Warn("不在 12306 营业时间内，程序将进入等待状态...",
			zap.String("再次运行时间", now.Add(delta).Format(time.RFC3339)),
		)

		time.Sleep(delta)
	}
}
