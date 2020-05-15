package handler

import (
	"context"
	"sync"
	"time"
)

const balanceMinPeriodSeconds = 40

type Balance struct {
	balance float64
	last    time.Time
	mux     sync.RWMutex
	check   sync.Mutex
}

func (b *Balance) CheckBegin() {
	b.check.Lock()
}
func (b *Balance) CheckDone() {
	b.check.Unlock()
}

func (b *Balance) Get() (float64, bool) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	var isFresh bool
	if time.Since(b.last).Seconds() < balanceMinPeriodSeconds {
		isFresh = true
	}
	return b.balance, isFresh
}

func (b *Balance) Set(bal float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.last = time.Now()
	b.balance = bal
}

func (b *Balance) Sub(risk float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.balance = b.balance - risk
}

func (h *Handler) GetBalance() float64 {
	got, b := h.balance.Get()
	if !b {
		go h.CheckBalance(false)
	}
	return got
}
func (h *Handler) CheckBalance(force bool) {
	//h.log.Info("got check balance run")
	h.balance.CheckBegin()
	defer h.balance.CheckDone()
	//h.log.Info("begin check balance")
	_, b := h.balance.Get()
	if b && !force {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	h.log.Info("send balance request")

	balance, err := h.api.GetBalance(ctx, h.auth.Base())
	if err != nil {
		h.log.Error(err)
		return
	}
	//h.log.Infow("got balance response", "b", balance)
	h.balance.Set(balance)
}
