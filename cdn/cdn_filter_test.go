package cdn_test

import (
	"gogo12306/cdn"
	"gogo12306/logger"
	"testing"
)

func TestCDNFilter(t *testing.T) {
	logger.Init(true, "test.go", "info", 1024, 7)

	cdn.FilterCDN("../cdn.txt", "../good_cdn.txt")
}
