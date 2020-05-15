package auth

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	pb "github.com/aibotsoft/gen/confpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"go.uber.org/zap"
	"net/url"
	"time"
)

type Auth struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	store   *store.Store
	api     *api.Api
	Account pb.Account
	token   *api.Token
	base    *url.URL
	conf    *config_client.ConfClient
}

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, api *api.Api, conf *config_client.ConfClient) *Auth {
	a := &Auth{cfg: cfg, log: log, store: store,
		//Account: account,
		api: api, conf: conf}
	var err error
	a.Account, err = a.conf.GetAccount(context.Background(), "Dafabet")
	if err != nil {
		log.Panic(err)
	}
	err = a.Login(context.Background())
	if err != nil {
		log.Error(err)
	}
	//go a.AuthJob()
	return a
}
func (a *Auth) Base() *url.URL {
	if a.base == nil {
		a.BuildBase()
	}
	return a.base
}
func (a *Auth) Login(ctx context.Context) error {
	if a.token == nil {
		var err error
		a.token, err = a.store.LoadToken(ctx)
		if err != nil {
			a.log.Warn("load token from db error")
			//a.token, err = a.api.Login(ctx, a.Account.Username, a.Account.Password)
			//if err != nil {
			//	return errors.Wrap(err, "login error")
			//}
		}
	}
	if a.base != nil && time.Since(a.token.Last).Minutes() < 10 {
		a.log.Debugw("token fresh", "host", a.token.Host, "token", a.token.Token, "last", a.token.Last)
		return nil
	}

	a.log.Infow("begin check login", "host", a.token.Host, "token", a.token.Token, "id", a.token.Id)
	err := a.api.CheckLogin(ctx, a.Base())
	if err != nil {
		a.log.Info("token not active, need login..")
		//	a.token, err = a.api.Login(ctx, a.Account.Username, a.Account.Password)
		//	if err != nil {
		//		return errors.Wrap(err, "login error")
		//	}
		//	a.BuildBase()
		//} else {
		//	a.log.Debug("token was already active")
		//}
		//a.token.Last = time.Now()
		//oldId := a.token.Id
		//err = a.store.SaveToken(a.token)
		//if err != nil {
		//	return errors.Wrap(err, "save token error")
		//}
		//if oldId != a.token.Id {
		//	a.log.Info("new auth id")
	}
	return nil
}

func (a *Auth) BuildBase() {
	hostUrl, _ := url.Parse("http://" + a.token.Host)
	tokenPath, _ := url.Parse("(S(" + a.token.Token + "))/")
	a.base = hostUrl.ResolveReference(tokenPath)
	a.log.Infow("BuildBase done", "base", a.base)
}
