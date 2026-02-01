---
name: naver-place-rank-cli
description: Use to run the local Go CLI that queries Naver Place rankings for a keyword + shop and returns JSON. Trigger when asked to check Naver Place rank, list all places for a keyword, or produce machine-readable ranking output from this repo.
---

# Naver Place Rank CLI

## Quick start

Build and run from the repo root:

```bash
go build -o naver-place-rank
./naver-place-rank --keyword "<keyword>" --shop "<shop name>" --pretty
```

Or run directly:

```bash
go run . --keyword "<keyword>" --shop "<shop name>"
```

## Output (minimal JSON)

Fields:
- `ok` (bool)
- `keyword` (string)
- `shop_name` (string)
- `found` (bool)
- `rank` (int, -1 if not found)
- `matched_name` (string)
- `items` (array of `{rank, name}` for all non-ad results)
- `error` (object or null)

Semantics:
- `ok=true` + `found=false` => request succeeded, but shop not found.
- `ok=false` => request failed; check `error.code` and `error.message`.

## Extended output

Use `--full` to add debug fields:
- `match_strategy`
- `items_scanned`
- `search_url`
- `iframe_url`
- `timestamp`
- `duration_ms`

## Options

- `--match` (`partial` default, or `exact`)
- `--timeout` (default `10s`)
- `--user-agent` (override UA)
- `--full` (extended JSON)
- `--pretty` (pretty JSON)
- `--debug` (debug logs to stderr)

## Guidance

- Prefer `partial` for brand/alias matching; use `exact` for strict equality.
- Use `--full` only when debugging parsing issues.
- Avoid high-frequency queries to reduce load on Naver.
