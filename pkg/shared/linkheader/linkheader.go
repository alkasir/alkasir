package linkheader

import (
	"net/http"
	"net/url"
)

func NewLinkHeader(u *url.URL, pageParam string) *LinkHeader {
	urlstring := u.String()
	urlcopy, err := url.Parse(urlstring)
	if err != nil {
		panic(err)
	}

	return &LinkHeader{
		pageParam: pageParam,
		links:     make([]Link, 0),
		url:       urlcopy,
	}
}

// LinkHeader .
type LinkHeader struct {
	pageParam string
	links     []Link
	url       *url.URL
}

func (l *LinkHeader) First(pageValue string) {
	l.links = append(l.links, Link{
		URI: l.pageString(pageValue),
		Rel: "first",
	})
}

func (l *LinkHeader) Next(pageValue string) {
	l.links = append(l.links, Link{
		URI: l.pageString(pageValue),
		Rel: "next",
	})
}

func (l *LinkHeader) Current(pageValue string) {
	l.links = append(l.links, Link{
		URI: l.pageString(pageValue),
		Rel: "current",
	})
}

func (l *LinkHeader) Previous(pageValue string) {
	l.links = append(l.links, Link{
		URI: l.pageString(pageValue),
		Rel: "previous",
	})
}

func (l *LinkHeader) Last(pageValue string) {
	l.links = append(l.links, Link{
		URI: l.pageString(pageValue),
		Rel: "last",
	})
}

func (l *LinkHeader) SetHeader(header http.Header) {
	header.Add("Link", Format(l.links))
}

func (l *LinkHeader) pageString(pageValue string) string {
	u := l.url
	q := u.Query()
	q.Set(l.pageParam, pageValue)
	u.RawQuery = q.Encode()
	return u.String()
}
