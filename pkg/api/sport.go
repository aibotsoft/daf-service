package api

import (
	"context"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var sportPath, _ = url.Parse(`Sports.aspx`)

const sportStopList = "Virtual Sports Finance"

type Sport struct {
	Id         int64
	Name       string
	Count      int
	EventState string
}

func (a *Api) GetSports(ctx context.Context, base *url.URL, eventState string) ([]Sport, error) {
	var sports []Sport
	resp, err := a.client.R().EnableTrace().SetContext(ctx).SetDoNotParseResponse(true).
		SetQueryParam("eventState", eventState).
		Get(base.ResolveReference(sportPath).String())
	if err != nil {
		return nil, errors.Wrapf(err, "get sports error in page %v", eventState)
	}
	page, err := parse(resp)
	if err != nil {
		return nil, err
	}
	linkNodes, err := htmlquery.QueryAll(page, `//a`)
	if err != nil {
		return nil, errors.Wrap(err, "query //a in sport page error")
	}
	for _, link := range linkNodes {
		sportId, err := IdFromNode(link)
		if err != nil {
			continue
		}
		name, count, err := NameAndCount(htmlquery.InnerText(link))
		if err != nil {
			continue
		}
		sports = append(sports, Sport{Id: sportId, Name: strings.TrimSpace(name), Count: count, EventState: strings.TrimSpace(eventState)})
	}
	return sports, nil
}
func IdFromNode(node *html.Node) (int64, error) {
	href := htmlquery.SelectAttr(node, "href")
	if len(href) < 2 {
		return 0, errors.Errorf("not found id in link")
	}
	params := strings.Split(href, "=")
	if len(params) < 3 {
		return 0, errors.Errorf("not found id in link")
	}
	sportId := params[2]
	id, err := strconv.ParseInt(sportId, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "find id in link error")
	}
	return id, nil
}

var nameCountRe = regexp.MustCompile(`(.+)\s\((\d{1,5})\)`)

func NameAndCount(str string) (string, int, error) {
	match := nameCountRe.FindStringSubmatch(str)
	if len(match) < 2 {
		return "", 0, errors.Errorf("not fount name and count in link text %s", str)
	}
	count, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, errors.Wrap(err, "convert count to int error")
	}
	return match[1], count, nil

}
