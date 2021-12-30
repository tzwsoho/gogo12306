package blacklist

import (
	"fmt"
	"sync"
	"time"
)

const BlackTime = time.Minute

// 小黑屋
// Key: 任务 ID + "_" + CDN IP + "_" + 车次 + "_" + 席位类型
// Value: 解除时间
var blackList sync.Map

func makeKey(taskID int64, cdnIP, trainCode, seatType string) string {
	return fmt.Sprintf("%d_%s_%s_%s", taskID, cdnIP, trainCode, seatType)
}

func AddToBlackList(taskID int64, cdnIP, trainCode, seatType string) {
	blackList.Store(makeKey(taskID, cdnIP, trainCode, seatType), time.Now().Add(BlackTime))
}

func DelFromBlackList(taskID int64, cdnIP, trainCode, seatType string) {
	blackList.Delete(makeKey(taskID, cdnIP, trainCode, seatType))
}

func IsInBlackList(taskID int64, cdnIP, trainCode, seatType string) bool {
	if t, exists := blackList.Load(makeKey(taskID, cdnIP, trainCode, seatType)); exists {
		if tt, ok := t.(time.Time); ok {
			if time.Now().Before(tt) {
				return true
			}
		}
	}

	return false
}
