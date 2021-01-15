package help

import (
	"go.uber.org/zap"
	"strings"
	"time"
)

func ClearName(name string) string {
	split := strings.Split(name, "|")
	return split[0]
}
func ConvertTimeZone(dtStr string) string {
	startsTime, err := time.Parse(time.RFC3339, dtStr+"-04:00")
	if err != nil {
		return dtStr
	}
	return startsTime.Format(time.RFC3339)
}
func ErrLog(err error, log *zap.SugaredLogger) {
	if err != nil {
		log.Info(err)
	}
}
