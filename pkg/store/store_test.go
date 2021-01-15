package store

import (
	"context"
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

func TestStore_GetDemoEvents(t *testing.T) {
	events, err := s.GetDemoEvents(context.Background(), 2, 1, "%%")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, events)
		t.Log(events)
	}
}

func TestStore_Demo(t *testing.T) {
	s.Demo()
}

func TestStore_FindLine(t *testing.T) {
	got, err := s.FindLine(context.Background(), 462, "36150575")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}

func TestStore_DeleteLine(t *testing.T) {
	err := s.DeleteLine(context.Background(), 257848417)
	assert.NoError(t, err)
}
