package voiceml

import (
	"context"
	"net/url"
	"reflect"
	"strings"
)

// Page is the pagination envelope embedded in every /2010-04-01 list
// response. The concrete resource list field (Calls, Conferences, ...) is
// declared in its own list type alongside this embed.
//
// NextPageURI is the only pagination input the SDK uses: it is a relative
// URI (path plus query, carrying the server's PageToken) that
// Client.NextPage and the Iterate methods request verbatim on the host
// that served the page. Page numbers are never synthesised.
type Page struct {
	Page            int    `json:"page"`
	PageSize        int    `json:"page_size"`
	Start           *int   `json:"start,omitempty"`
	End             *int   `json:"end,omitempty"`
	FirstPageURI    string `json:"first_page_uri,omitempty"`
	NextPageURI     string `json:"next_page_uri,omitempty"`
	PreviousPageURI string `json:"previous_page_uri,omitempty"`
	URI             string `json:"uri,omitempty"`

	origin pageOrigin
}

// VoiceV1Meta is the pagination envelope embedded in every /v1 list
// response (Voice, Conversations, Messaging, Assistants, Pricing) under the
// `meta` key. NextPageURL is an absolute URL; the SDK follows it only when
// its scheme and host equal those of the host that served the page.
type VoiceV1Meta struct {
	FirstPageURL    *string `json:"first_page_url,omitempty"`
	NextPageURL     *string `json:"next_page_url,omitempty"`
	PreviousPageURL *string `json:"previous_page_url,omitempty"`
	URL             *string `json:"url,omitempty"`
	Page            *int    `json:"page,omitempty"`
	PageSize        *int    `json:"page_size,omitempty"`
	Key             *string `json:"key,omitempty"`

	origin pageOrigin
}

// pageOrigin records where a page was fetched from.
type pageOrigin struct {
	base string // scheme://host[:port]
	self string // full request URL
}

// PageLinks is implemented by the two pagination envelopes, Page (and every
// /2010-04-01 list type, which embeds it) and VoiceV1Meta. Pass one to
// Client.NextPage.
type PageLinks interface {
	nextPageLink() (link string, origin pageOrigin)
}

func (p Page) nextPageLink() (string, pageOrigin) { return p.NextPageURI, p.origin }

func (m VoiceV1Meta) nextPageLink() (string, pageOrigin) {
	if m.NextPageURL == nil {
		return "", m.origin
	}
	return *m.NextPageURL, m.origin
}

// HasNextPage reports whether the page links to another page.
func (p Page) HasNextPage() bool { return p.NextPageURI != "" }

// HasNextPage reports whether the page links to another page.
func (m VoiceV1Meta) HasNextPage() bool { return m.NextPageURL != nil && *m.NextPageURL != "" }

type originSetter interface{ setOrigin(pageOrigin) }

func (p *Page) setOrigin(o pageOrigin)        { p.origin = o }
func (m *VoiceV1Meta) setOrigin(o pageOrigin) { m.origin = o }

var metaType = reflect.TypeOf(VoiceV1Meta{})

// stampOrigin records the request URL on a decoded list envelope so the
// next page can be resolved against the host that served it.
func stampOrigin(out any, u *url.URL) {
	if u == nil {
		return
	}
	o := pageOrigin{base: u.Scheme + "://" + u.Host, self: u.String()}
	if s, ok := out.(originSetter); ok {
		s.setOrigin(o)
		return
	}
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}
	if f := v.Elem().FieldByName("Meta"); f.IsValid() && f.Type() == metaType && f.CanAddr() {
		f.Addr().Interface().(*VoiceV1Meta).origin = o
	}
}

// NextPage fetches the page after current and decodes it into next, which
// must be a pointer to the same list type. It reports false, without
// sending anything, when current is the last page.
//
// For /2010-04-01 lists pass the list itself (or its embedded Page); for
// /v1 lists pass the list's Meta:
//
//	var more voiceml.CallList
//	ok, err := c.NextPage(ctx, page, &more)
//
// The link is requested verbatim. A link that resolves to a different
// scheme or host than the page was served from is refused with
// *PaginationError and no request is made.
func (c *Client) NextPage(ctx context.Context, current PageLinks, next any) (bool, error) {
	link, origin := current.nextPageLink()
	if link == "" {
		return false, nil
	}
	target, err := c.resolvePageLink(link, origin)
	if err != nil {
		return false, err
	}
	if err := c.t.do(ctx, requestOpts{method: "GET", rawURL: target}, next); err != nil {
		return false, err
	}
	return true, nil
}

