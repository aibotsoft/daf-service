package api

import (
	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"strings"
)

func parse(resp *resty.Response) (*html.Node, error) {
	body := resp.RawBody()
	defer func() {
		_ = body.Close()
	}()
	page, err := html.Parse(resp.RawBody())
	if err != nil {
		return nil, errors.Wrap(err, "parse body error")
	}
	return page, nil
}
func InnerTextTrimmed(node *html.Node) string {
	text := htmlquery.InnerText(node)
	return strings.TrimSpace(text)
}
func parseTeams(resp string) (string, string, error) {
	teamsFound := teamsRe.FindStringSubmatch(resp)
	if len(teamsFound) < 3 {
		return "", "", errors.Errorf("not found teams in resp: %v", resp)
	}
	home := strings.TrimSpace(teamsFound[1])
	away := strings.TrimSpace(teamsFound[2])
	return home, away, nil
}
