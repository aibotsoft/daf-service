package handler

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"sync"
	"time"
)

//func (h *Handler) GetCurrency(sb *fortedpb.Surebet) float64 {
//	for _, currency := range sb.Currency {
//		if currency.Code == h.auth.Account.CurrencyCode {
//			return currency.Value
//		}
//	}
//	return 0
//}

func (h *Handler) GetCurrency(sb *pb.Surebet) float64 {
	for _, currency := range sb.Currency {
		if currency.Code == h.auth.Account.CurrencyCode {
			return currency.Value
		}
	}
	return 0
}

//const MinPeriod = 300 * time.Millisecond

type TimeoutLock struct {
	mu        sync.Mutex
	last      time.Time
	minPeriod time.Duration
}

func NewTimeoutLock(minPeriod time.Duration) *TimeoutLock {
	return &TimeoutLock{minPeriod: minPeriod}
}

func (l *TimeoutLock) Take() {
	l.mu.Lock()
	needSleep := l.minPeriod - time.Since(l.last)
	//log.Debugw("need sleep", "duration", needSleep, "since", time.Since(l.last))
	time.Sleep(needSleep)
}

func (l *TimeoutLock) Release() {
	l.last = time.Now()
	l.mu.Unlock()
}
