package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

type MinimalResponse struct {
	OK          bool        `json:"ok"`
	Keyword     string      `json:"keyword"`
	ShopName    string      `json:"shop_name"`
	Found       bool        `json:"found"`
	Rank        int         `json:"rank"`
	MatchedName string      `json:"matched_name"`
	Items       []PlaceItem `json:"items"`
	Error       *ErrorInfo  `json:"error"`
}

type FullResponse struct {
	MinimalResponse
	MatchStrategy string `json:"match_strategy"`
	ItemsScanned  int    `json:"items_scanned"`
	SearchURL     string `json:"search_url"`
	IframeURL     string `json:"iframe_url"`
	Timestamp     string `json:"timestamp"`
	DurationMs    int64  `json:"duration_ms"`
}

type options struct {
	Keyword   string
	ShopName  string
	Match     MatchStrategy
	Timeout   time.Duration
	UserAgent string
	Pretty    bool
	Debug     bool
	Full      bool
}

func main() {
	start := time.Now()
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(os.Stderr)
			return
		}
		minimal := MinimalResponse{
			OK:          false,
			Keyword:     opts.Keyword,
			ShopName:    opts.ShopName,
			Found:       false,
			Rank:        -1,
			MatchedName: "",
			Items:       []PlaceItem{},
			Error: &ErrorInfo{
				Code:    "invalid_args",
				Message: err.Error(),
			},
		}
		if opts.Full {
			writeJSON(FullResponse{
				MinimalResponse: minimal,
				MatchStrategy:   string(opts.Match),
				ItemsScanned:    0,
				SearchURL:       "",
				IframeURL:       "",
				Timestamp:       time.Now().UTC().Format(time.RFC3339),
				DurationMs:      time.Since(start).Milliseconds(),
			}, opts.Pretty)
		} else {
			writeJSON(minimal, opts.Pretty)
		}
		os.Exit(2)
	}

	debugf := func(string, ...interface{}) {}
	if opts.Debug {
		debugf = func(format string, args ...interface{}) {
			fmt.Fprintf(os.Stderr, "debug: "+format+"\n", args...)
		}
	}

	engine := NewEngine(opts.Timeout, opts.UserAgent, debugf)
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	result, searchErr := engine.Search(ctx, opts.Keyword, opts.ShopName, opts.Match)

	minimal := MinimalResponse{
		OK:          searchErr == nil,
		Keyword:     opts.Keyword,
		ShopName:    opts.ShopName,
		Found:       result.Found,
		Rank:        result.Rank,
		MatchedName: result.MatchedName,
		Items:       result.Items,
		Error:       nil,
	}

	if searchErr != nil {
		minimal.OK = false
		minimal.Found = false
		minimal.Rank = -1
		minimal.MatchedName = ""
		minimal.Items = []PlaceItem{}
		minimal.Error = mapError(searchErr)
	}

	if opts.Full {
		itemsScanned := result.ItemsScanned
		searchURL := result.SearchURL
		iframeURL := result.IframeURL
		if searchErr != nil {
			itemsScanned = 0
			searchURL = ""
			iframeURL = ""
		}
		writeJSON(FullResponse{
			MinimalResponse: minimal,
			MatchStrategy:   string(opts.Match),
			ItemsScanned:    itemsScanned,
			SearchURL:       searchURL,
			IframeURL:       iframeURL,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			DurationMs:      time.Since(start).Milliseconds(),
		}, opts.Pretty)
	} else {
		writeJSON(minimal, opts.Pretty)
	}

	if searchErr != nil {
		os.Exit(1)
	}
}

func parseFlags(args []string) (options, error) {
	opts := options{
		Match:   MatchPartial,
		Timeout: 10 * time.Second,
	}

	fs := flag.NewFlagSet("naver-place-rank", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&opts.Keyword, "keyword", "", "Search keyword")
	fs.StringVar(&opts.ShopName, "shop", "", "Shop name to match")
	match := fs.String("match", string(MatchPartial), "Match strategy: partial or exact")
	fs.DurationVar(&opts.Timeout, "timeout", 10*time.Second, "HTTP timeout (e.g., 10s)")
	fs.StringVar(&opts.UserAgent, "user-agent", defaultUserAgent, "User-Agent header")
	fs.BoolVar(&opts.Pretty, "pretty", false, "Pretty-print JSON")
	fs.BoolVar(&opts.Debug, "debug", false, "Debug logs to stderr")
	fs.BoolVar(&opts.Full, "full", false, "Include extended fields in JSON output")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}

	opts.Match = MatchStrategy(strings.ToLower(strings.TrimSpace(*match)))
	if opts.Match != MatchPartial && opts.Match != MatchExact {
		return opts, fmt.Errorf("invalid match strategy: %s", *match)
	}

	remaining := fs.Args()
	switch {
	case opts.Keyword == "" && opts.ShopName == "" && len(remaining) == 2:
		opts.Keyword = remaining[0]
		opts.ShopName = remaining[1]
	case opts.Keyword == "" && opts.ShopName != "" && len(remaining) == 1:
		opts.Keyword = remaining[0]
	case opts.Keyword != "" && opts.ShopName == "" && len(remaining) == 1:
		opts.ShopName = remaining[0]
	case len(remaining) != 0:
		return opts, fmt.Errorf("unexpected positional arguments")
	}

	if strings.TrimSpace(opts.Keyword) == "" {
		return opts, fmt.Errorf("keyword is required")
	}
	if strings.TrimSpace(opts.ShopName) == "" {
		return opts, fmt.Errorf("shop name is required")
	}

	return opts, nil
}

func mapError(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		return &ErrorInfo{
			Code:    "http_status",
			Message: err.Error(),
			Status:  statusErr.Status,
		}
	}

	var parseErr *ParseError
	if errors.As(err, &parseErr) {
		return &ErrorInfo{
			Code:    "parse_error",
			Message: err.Error(),
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &ErrorInfo{
			Code:    "timeout",
			Message: err.Error(),
		}
	}

	return &ErrorInfo{
		Code:    "request_error",
		Message: err.Error(),
	}
}

func writeJSON(payload interface{}, pretty bool) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	_ = encoder.Encode(payload)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  naver-place-rank --keyword <keyword> --shop <shop name> [--match partial|exact]")
	fmt.Fprintln(w, "  naver-place-rank <keyword> <shop name>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --keyword      Search keyword (required if not positional)")
	fmt.Fprintln(w, "  --shop         Shop name to match (required if not positional)")
	fmt.Fprintln(w, "  --match        Match strategy: partial (default) or exact")
	fmt.Fprintln(w, "  --timeout      HTTP timeout (default 10s)")
	fmt.Fprintln(w, "  --user-agent   User-Agent header")
	fmt.Fprintln(w, "  --full         Include extended fields in JSON output")
	fmt.Fprintln(w, "  --pretty       Pretty-print JSON")
	fmt.Fprintln(w, "  --debug        Debug logs to stderr")
}
