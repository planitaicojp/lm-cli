# lm-cli — LINE Messaging API CLI

LINE Messaging API を操作するためのコマンドラインツールです。DevOps エンジニア、Bot 開発者、マーケターが
コマンドラインから LINE OA を管理できます。

## 特徴

- **3種類のトークン認証**: 長期トークン (longterm)、ステートレス (stateless)、JWT (v2、Phase 4)
- **全メッセージタイプ対応**: push、multicast、broadcast、narrowcast、reply
- **出力フォーマット**: table / json / yaml / csv
- **CI/CD フレンドリー**: `--no-input`、`LM_TOKEN` 環境変数、終了コード
- **複数プロファイル管理**: 本番/ステージング/テスト環境を切り替え可能

## インストール

```bash
# Go でインストール
go install github.com/crowdy/lm-cli@latest

# バイナリ（GitHub Releases）
# https://github.com/crowdy/lm-cli/releases
```

## クイックスタート

```bash
# 認証（長期トークン）
lm auth login

# Bot 情報確認
lm bot info

# メッセージ送信
lm message push <userId> "こんにちは！"

# JSON 出力
lm bot info --format json

# 環境変数でトークン指定（CI/CD）
LM_TOKEN=xxx lm message push <userId> "デプロイ完了"
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `lm auth login` | 認証設定 |
| `lm auth status` | 認証状態確認 |
| `lm auth list` | プロファイル一覧 |
| `lm auth switch <profile>` | プロファイル切り替え |
| `lm message push <userId> <text>` | ユーザーへメッセージ送信 |
| `lm message multicast <text> --to id,id` | 複数ユーザーへ送信 |
| `lm message broadcast <text>` | 全フォロワーへ送信 |
| `lm message reply <replyToken> <text>` | Webhook イベントに返信 |
| `lm bot info` | Bot 情報取得 |
| `lm bot quota` | メッセージ配信数上限取得 |
| `lm bot consumption` | 配信数使用量取得 |
| `lm user profile <userId>` | ユーザープロファイル取得 |
| `lm user followers` | フォロワー ID 一覧 |
| `lm group info <groupId>` | グループ情報取得 |
| `lm group members <groupId>` | グループメンバー一覧 |
| `lm group leave <groupId>` | グループ退出 |
| `lm richmenu list` | リッチメニュー一覧 |
| `lm richmenu create --file menu.json` | リッチメニュー作成 |
| `lm richmenu upload <id> image.jpg` | 画像アップロード |
| `lm webhook get` | Webhook URL 取得 |
| `lm webhook set <url>` | Webhook URL 設定 |
| `lm webhook test` | Webhook テスト |
| `lm audience list` | オーディエンス一覧 |
| `lm insight followers` | フォロワー統計 |
| `lm content get <messageId>` | コンテンツダウンロード |

## 設定

設定ディレクトリ: `~/.config/lm/` (`LM_CONFIG_DIR` で変更可)

### 環境変数

| 変数 | 説明 |
|------|------|
| `LM_TOKEN` | トークン直接指定（auth バイパス） |
| `LM_CHANNEL_ID` | チャネル ID |
| `LM_CHANNEL_SECRET` | チャネルシークレット |
| `LM_PROFILE` | 使用プロファイル名 |
| `LM_FORMAT` | 出力フォーマット (table/json/yaml/csv) |
| `LM_CONFIG_DIR` | 設定ディレクトリパス |
| `LM_NO_INPUT` | 非対話モード (1 or true) |
| `LM_ENDPOINT` | API ベース URL 上書き（テスト用） |
| `LM_DEBUG` | デバッグログ (1, true, api) |

## 認証

### 長期トークン（最シンプル）
LINE Developers Console から発行した長期トークンを使用します。

```bash
lm auth login
# Channel ID と長期トークンを入力
```

### ステートレストークン（CI/CD 推奨）
Channel ID + Channel Secret からトークンを自動発行・更新します。

```bash
lm auth login --type stateless
# Channel ID と Channel Secret を入力
```

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | 認証エラー (401/403) |
| 3 | リソースが見つからない (404) |
| 4 | バリデーションエラー |
| 5 | LINE API エラー |
| 6 | ネットワークエラー |
| 10 | キャンセル (Ctrl+C) |
| 11 | レート制限 (429) |

## エージェント統合

```bash
# 失敗を検出
lm message push $USER_ID "deploy done" || echo "LINE 送信失敗: exit $?"

# JSON 出力でスクリプト処理
lm bot info --format json | jq .displayName

# 非対話モード
LM_TOKEN=$TOKEN LM_NO_INPUT=1 lm message push $USER_ID "CI 通知"
```

## 開発

```bash
make build    # ./lm にビルド
make test     # go test ./...
make lint     # golangci-lint run ./...
make coverage # カバレッジレポート生成
```

## ライセンス

MIT
