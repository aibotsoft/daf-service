package auth

import (
	"context"
	api "github.com/aibotsoft/gen/dafapi"
	"github.com/pkg/errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/aibotsoft/daf-service/pkg/store"
	pb "github.com/aibotsoft/gen/confpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"go.uber.org/zap"
)

type Auth struct {
	cfg           *config.Config
	log           *zap.SugaredLogger
	store         *store.Store
	client        *api.APIClient
	Account       pb.Account
	conf          *config_client.ConfClient
	jar           *cookiejar.Jar
	session       string
	token         string
	url           *url.URL
	bettingStatus bool
	userNameEqual bool
}

const (
	Ok                = 0
	sessionTerminated = 100
	badBettingStatus  = 101
	requestError      = 999
)

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, conf *config_client.ConfClient) *Auth {
	clientConfig := api.NewConfiguration()
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://ismart.dafabet.com")
	tr := &http.Transport{TLSHandshakeTimeout: 0 * time.Second, IdleConnTimeout: 0 * time.Second}
	clientConfig.HTTPClient = &http.Client{Jar: jar, Transport: tr}
	clientConfig.Debug = cfg.Service.Debug
	client := api.NewAPIClient(clientConfig)
	a := &Auth{cfg: cfg, log: log, store: store, client: client, conf: conf, jar: jar, url: u}
	return a
}
func (a *Auth) Close() {
	apiConfig := a.client.GetConfig()
	apiConfig.HTTPClient.CloseIdleConnections()
}
func (a *Auth) CheckAndLogin(ctx context.Context) error {
	ctxAuth, _ := a.Auth(ctx)
	code, err := a.checkToken(ctxAuth)
	if err != nil {
		if code == sessionTerminated {
			a.log.Info("session_terminated")
			err := a.Login(ctx)
			if err != nil {
				return err
			}
		} else if code == badBettingStatus {
			a.log.Info("bad_bet_status")
			a.bettingStatus = false
			return err
		} else {
			a.log.Info("other_check_token_error: ", err)
			return err
		}
	}
	if !a.userNameEqual {
		a.log.Info("begin check user name")
		userName, err := a.getAccountUserName(ctx)
		if err != nil {
			return err
		}
		if strings.ToLower(userName) != strings.ToLower(a.Account.Username) {
			a.log.Info("user name not equal, begin login")
			err := a.Login(ctx)
			if err != nil {
				return err
			}
		}
		a.userNameEqual = true
		a.log.Info("userName equal: ", userName)
	}
	return nil
}
func (a *Auth) AuthJob() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		err := a.CheckAndLogin(ctx)
		if err != nil {
			a.log.Error(err)
		}
		cancel()
		time.Sleep(3 * time.Minute)
	}
}

var BettingStatusError = errors.New("false_betting_status")

func (a *Auth) Auth(ctx context.Context) (auth context.Context, err error) {
	if a.Account.Username == "" {
		a.Account, err = a.conf.GetAccount(ctx, a.cfg.Service.Name)
		if err != nil {
			a.log.Info(err)
			return nil, BettingStatusError
		}
		a.log.Infow("got_account", "acc", a.Account)
	}
	if a.session == "" {
		a.session, a.token, err = a.store.LoadToken(ctx)
		if err != nil {
			a.log.Info(err)
			return nil, BettingStatusError
		}
	}
	auth = context.WithValue(ctx, api.ContextAPIKeys, map[string]api.APIKey{"x-session": {Key: "ASP.NET_SessionId=" + a.session}})
	if !a.bettingStatus {
		return auth, BettingStatusError
	}
	return auth, nil
}

func (a *Auth) checkToken(ctxAuth context.Context) (int64, error) {
	start := time.Now()
	resp, _, err := a.client.LoginApi.CheckToken(ctxAuth).Execute()
	if err != nil {
		a.log.Infow("check_token_not_ok", "resp", resp)
		return requestError, err
	}
	if resp.ErrorCode != nil {
		a.log.Infow("error_code_not_ok", "resp", resp)
		return resp.GetErrorCode(), errors.Errorf("session_error, code: %v, msg: %v ", resp.GetErrorCode(), resp.GetErrorMsg())
	}

	a.bettingStatus = true
	a.log.Infow("token_fresh", "time", time.Since(start))
	return 0, nil
}

func (a *Auth) getAccountUserName(ctx context.Context) (string, error) {
	ctxAuth, err := a.Auth(ctx)
	if err != nil {
		return "", err
	}
	resp, _, err := a.client.LoginApi.GetAccount(ctxAuth).Execute()
	if err != nil {
		return "", err
	}
	//a.log.Infow("", "", resp.Data.GetLoginUserName())
	return resp.Data.GetLoginUserName(), nil
}
func (a *Auth) Login(ctx context.Context) error {
	a.log.Info("begin login")
	err := a.authenticate(ctx)
	if err != nil {
		return err
	}
	err = a.processLogin(ctx)
	if err != nil {
		return err
	}
	a.log.Info("begin save session")
	err = a.store.SaveSession(a.session, a.token)
	if err != nil {
		return err
	}
	return nil
}
func (a *Auth) processLogin(ctx context.Context) error {
	_, i, err := a.client.LoginApi.ProcessLogin(ctx).Lang("en").St(a.token).HomeURL("https://m.dafabet.com/en").
		ExtendSessionURL("https://m.dafabet.com/en&OType=01&oddstype=1").Execute()
	if err != nil {
		a.log.Info(err)
		a.log.Info(i)
		return nil
	}
	a.session = ""
	for _, c := range a.jar.Cookies(a.url) {
		if c.Name == "ASP.NET_SessionId" {
			a.session = c.Value
			break
		}
	}
	if a.session == "" {
		return errors.Errorf("not found session in cookies")
	}
	return nil
}

//func (a *Auth) setOddsType(ctx context.Context) error {
//	ctxAuth, err := a.Auth(ctx)
//	if err != nil {
//		return err
//	}
//	resp, _, err := a.client.LoginApi.SetOddsType(ctxAuth).Set(1).Execute()
//	if err != nil {
//		return err
//	}
//	a.log.Infow("", "", resp)
//	return nil
//}

func (a *Auth) authenticate(ctx context.Context) error {
	resp, r, err := a.client.LoginApi.Authenticate(ctx).Username(a.Account.Username).Password(a.Account.Password).Execute()
	if err != nil {
		a.log.Info(err)
		a.log.Info(r)
		return err
	}
	a.token = resp.GetToken()
	return nil
}
