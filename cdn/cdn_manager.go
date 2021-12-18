package cdn

import (
	"bufio"
	"gogo12306/logger"
	"math/rand"
	"os"

	"go.uber.org/zap"
)

var cdns []string

func init() {
	cdns = make([]string, 0)
}

func LoadCDN(goodCDNPath string) (err error) {
	var fCDN *os.File
	if fCDN, err = os.Open(goodCDNPath); err != nil {
		logger.Error("打开可用 CDN 文件失败", zap.String("cdnPath", goodCDNPath), zap.Error(err))
		return
	}

	reader := bufio.NewReader(fCDN)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		ip := scanner.Text()
		cdns = append(cdns, ip)
	}

	fCDN.Close()

	logger.Debug("可用 CDN 数量", zap.Int("count", len(cdns)))
	return nil
}

func GetCDN() string {
	return cdns[rand.Intn(len(cdns)*10000)/10000]
}
