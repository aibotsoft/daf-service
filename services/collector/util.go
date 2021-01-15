package collector

import (
	api "github.com/aibotsoft/gen/dafapi"
	"time"
)

func (c *Collector) errLogAndSleep(err error, start time.Time) bool {
	var isErr bool
	if err != nil {
		c.log.Info(err)
		isErr = true
	}
	needSleep := collectMinPeriod - time.Since(start)
	//c.log.Debugw("need sleep", "duration", needSleep)
	time.Sleep(needSleep)
	return isErr
}
func (c *Collector) errLog(err error) {
	if err != nil {
		c.log.Info(err)
	}
}
func eventCont(dateType string, m0 api.SportItemM0) int64 {
	if dateType == "t" {
		return m0.T
	} else {
		return m0.E
	}
}
