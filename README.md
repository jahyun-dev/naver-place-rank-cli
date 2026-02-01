# naver-place-rank (Go CLI)

Minimal, JSON-first CLI that looks up a shop's rank for a given keyword on Naver Place.

## Build

```bash
make build
```

## Test

```bash
make test
```

## Usage

```bash
./naver-place-rank --keyword "uijeongbu hair salon" --shop "Juno Hair Uijeongbu Station"
./naver-place-rank "uijeongbu hair salon" "Juno Hair Uijeongbu Station"
```

Optional flags:

- `--match` : `partial` (default) or `exact`
- `--timeout` : HTTP timeout (default `10s`)
- `--user-agent` : override User-Agent header
- `--full` : include extended JSON fields
- `--pretty` : pretty-print JSON
- `--debug` : debug logs to stderr

## Output (JSON)

Default (minimal):

```json
{
  "ok": true,
  "keyword": "uijeongbu hair salon",
  "shop_name": "Juno Hair Uijeongbu Station",
  "found": true,
  "rank": 3,
  "matched_name": "Juno Hair Uijeongbu Station",
  "items": [
    { "rank": 1, "name": "First Place" },
    { "rank": 2, "name": "Second Place" },
    { "rank": 3, "name": "Juno Hair Uijeongbu Station" }
  ],
  "error": null
}
```

`items` contains all non-ad place names in rank order from the search results.

Extended output with `--full` adds:

- `match_strategy`
- `items_scanned`
- `search_url`
- `iframe_url`
- `timestamp`
- `duration_ms`

When the request succeeds but the shop is not found, `ok` is `true`, `found` is `false`, and `rank` is `-1`.
When a request fails, `ok` is `false` and `error` is populated.

## Exit codes

- `0` success (including not-found)
- `1` request/parse error
- `2` invalid arguments

## Release (multi-platform)

Automated via GitHub Actions on tag push (`v*`) using GoReleaser. You can also run it locally.

```bash
make release-snapshot   # local snapshot builds into dist/
make release            # real release to GitHub (requires token)
```

To validate GoReleaser config:

```bash
make check
```

## Lint

```bash
make lint
```
