---
name: naver-place-rank-cli
description: Use to run the local Go CLI that queries Naver Place rankings for a keyword + shop and returns JSON. Trigger when asked to check Naver Place rank, list all places for a keyword, or produce machine-readable ranking output from this repo.
---

# Naver Place Rank CLI

## Download (prebuilt binaries)

Get a platform binary from GitHub Releases and extract it.

Release page:

```text
https://github.com/jahyun-dev/naver-place-rank-cli/releases
```

Example (macOS arm64):

```bash
VERSION="v0.1.0"
curl -L -o naver-place-rank.tar.gz "https://github.com/jahyun-dev/naver-place-rank-cli/releases/download/${VERSION}/naver-place-rank_${VERSION#v}_darwin_arm64.tar.gz"
tar -xzf naver-place-rank.tar.gz
./naver-place-rank --help
```

Example (Linux amd64):

```bash
VERSION="v0.1.0"
curl -L -o naver-place-rank.tar.gz "https://github.com/jahyun-dev/naver-place-rank-cli/releases/download/${VERSION}/naver-place-rank_${VERSION#v}_linux_amd64.tar.gz"
tar -xzf naver-place-rank.tar.gz
./naver-place-rank --help
```

Example (Windows amd64, PowerShell):

```powershell
$VERSION = "v0.1.0"
Invoke-WebRequest -Uri "https://github.com/jahyun-dev/naver-place-rank-cli/releases/download/$VERSION/naver-place-rank_$($VERSION.Substring(1))_windows_amd64.zip" -OutFile "naver-place-rank.zip"
Expand-Archive -Path "naver-place-rank.zip" -DestinationPath "."
.\naver-place-rank.exe --help
```

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
