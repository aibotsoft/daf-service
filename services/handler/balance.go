package handler

import (
	"context"
	"github.com/aibotsoft/micro/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

const balanceMinPeriodSeconds = 40

type Balance struct {
	balance     float64
	outstanding float64
	last        time.Time
	mux         sync.RWMutex
	check       sync.Mutex
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

func (h *Handler) BalanceJob() {
	time.Sleep(time.Second * 10)
	h.GetBalance()
	for {
		h.GetBalance()
		time.Sleep(time.Minute)
	}
}
func (h *Handler) GetBalance() float64 {
	got, b := h.balance.Get()
	if !b {
		go h.CheckBalance(false)
	}
	return got
}
func (b *Balance) SetOutstanding(value float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.outstanding = value
}
func (b *Balance) FullBalance() float64 {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return util.TruncateFloat(b.balance+b.outstanding, 2)
}
func (b *Balance) CalcFillFactor() float64 {
	if b.FullBalance() == 0 {
		return 0
	}
	return util.TruncateFloat(b.outstanding/b.FullBalance(), 2)
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
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	ctxAuth, err := h.auth.Auth(ctx)
	if err != nil {
		return
	}
	start := time.Now()
	resp, _, err := h.client.BalanceApi.GetBalance(ctxAuth).Execute()
	if err != nil {
		h.log.Info(err)
		time.Sleep(5 * time.Second)
		return
	}
	//h.log.Infow("", "resp", resp.GetData())
	data := resp.GetData()
	balance, err := strconv.ParseFloat(strings.Replace(data.GetBCredit(), ",", "", -1), 64)
	if err != nil {
		h.log.Infow("parse_balance_error", "balance_str", data.GetBal())
	}
	outstanding, err := strconv.ParseFloat(strings.Replace(data.GetOutSd(), ",", "", -1), 64)
	if err != nil {
		h.log.Infow("parse_balance_error", "balance_str", data.GetBal())
	}
	//full := util.TruncateFloat(balance+outstanding, 2)
	//FillFactor := util.TruncateFloat(balance/full, 3)
	h.balance.Set(balance)
	h.balance.SetOutstanding(outstanding)

	h.log.Debugw("got_balance", "balance", balance, "outstanding", outstanding, "full", h.balance.FullBalance(), "fill_factor", h.balance.CalcFillFactor(), "time", time.Since(start))
}
