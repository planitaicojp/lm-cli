# lm-cli Developer Guide

## Overview
- Binary: `lm`
- Module: `github.com/crowdy/lm-cli`
- Purpose: CLI for the LINE Messaging API
- Pattern reference: `../conoha-cli` (same author, same conventions)

## Architecture Rules
1. コマンドは `cmd/<domain>/<domain>.go` に配置
2. HTTP 呼び出しは必ず `internal/api/` 経由。`cmd/` から `net/http` を直接呼ばない
3. データ出力は必ず `internal/output.Formatter` 経由。`fmt.Printf` でデータを出力しない
4. エラーは `internal/errors` の型付きエラーを使用（exit code 付き）
5. 人間向けメッセージ（確認、進捗）は `fmt.Fprintf(os.Stderr, ...)`
6. API クライアント取得は `cmd/cmdutil.NewClient(cmd)` のみ

## Config
- Config dir: `~/.config/lm/` (LM_CONFIG_DIR で上書き可)
- Files: config.yaml, credentials.yaml (0600), tokens.yaml (0600)

## Auth
- `LM_TOKEN` 環境変数で全 auth 処理をバイパス可能
- EnsureToken() は `internal/api/auth.go` に実装

## LINE API
- Base URL: `https://api.line.me` (LM_ENDPOINT で上書き可)
- Auth header: `Authorization: Bearer <token>`
- Error: `{"message": "...", "details": [{"message": "...", "property": "..."}]}`
- richmenu image upload: `Content-Type: image/jpeg|png`（JSON ではない）
- content download: バイナリレスポンス

## Common Commands
```
make build    # ./lm にビルド
make test     # go test ./...
make lint     # golangci-lint run ./...
make install  # $GOPATH/bin にインストール
```

## Adding a New Command
1. `cmd/<domain>/<domain>.go` に Cmd var を作成
2. `cmd/root.go` で AddCommand 登録
3. `internal/api/<domain>.go` にドメイン API struct を作成
4. `internal/model/<domain>.go` にリクエスト/レスポンス型を作成
5. `internal/api/<domain>_test.go` に httptest でテストを書く

## Testing
- Unit tests: table-driven, t.Run()
- HTTP tests: httptest.NewServer() でモックサーバー
- Integration: build tag `//go:build integration`, LM_TOKEN + LM_CHANNEL_ID 必要
- Coverage: make coverage → coverage.html

## LINE Messaging API Reference
- https://developers.line.biz/en/reference/messaging-api/
- Rate limits: push=500req/s, broadcast=1req/s, multicast=500req/s
- Message quota: プランによる（Free: 200通/月）
