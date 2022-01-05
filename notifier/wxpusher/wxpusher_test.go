package wxpusher_test

import (
	"gogo12306/config"
	"gogo12306/notifier/wxpusher"
	"testing"
)

func TestBroadcast(t *testing.T) {
	if err := wxpusher.Notify(&config.WXPusher{
		On:       true,
		AppToken: "",
		TopicIDs: []int64{},
		UIDs:     []string{},
	}, "测试: 购票成功"); err != nil {
		t.Error(err.Error())
	}
}
