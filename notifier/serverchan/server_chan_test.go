package serverchan_test

import (
	"gogo12306/config"
	"gogo12306/notifier/serverchan"
	"testing"
)

func TestBroadcast(t *testing.T) {
	if err := serverchan.Notify(&config.ServerChan{
		On:   true,
		SKey: "",
	}, "测试: 购票成功"); err != nil {
		t.Error(err.Error())
	}
}
