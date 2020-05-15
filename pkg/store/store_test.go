package store

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

var s *Store

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	s = NewStore(cfg, log, db)
	m.Run()
	s.Close()
}

func ticketsHelper(t *testing.T) []api.Ticket {
	t.Helper()
	return []api.Ticket{{
		Id:         0,
		BetTeam:    "",
		Price:      0,
		BetTypeId:  1,
		Points:     nil,
		EventId:    0,
		MarketName: "Handicap",
		IsLive:     false,
		MinBet:     0,
		MaxBet:     0,
		Home:       "",
		Away:       "",
	}, {
		Id:         1,
		BetTeam:    "",
		Price:      0,
		BetTypeId:  3,
		Points:     nil,
		EventId:    0,
		MarketName: "Over/Under",
		IsLive:     false,
		MinBet:     0,
		MaxBet:     0,
		Home:       "",
		Away:       "",
	}}

}
func TestStore_SaveMarkets(t *testing.T) {
	err := s.SaveMarkets(context.Background(), 1, "Handicap")
	assert.NoError(t, err)

}

func TestStore_SaveLines(t *testing.T) {
	err := s.SaveLines(context.Background(), ticketsHelper(t))
	assert.NoError(t, err)
}

func TestStore_FindLine(t *testing.T) {
	p := 1.750000
	ticket := &api.Ticket{
		BetTeam:   "a",
		BetTypeId: 3,
		Points:    &p,
		EventId:   36064260,
	}
	err := s.FindLine(context.Background(), ticket)
	if assert.NoError(t, err) {
		t.Log(ticket)
	}
}
