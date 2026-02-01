package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MatchStrategy string

const (
	MatchPartial MatchStrategy = "partial"
	MatchExact   MatchStrategy = "exact"
)

const (
	defaultUserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	defaultAcceptLang   = "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7"
	defaultReferer      = "https://map.naver.com/"
	defaultSearchPrefix = "https://map.naver.com/p/search/"
)

var placeListSelectors = []string{
	"div.Ryr1F#_pcmap_list_scroll_container > ul > li",
	"li.VLTHu",
	"li.UEzoS",
	"ul._3l82D > li",
	"ul._1s-8x > li",
	"div.place_section > ul > li",
	".api_subject_bx > ul > li",
	"div._1EKsQ li.YjsMB",
}

var adSelectors = []string{
	".gU6bV._DHlh",
	".ad_area",
	".ad-badge",
	".OErwL",
	"span.OErwL",
}

var shopNameSelectors = []string{
	"a.place_bluelink span.YwYLL",
	"span.YwYLL",
	".place_bluelink.tWIhh > span.O_Uah",
	"span.place_bluelink",
	"span.TYaxT",
	"span.LDgIH",
	"span.OXiLu",
	"span._3Apve",
	"span.place_bluelink._3Apve",
	".place_bluelink",
	"a.place_link > span",
}

type PlaceItem struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
}

type SearchResult struct {
	Keyword       string
	ShopName      string
	MatchStrategy MatchStrategy
	Rank          int
	Found         bool
	MatchedName   string
	ItemsScanned  int
	Items         []PlaceItem
	SearchURL     string
	IframeURL     string
}

type Engine struct {
	client     *http.Client
	userAgent  string
	acceptLang string
	referer    string
	debugf     func(string, ...interface{})
}

func NewEngine(timeout time.Duration, userAgent string, debugf func(string, ...interface{})) *Engine {
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	if debugf == nil {
		debugf = func(string, ...interface{}) {}
	}
	return &Engine{
		client:     &http.Client{Timeout: timeout},
		userAgent:  userAgent,
		acceptLang: defaultAcceptLang,
		referer:    defaultReferer,
		debugf:     debugf,
	}
}

type HTTPStatusError struct {
	URL    string
	Status int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("http status %d for %s", e.Status, e.URL)
}

type ParseError struct {
	Step    string
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Step, e.Message)
}

func (e *Engine) Search(ctx context.Context, keyword, shopName string, match MatchStrategy) (SearchResult, error) {
	result := SearchResult{
		Keyword:       keyword,
		ShopName:      shopName,
		MatchStrategy: match,
		Rank:          -1,
		Items:         make([]PlaceItem, 0),
	}

	searchURL := buildSearchURL(keyword)
	result.SearchURL = searchURL

	body, err := e.fetch(ctx, searchURL)
	if err != nil {
		return result, err
	}

	iframeURL, err := extractIframeURL(searchURL, body)
	if err != nil {
		return result, err
	}
	result.IframeURL = iframeURL

	iframeBody, err := e.fetch(ctx, iframeURL)
	if err != nil {
		return result, err
	}

	rank, matchedName, items, itemsScanned, err := findRankInHTML(iframeBody, shopName, match)
	if err != nil {
		return result, err
	}

	result.ItemsScanned = itemsScanned
	result.Items = items
	if rank > 0 {
		result.Rank = rank
		result.Found = true
		result.MatchedName = matchedName
	}

	return result, nil
}

func (e *Engine) fetch(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", e.userAgent)
	req.Header.Set("Accept-Language", e.acceptLang)
	req.Header.Set("Referer", e.referer)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	e.debugf("GET %s", urlStr)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPStatusError{URL: urlStr, Status: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func buildSearchURL(keyword string) string {
	encoded := url.PathEscape(keyword)
	return defaultSearchPrefix + encoded + "?searchType=place"
}

func extractIframeURL(searchURL string, html []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return "", &ParseError{Step: "parse_search_html", Message: err.Error()}
	}

	var iframeSrc string
	doc.Find("iframe").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		if id, _ := sel.Attr("id"); id == "searchIframe" {
			src, _ := sel.Attr("src")
			iframeSrc = strings.TrimSpace(src)
			return false
		}
		return true
	})

	if iframeSrc == "" {
		return fallbackIframeURL(keywordFromSearchURL(searchURL)), nil
	}

	base, err := url.Parse(searchURL)
	if err != nil {
		return "", &ParseError{Step: "parse_search_url", Message: err.Error()}
	}

	iframeURL, err := url.Parse(iframeSrc)
	if err != nil {
		return "", &ParseError{Step: "parse_iframe_url", Message: err.Error()}
	}

	resolved := base.ResolveReference(iframeURL)
	return resolved.String(), nil
}

func keywordFromSearchURL(searchURL string) string {
	parsed, err := url.Parse(searchURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(parsed.Path, "/")
	if len(parts) == 0 {
		return ""
	}
	encoded := parts[len(parts)-1]
	keyword, err := url.PathUnescape(encoded)
	if err != nil {
		return ""
	}
	return keyword
}

func fallbackIframeURL(keyword string) string {
	return "https://pcmap.place.naver.com/place/list?query=" + url.QueryEscape(keyword)
}

func findRankInHTML(html []byte, shopName string, match MatchStrategy) (int, string, []PlaceItem, int, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return -1, "", nil, 0, &ParseError{Step: "parse_iframe_html", Message: err.Error()}
	}

	items := findPlaceItems(doc)
	if items.Length() == 0 {
		return -1, "", nil, 0, &ParseError{Step: "find_place_items", Message: "no place items found"}
	}

	rank := 0
	scanned := 0
	matchedRank := -1
	matchedName := ""
	places := make([]PlaceItem, 0)

	items.Each(func(_ int, item *goquery.Selection) {
		if isAdItem(item) {
			return
		}
		rank++
		scanned++

		name := extractShopName(item)
		if name == "" {
			return
		}

		places = append(places, PlaceItem{Rank: rank, Name: name})

		if matchedName == "" && matches(name, shopName, match) {
			matchedName = name
			matchedRank = rank
		}
	})

	if matchedName == "" {
		return -1, "", places, scanned, nil
	}

	return matchedRank, matchedName, places, scanned, nil
}

func findPlaceItems(doc *goquery.Document) *goquery.Selection {
	for _, selector := range placeListSelectors {
		selection := doc.Find(selector)
		if selection.Length() > 0 {
			return selection
		}
	}
	return &goquery.Selection{}
}

func isAdItem(item *goquery.Selection) bool {
	for _, selector := range adSelectors {
		if item.Is(selector) || item.Find(selector).Length() > 0 {
			return true
		}
	}
	return false
}

func extractShopName(item *goquery.Selection) string {
	for _, selector := range shopNameSelectors {
		text := strings.TrimSpace(item.Find(selector).First().Text())
		if text != "" {
			return text
		}
	}
	return ""
}

func matches(candidate, target string, match MatchStrategy) bool {
	left := normalize(candidate)
	right := normalize(target)
	if left == "" || right == "" {
		return false
	}

	switch match {
	case MatchExact:
		return left == right
	default:
		return strings.Contains(left, right) || strings.Contains(right, left)
	}
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}
