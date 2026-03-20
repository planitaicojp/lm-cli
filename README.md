# lm — LINE Messaging API CLI

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

LINE Messaging API をコマンドラインから操作するための CLI ツールです。
Bot 開発・DevOps 通知・マーケティング配信をスクリプトや CI/CD から手軽に実行できます。

---

## インストール

```bash
# ソースからビルド
git clone https://github.com/t-kim-planitai/lm-cli.git
cd lm-cli
make install        # $GOPATH/bin/lm にインストール

# または手元でビルドするだけ
make build          # ./lm
```

> **Note:** Go モジュールパスは `github.com/crowdy/lm-cli` です。

---

## クイックスタート

```bash
# 1. 認証（長期トークン）
lm auth login

# 2. Bot 情報確認
lm bot info

# 3. メッセージ送信
lm message push <userId> "こんにちは！"

# 4. JSON 出力
lm bot info --format json

# 5. 環境変数でトークン指定（CI/CD）
LM_TOKEN=xxx lm message push <userId> "デプロイ完了"
```

---

## 認証

3 種類のトークンモードをサポートしています。

### 長期トークン（最シンプル）

LINE Developers Console で発行した長期チャネルアクセストークンを使用します。

```bash
lm auth login
# → Channel ID とトークンを対話入力
```

### ステートレストークン（CI/CD 推奨）

Channel ID + Channel Secret からトークンを自動発行・更新します。

```bash
lm auth login --type stateless
# → Channel ID と Channel Secret を対話入力
```

### 複数プロファイル

```bash
lm auth login --profile prod
lm auth login --profile staging
lm auth switch prod
lm auth list
```

---

## コマンド一覧

### auth

| コマンド | 説明 |
|---------|------|
| `lm auth login [--type longterm\|stateless] [--profile name]` | 認証設定 |
| `lm auth logout` | トークン・認証情報の削除 |
| `lm auth status` | 認証状態の確認 |
| `lm auth list` | プロファイル一覧 |
| `lm auth switch <profile>` | アクティブプロファイルの切り替え |
| `lm auth token` | 現在のトークンを stdout に出力（スクリプト用） |
| `lm auth remove <profile>` | プロファイルの完全削除 |

### message

| コマンド | 説明 |
|---------|------|
| `lm message push <userId> <text>` | ユーザーへメッセージ送信 |
| `lm message multicast <text> --to id,id,...` | 複数ユーザーへ一斉送信 |
| `lm message broadcast <text>` | 全フォロワーへ送信 |
| `lm message narrowcast <text> [--filter-file filter.json]` | ナローキャスト |
| `lm message reply <replyToken> <text>` | Webhook イベントへ返信 |

メッセージタイプ指定:

```bash
lm message push <userId> <text>                        # テキスト（デフォルト）
lm message push <userId> --type sticker <pkgId> <id>   # スタンプ
lm message push <userId> --file msg.json               # JSON ファイル
```

### bot

| コマンド | 説明 |
|---------|------|
| `lm bot info` | Bot プロフィール取得 |
| `lm bot quota` | メッセージ配信数上限取得 |
| `lm bot consumption` | 当月の配信数使用量取得 |

### user / group

| コマンド | 説明 |
|---------|------|
| `lm user profile <userId>` | ユーザープロフィール取得 |
| `lm user followers [--limit N] [--start token]` | フォロワー ID 一覧 |
| `lm group info <groupId>` | グループ情報取得 |
| `lm group members <groupId>` | グループメンバー一覧 |
| `lm group leave <groupId>` | グループ退出 |

### richmenu

| コマンド | 説明 |
|---------|------|
| `lm richmenu create --file menu.json` | リッチメニュー作成 |
| `lm richmenu get <richMenuId>` | リッチメニュー取得 |
| `lm richmenu list` | リッチメニュー一覧 |
| `lm richmenu delete <richMenuId>` | リッチメニュー削除 |
| `lm richmenu upload <richMenuId> <image.jpg\|png>` | 画像アップロード |
| `lm richmenu default get\|set\|unset` | デフォルトリッチメニュー管理 |
| `lm richmenu alias create --file alias.json` | エイリアス作成 |
| `lm richmenu alias list` | エイリアス一覧 |

