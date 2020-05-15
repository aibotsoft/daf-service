package api

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"time"
)

var balPath, _ = url.Parse(`Balance.aspx`)

func (a *Api) CheckLogin(ctx context.Context, base *url.URL) error {
	_, err := a.client.R().EnableTrace().SetContext(ctx).Get(base.ResolveReference(balPath).String())
	if err != nil {
		return errors.Wrap(err, "CheckLogin error")
	}
	return nil
}

type Token struct {
	Id    int
	Host  string
	Token string
	Last  time.Time
}