// resolvePageLink turns a next-page link into an absolute URL on an allowed
// host. Relative links ("/2010-04-01/...") resolve against the host that
// served the page; absolute links must already point there. Pages decoded
// outside the SDK carry no origin: relative links then resolve against
// BaseURL and absolute links must match one of the configured product hosts.
func (c *Client) resolvePageLink(link string, origin pageOrigin) (string, error) {
	allowed := []string{origin.base}
	if origin.base == "" {
		allowed = []string{originOf(c.BaseURL), originOf(c.MessagingBaseURL), originOf(c.ConversationsBaseURL)}
	}
	var target string
	switch {
	case strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "//"):
		target = allowed[0] + link
	default:
		u, err := url.Parse(link)
		if err != nil || !u.IsAbs() || u.Host == "" {
			return "", &PaginationError{Message: "next page link " + quoteLink(link) + " is neither a rooted path nor an absolute URL"}
		}
		ok := false
		for _, a := range allowed {
			au, err := url.Parse(a)
			if err == nil && a != "" && sameOrigin(au, u) {
				ok = true
				break
			}
		}
		if !ok {
			return "", &PaginationError{Message: "refusing to follow next page link to " + u.Scheme + "://" + u.Host + ": it is not the host that served the page"}
		}
		target = link
	}
	if origin.self != "" && target == origin.self {
		return "", &PaginationError{Message: "next page link repeats the page just fetched"}
	}
	return target, nil
}

func originOf(base string) string {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func quoteLink(s string) string {
	if len(s) > 120 {
		s = s[:120] + "..."
	}
	return `"` + s + `"`
}

// Pager walks a paginated list lazily, one page at a time, following each
// page's next link verbatim. No request is made until the first call to
// Next, and none after the caller stops calling it. Iteration ends after a
// page with no next link or no items.
//
//	it := c.Calls.Iterate(ctx, voiceml.ListCallsParams{PageSize: voiceml.Int(100)})
//	for it.Next() {
//		call := it.Current()
//		...
//	}
//	if err := it.Err(); err != nil { ... }
//
// A Pager is not safe for concurrent use. On Go 1.23+ Pager.All adapts it
// to a range-over-func iterator.
type Pager[T any] struct {
	ctx     context.Context
	c       *Client
	fetch   func(ctx context.Context, opts requestOpts) ([]T, PageLinks, error)
	pending *requestOpts
	stopErr error // reported once the current page's items are consumed
	buf     []T
	idx     int
	cur     T
	err     error
	pages   int
}

func newPager[L any, T any](ctx context.Context, c *Client, t *transport, first requestOpts, split func(*L) ([]T, PageLinks)) *Pager[T] {
	return &Pager[T]{
		ctx:     ctx,
		c:       c,
		pending: &first,
		fetch: func(ctx context.Context, opts requestOpts) ([]T, PageLinks, error) {
			var page L
			if err := t.do(ctx, opts, &page); err != nil {
				return nil, nil, err
			}
			items, links := split(&page)
			return items, links, nil
		},
	}
}

// Next advances to the next item, fetching the next page when the current
// one is exhausted. It returns false at the end of the list or on error;
// check Err afterwards.
func (p *Pager[T]) Next() bool {
	for {
		if p.err != nil {
			return false
		}
		if p.idx < len(p.buf) {
			p.cur = p.buf[p.idx]
			p.idx++
			return true
		}
		if p.pending == nil {
			if p.stopErr != nil {
				p.err, p.stopErr = p.stopErr, nil
			}
			return false
		}
		opts := *p.pending
		p.pending = nil
		items, links, err := p.fetch(p.ctx, opts)
		if err != nil {
			p.err = err
			return false
		}
		p.pages++
		p.buf, p.idx = items, 0
		if len(items) == 0 {
			return false
		}
		link, origin := links.nextPageLink()
		if link == "" {
			continue
		}
		target, err := p.c.resolvePageLink(link, origin)
		if err != nil {
			// Deliver this page's items first; report the error after.
			p.stopErr = err
			continue
		}
		p.pending = &requestOpts{method: "GET", rawURL: target}
	}
}

// Current returns the item most recently produced by Next.
func (p *Pager[T]) Current() T { return p.cur }

// Err returns the first error encountered, if any.
func (p *Pager[T]) Err() error { return p.err }

// PagesFetched reports how many pages have been requested so far.
func (p *Pager[T]) PagesFetched() int { return p.pages }
