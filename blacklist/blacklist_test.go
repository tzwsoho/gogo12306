package blacklist_test

import (
	"gogo12306/blacklist"
	"testing"
	"time"
)

func TestBlackList(t *testing.T) {
	const (
		TRAINCODE = "D3838"
		SEATINDEX = 3 // 座席索引
	)
	taskID := time.Now().UnixNano()

	blacklist.AddToBlackList(taskID, TRAINCODE, SEATINDEX, 60)

	// time.Sleep(time.Minute)

	if !blacklist.IsInBlackList(taskID, TRAINCODE, SEATINDEX) {
		t.Error("blacklist item not found")
		return
	}
}