### webhook / audience / insight / content

| コマンド | 説明 |
|---------|------|
| `lm webhook get` | Webhook URL 取得 |
| `lm webhook set <url>` | Webhook URL 設定 |
| `lm webhook test` | Webhook テスト |
| `lm audience create --file audience.json` | オーディエンス作成 |
| `lm audience get\|list\|delete` | オーディエンス管理 |
| `lm insight followers [--date YYYYMMDD]` | フォロワー統計 |
| `lm insight delivery --type broadcast [--date YYYYMMDD]` | 配信統計 |
| `lm content get <messageId> [--output file]` | コンテンツダウンロード |

---

## グローバルフラグ

| フラグ | 説明 |
|--------|------|
| `--profile string` | 使用するプロファイル名 |
| `--format string` | 出力形式: `table`\|`json`\|`yaml`\|`csv` (デフォルト: `table`) |
| `--no-input` | 対話入力を無効化（必須項目が欠けていれば exit 4） |
| `--quiet` | 進捗メッセージを抑制 |
| `--verbose` | HTTP リクエスト/レスポンスをログ出力 |

---

## 設定

設定ディレクトリ: `~/.config/lm/`（`LM_CONFIG_DIR` で変更可）

| ファイル | 内容 | パーミッション |
|---------|------|--------------|
| `config.yaml` | プロファイル・デフォルト設定 | 0644 |
| `credentials.yaml` | Channel Secret | 0600 |
| `tokens.yaml` | アクセストークン・有効期限 | 0600 |

### 環境変数

優先順位: `config.yaml` < 環境変数 < CLI フラグ

| 変数 | 説明 |
|------|------|
| `LM_TOKEN` | トークン直接指定（auth 処理を完全バイパス） |
| `LM_CHANNEL_ID` | チャネル ID |
| `LM_CHANNEL_SECRET` | チャネルシークレット |
| `LM_PROFILE` | 使用プロファイル名 |
| `LM_FORMAT` | 出力フォーマット |
| `LM_CONFIG_DIR` | 設定ディレクトリパス |
| `LM_NO_INPUT` | 非対話モード (`1` or `true`) |
| `LM_ENDPOINT` | API ベース URL 上書き（テスト用） |
| `LM_DEBUG` | デバッグログ (`1`, `true`, `api`) |

---

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | 認証エラー (401/403) |
| 3 | リソースが見つからない (404) |
| 4 | バリデーションエラー / `--no-input` で入力不足 |
| 5 | LINE API エラー |
| 6 | ネットワークエラー |
| 10 | キャンセル (Ctrl+C) |
| 11 | レート制限 (429) |

---

## CI/CD・スクリプト連携

```bash
# 送信失敗を検出
lm message push "$USER_ID" "deploy done" || echo "LINE 送信失敗 (exit $?)"

# JSON でパイプ処理
lm bot info --format json | jq -r .displayName

# 非対話モード（GitHub Actions など）
LM_TOKEN="${{ secrets.LINE_TOKEN }}" LM_NO_INPUT=1 \
  lm message push "$USER_ID" "CI 通知: $MESSAGE"

# トークンを他のコマンドへ渡す
TOKEN=$(lm auth token)
```

---

## 開発

```bash
make build     # ./lm をビルド
make test      # go test ./...
make lint      # golangci-lint run ./...
make coverage  # coverage.html を生成
make install   # $GOPATH/bin/lm にインストール
```

### 新コマンドの追加手順

1. `cmd/<domain>/<domain>.go` に `Cmd` var を作成
2. `cmd/root.go` で `AddCommand` 登録
3. `internal/api/<domain>.go` に API struct を実装
4. `internal/model/<domain>.go` にリクエスト/レスポンス型を定義
5. `internal/api/<domain>_test.go` に `httptest` でテストを記述

---

## ライセンス

MIT
