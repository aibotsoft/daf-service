package api

import (
	"context"
	"net/url"
	"regexp"
)

var balancePath, _ = url.Parse(`Balance.aspx`)
var balanceRe = regexp.MustCompile(`Available Funds\s*:\s*(.*)\s*<br>`)

func (a *Api) GetBalance(ctx context.Context, base *url.URL) (float64, error) {
	resp, err := a.client.R().EnableTrace().SetContext(ctx).Get(base.ResolveReference(balancePath).String())
	if err != nil {
		return 0, err
	}
	float, err := parseFloat(balanceRe, resp.String())
	if err != nil {
		return 0, err
	}
	return float, nil
}
