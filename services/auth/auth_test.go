package auth

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

var a *Auth

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	conf := config_client.New(cfg, log)
	a = New(cfg, log, sto, conf)
	err := a.CheckAndLogin(context.Background())
	if err != nil {
		a.log.Info(err)
	}
	m.Run()
	sto.Close()
}

func TestAuth_Login(t *testing.T) {
	err := a.Login(context.Background())
	if assert.NoError(t, err) {
		assert.NotEmpty(t, a.session)
		assert.NotEmpty(t, a.token)
		t.Log(a.token)
		t.Log(a.session)
	}
}

func TestAuth_getAccount(t *testing.T) {
	_, err := a.getAccountUserName(context.Background())
	assert.NoError(t, err)
}

func TestAuth_setOddsType(t *testing.T) {
	err := a.setOddsType(context.Background())
	assert.NoError(t, err)
}
