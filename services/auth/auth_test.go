package auth

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/gen/confpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/url"
	"testing"
)

var a *Auth

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	serviceApi := api.New(cfg, log)
	conf := config_client.New(cfg, log)
	a = New(cfg, log, sto, serviceApi, conf)
	m.Run()
	sto.Close()
}

func TestAuth_Login(t *testing.T) {
	err := a.Login(context.Background())
	assert.NoError(t, err)
}

func TestAuth_Login1(t *testing.T) {
	err := a.Login(context.Background())
	assert.NoError(t, err)
}

func TestAuth_Login2(t *testing.T) {
	type fields struct {
		cfg     *config.Config
		log     *zap.SugaredLogger
		store   *store.Store
		api     *api.Api
		Account confpb.Account
		token   *api.Token
		base    *url.URL
		conf    *config_client.ConfClient
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				cfg:     tt.fields.cfg,
				log:     tt.fields.log,
				store:   tt.fields.store,
				api:     tt.fields.api,
				Account: tt.fields.Account,
				token:   tt.fields.token,
				base:    tt.fields.base,
				conf:    tt.fields.conf,
			}
			if err := a.Login(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
