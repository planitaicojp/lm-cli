# lm — LINE Messaging API CLI Specification

**Version**: 0.1.0
**Status**: Initial Release
**Module**: `github.com/crowdy/lm-cli`
**Binary**: `lm`

---

## Scope

LINE Messaging API の主要エンドポイントをカバーする CLI ツール。
Bot 開発・DevOps 通知・マーケティング配信を、スクリプト・CI/CD・ターミナルから均一に操作できることを目的とする。

「コマンドが正しく動く」だけでは不十分。**実行前後の状態変化が予測可能で、失敗が安全で、出力がスクリプタブル** であることを品質の基準とする。

---

## Supported Endpoints

| Domain | Endpoint | Commands |
|--------|----------|----------|
| OAuth | `api.line.me/oauth2/v2.1/token` | `auth login --type stateless` |
| Messaging | `api.line.me/v2/bot/message/*` | `message push/multicast/broadcast/narrowcast/reply` |
| Bot | `api.line.me/v2/bot/info` | `bot info` |
| Quota | `api.line.me/v2/bot/message/quota*` | `bot quota`, `bot consumption` |
| User | `api.line.me/v2/bot/profile/{userId}` | `user profile` |
| Followers | `api.line.me/v2/bot/followers/ids` | `user followers` |
| Group | `api.line.me/v2/bot/group/{groupId}/*` | `group info/members/leave` |
| Rich Menu | `api.line.me/v2/bot/richmenu/*` | `richmenu create/get/list/delete/upload/default/alias` |
| Rich Menu Image | **`api-data.line.me`**`/v2/bot/richmenu/{id}/content` | `richmenu upload` ※別ホスト |
| Webhook | `api.line.me/v2/bot/channel/webhook/endpoint` | `webhook get/set/test` |
| Audience | `api.line.me/v2/bot/audienceGroup/*` | `audience create/get/list/delete` |
| Insight | `api.line.me/v2/bot/insight/*` | `insight followers/delivery` |
| Content | `api.line.me/v2/bot/message/{id}/content` | `content get` |

> **Note on dual hosts**: rich menu 画像アップロードのみ `api-data.line.me` を使用する。
> `LM_ENDPOINT` による override はメッセージ系エンドポイントのみに効く。テスト用モックでは
> `api-data.line.me` も別途 stub が必要。

---

## LINE API Constraints (5年の現場知識)

これらの制約を CLI が透過的に扱わない場合、利用者は API エラーで初めて気づく。SPEC として明記し実装で対処する。

| API | 制約 | 現在の対応 | 要対応バージョン |
|-----|------|-----------|----------------|
| push | 500 req/s | なし | — |
| broadcast | **1 req/s** | なし | v0.1.2 |
| multicast | 500 req/s、**1リクエストあたり最大 500 userID** | 上限なし送信 → 400 | v0.1.2 |
| reply token | **有効期限 30 秒**。Webhook受信から30秒以内に送信必須 | 未記載 | v0.1.1 (ドキュメント) |
| narrowcast | 対象オーディエンス **50名以上** でないと 400 | 未チェック | v0.2.0 |
| richmenu image | サイズ: 2500×1686px または 2500×843px、**最大 1MB** | 未チェック | v0.1.2 |
| stateless token | 有効期限 **30日**。LINE 側に状態を持たない | 正しく実装済み | — |
| `group leave` | HTTP メソッドは **POST** | **DELETE を使っている (BUG)** | v0.1.1 |
| insight date | フォーマット **YYYYMMDD**。当日・未来日は不可 | 未検証 | v0.1.2 |
| audience create | type: `UPLOAD`(UID), `CLICK`, `IMP` など | type 未検証 | v0.2.0 |

---

## Auth Flow

### longterm（デフォルト）

```
1. lm auth login
   → prompt: Channel ID
   → prompt: Long-term Channel Access Token (masked)
2. tokens.yaml に {token, token_type: "longterm"} を保存 (expires_at は空)
3. EnsureToken(): tokens.yaml から返す。IsValid は token != "" で判定
4. 自動更新なし。期限切れは LINE 側で HTTP 401 として返る
```

### stateless（CI/CD 推奨）

```
1. lm auth login --type stateless
   → prompt: Channel ID
   → prompt: Channel Secret (masked)
   → POST api.line.me/oauth2/v2.1/token (検証を兼ねる)
2. tokens.yaml に {token, expires_at, token_type: "stateless"} を保存
   credentials.yaml に {channel_secret} を保存 (0600)
3. EnsureToken():
   a. LM_TOKEN 環境変数 → そのまま返す (全 auth バイパス)
   b. tokens.yaml が IsValid (expires_at まで 5分以上残) → そのまま返す
   c. stateless: POST /oauth2/v2.1/token で再発行 → tokens.yaml 更新
   d. 上記すべて失敗 → AuthError (exit 2): "no token found, run 'lm auth login'"
```

### Token Resolution 優先順位

```
LM_TOKEN env  >  tokens.yaml cache  >  stateless 再発行  >  error
```

---

## Output Contract

| フォーマット | stdout | stderr |
|------------|--------|--------|
| table | タブ区切りテーブル（ヘッダー大文字） | 進捗・確認・エラー |
| json | インデント付き JSON のみ | 進捗・確認・エラー |
| yaml | YAML のみ | 進捗・確認・エラー |
| csv | RFC 4180 CSV（ヘッダー小文字） | 進捗・確認・エラー |

- **stdout は純粋なデータ専用**。人間向けメッセージは必ず stderr
- `--quiet` でも **エラーは stderr に出力する**（抑制しない）
- `--no-input` 時、確認プロンプトが必要な操作は exit 4 で失敗する
- バイナリデータ（`content get`）は stdout に直接書き出す。`--format` は無視

---

## Error Contract

| Exit Code | 定数 | 条件 |
|-----------|------|------|
| 0 | ExitOK | 成功 |
| 1 | ExitGeneral | 未分類、config YAML parse エラー |
| 2 | ExitAuth | 401/403、トークン未設定 |
| 3 | ExitNotFound | 404 |
| 4 | ExitValidation | 引数不正、--no-input で必須入力欠如 |
| 5 | ExitAPI | LINE API エラー（401/403/404/429 以外の 4xx/5xx） |
| 6 | ExitNetwork | タイムアウト、接続拒否 |
| 10 | ExitCancelled | Ctrl+C 割り込み |
| 11 | ExitRateLimit | 429 Too Many Requests |

**エラーメッセージの書式**（stderr）:
```
Error: {型}: {human-readable message}
```

例:
```
Error: auth error: no token found, run 'lm auth login'
Error: API error (HTTP 400): The request body has 1 error(s) (messages[0].text: May not be empty)
Error: network error: Post "https://api.line.me/...": context deadline exceeded
```

---

## Configuration

```
~/.config/lm/         (LM_CONFIG_DIR で変更可)
  config.yaml         0644  プロファイル・デフォルト設定
  credentials.yaml    0600  Channel Secret (channel_secret)
  tokens.yaml         0600  access token + expires_at + token_type
```

### 優先順位（低 → 高）

```
config.yaml defaults  <  環境変数  <  CLI フラグ
```

### 環境変数一覧

| 変数 | 型 | 説明 |
|------|---|------|
| `LM_TOKEN` | string | トークン直接指定。全 auth 処理をバイパス |
| `LM_CHANNEL_ID` | string | Channel ID。`auth login` プロンプトを省略 |
| `LM_CHANNEL_SECRET` | string | Channel Secret。`auth login --type stateless` プロンプトを省略 |
| `LM_PROFILE` | string | 使用プロファイル名 |
| `LM_FORMAT` | string | 出力フォーマット: table/json/yaml/csv |
| `LM_CONFIG_DIR` | string | 設定ディレクトリパス |
| `LM_NO_INPUT` | 1/true | 非対話モード。プロンプトが必要な場面で exit 4 |
| `LM_ENDPOINT` | string | API ベース URL 上書き（テスト用）。**HTTPS のみ許可** |
| `LM_DEBUG` | 1/true/api | デバッグログ: verbose(1/true) または headers+body(api) |

---

## Version History

| Version | Date | Summary |
|---------|------|---------|
| 0.1.2 | TBD | Reliability: broadcast confirm, URL encoding, error context, auth verify |
| 0.1.1 | TBD | Critical bug fixes: group leave, alias create, quiet flag, content error |
| 0.1.0 | 2026-03-20 | Initial implementation |

---

## 0.1.1 Changes (Critical Bug Fixes)

### 1. `group leave`: DELETE → POST に修正

**Background**:
LINE Messaging API の `leave` エンドポイントは `POST /v2/bot/group/{groupId}/leave` を要求する。
現在の実装は `DELETE` を使用しており、実際の API に対して HTTP 405 Method Not Allowed が返る。
`group leave` は機能として完全に壊れている状態。

**Current**:
```go
// internal/api/group.go:32
func (a *GroupAPI) Leave(groupID string) error {
    url := fmt.Sprintf("%s/v2/bot/group/%s/leave", a.Client.BaseURL, groupID)
    return a.Client.Delete(url)  // ← 誤り: HTTP 405
}
```

**After**:
```go
func (a *GroupAPI) Leave(groupID string) error {
    url := fmt.Sprintf("%s/v2/bot/group/%s/leave", a.Client.BaseURL, groupID)
    _, err := a.Client.Post(url, nil, nil)
    return err
}
```

**Test**:
```go
// internal/api/group_test.go
func TestGroupAPI_Leave(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/v2/bot/group/G123/leave" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer srv.Close()
    c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
    if err := (&GroupAPI{Client: c}).Leave("G123"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/group.go` | `Leave()` を `Client.Delete` → `Client.Post(url, nil, nil)` に変更 |
| `internal/api/group_test.go` | **New** — `Leave` の HTTP メソッド検証テスト |

---

### 2. `richmenu alias create`: 空レスポンス body のデコードを除去

**Background**:
LINE API の `POST /v2/bot/richmenu/alias` は成功時に HTTP 200 + **空ボディ** を返す。
現在の実装は空ボディを `model.RichMenuAlias{}` にデコードしようとして `EOF` エラーを返す。
エイリアス作成は常に失敗している状態。

**Current**:
```go
// internal/api/richmenu.go:117
func (a *RichMenuAPI) CreateAlias(body any) (*model.RichMenuAlias, error) {
    var alias model.RichMenuAlias
    _, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/richmenu/alias", body, &alias)  // ← &alias で EOF
    if err != nil {
        return nil, err
    }
    return &alias, nil
}
```

**After**:
```go
// シグネチャも変更: 空レスポンスなので *model.RichMenuAlias は返さない
func (a *RichMenuAPI) CreateAlias(body any) error {
    _, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/richmenu/alias", body, nil)
    return err
}
```

**cmd 側の変更** (`cmd/richmenu/richmenu.go` の `aliasCreateCmd`):
```go
// Before
alias, err := rmAPI.CreateAlias(body)
fmt.Fprintf(os.Stderr, "Created alias: %s\n", alias.RichMenuAliasID)

// After
if err := rmAPI.CreateAlias(body); err != nil {
    return err
}
fmt.Fprintln(os.Stderr, "Created alias")
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/richmenu.go` | `CreateAlias()` シグネチャを `error` のみに変更、nil を Post に渡す |
| `cmd/richmenu/richmenu.go` | `aliasCreateCmd` の戻り値対応を修正 |
| `internal/api/richmenu_test.go` | **New** — `CreateAlias` テスト追加 |

---

### 3. `content get`: NetworkError ラッピング漏れを修正

**Background**:
`ContentAPI.Get` はネットワークエラーを素の `error` で返すため、exit code が ExitNetwork(6) ではなく ExitGeneral(1) になる。
他のすべての API メソッドは `&lmerrors.NetworkError{Err: err}` でラップしているのに、content だけ例外となっている。

**Current**:
```go
// internal/api/content.go:28
resp, err := a.Client.HTTP.Do(req)
if err != nil {
    return nil, err  // ← NetworkError でラップされていない
}
```

**After**:
```go
resp, err := a.Client.HTTP.Do(req)
if err != nil {
    return nil, &lmerrors.NetworkError{Err: err}
}
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/content.go` | ネットワークエラーを `&lmerrors.NetworkError{Err: err}` でラップ |

---

### 4. `--quiet` フラグ: persistent flag の正しい読み取りに修正

**Background**:
`--quiet` は `cmd/root.go` で `PersistentFlags()` に定義されているグローバルフラグ。
しかし `cmd/message/message.go` の `isQuiet()` は `cmd.Flags().GetBool("quiet")` を使っている。
`cmd.Flags()` はそのコマンド自身に定義されたフラグのみを返し、親から継承した persistent flag は含まれない。
結果として `lm --quiet message push ...` としても quiet モードが効かない。

**Current**:
```go
// cmd/message/message.go:281
func isQuiet(cmd *cobra.Command) bool {
    quiet, _ := cmd.Flags().GetBool("quiet")  // ← persistent flag を読めない
    return quiet
}
```

**After**:
```go
func isQuiet(cmd *cobra.Command) bool {
    // InheritedFlags() は親の PersistentFlags を含む
    quiet, _ := cmd.InheritedFlags().GetBool("quiet")
    if !quiet {
        // ローカルに定義されている場合のフォールバック
        quiet, _ = cmd.Flags().GetBool("quiet")
    }
    return quiet
}
```

あるいは `cmd/root.go` の `IsQuiet()` を使用:
```go
// cmd/root.go
func IsQuiet() bool { return flagQuiet }

// cmd/message/message.go
import "github.com/crowdy/lm-cli/cmd"
func isQuiet(_ *cobra.Command) bool { return cmd.IsQuiet() }
```

> **Note**: root パッケージへの循環 import を避けるため、`cmdutil` に `IsQuiet(cmd *cobra.Command) bool` を追加する方式が望ましい。

**Modified files**:

| File | Change |
|------|--------|
| `cmd/cmdutil/output.go` | `IsQuiet(cmd *cobra.Command) bool` を追加 |
| `cmd/message/message.go` | `isQuiet()` → `cmdutil.IsQuiet(cmd)` に置き換え |

---

### 5. `reply` コマンド: reply token の有効期限を警告表示

**Background**:
LINE Messaging API の reply token は **Webhook イベント受信から 30 秒以内** にしか使用できない。
期限切れの reply token で API を呼ぶと HTTP 400 が返る。
CLI 利用者（特に初心者）はこの制約を知らず、ターミナルで token をコピペして試すと必ず失敗する。
エラーメッセージだけでは原因がわかりにくいため、コマンド実行時に警告を出す。

**After**:
```go
// cmd/message/message.go の replyCmd
RunE: func(cmd *cobra.Command, args []string) error {
    if !cmdutil.IsQuiet(cmd) {
        fmt.Fprintln(os.Stderr, "Note: reply token is valid for 30 seconds from the webhook event.")
    }
    // ... 以下既存処理
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/message/message.go` | `replyCmd` に 30秒有効期限の警告を stderr 追加 |

---

## 0.1.2 Changes (Reliability & UX)

### 1. `broadcast` / `narrowcast`: 送信前確認プロンプト

**Background**:
`lm message broadcast "セール開催中！"` を実行すると、確認なしで全フォロワーにメッセージが送信される。
誤送信は取り消しができない。フォロワー数万人のアカウントでは深刻な事故につながる。
`--no-input` や `--quiet` での自動化用途ではプロンプトをスキップする設計とする。

**After UX**（TTY 上でのインタラクティブ実行）:
```
$ lm message broadcast "セール開催中！"
Broadcast to ALL followers. This cannot be undone.
Proceed? [y/N]: y
Broadcasted 1 message(s)
```

**`--no-input` または TTY なし時**:
```
$ LM_NO_INPUT=1 lm message broadcast "CI通知"
(確認なしで送信)
```

**設計**:
- `--no-input` (`LM_NO_INPUT=1`) の場合 → 確認なしで実行（CI/CD 用途）
- `--force` フラグで確認をスキップ（スクリプトから TTY なしで呼ぶ場合のエスケープハッチ）
- TTY でない場合（パイプ経由）かつ `--no-input` でも `--force` でもない場合 → exit 4: "confirmation required; use --force or --no-input to bypass"

```go
// cmd/message/message.go の broadcastCmd に追加
broadcastCmd.Flags().Bool("force", false, "skip confirmation prompt")
```

```go
// broadcast の確認ロジック
if !config.IsNoInput() && !forceFlag {
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return &lmerrors.ValidationError{
            Message: "broadcast requires confirmation; use --force or LM_NO_INPUT=1 to bypass",
        }
    }
    confirmed, err := prompt.Confirm("Broadcast to ALL followers. This cannot be undone.\nProceed?")
    if err != nil || !confirmed {
        return &lmerrors.CancelledError{}
    }
}
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/message/message.go` | `broadcastCmd`/`narrowcastCmd` に確認プロンプトを追加、`--force` フラグ追加 |
| `internal/errors/errors.go` | `CancelledError` 型を追加（ExitCancelled=10） |

---

### 2. Multicast: 500 userID 上限の自動バッチ処理

**Background**:
LINE API の `/v2/bot/message/multicast` は 1 リクエストあたり最大 500 userID を受け付ける。
現在は上限チェックなしで送信するため、501名以上では HTTP 400 エラーが返る。
大量送信の用途（例: --to-file で 10,000名のファイル指定）を想定し、自動分割する。

**After**:
```go
// internal/api/message.go
const multicastBatchSize = 500

func (a *MessageAPI) Multicast(to []string, messages []any) (*model.MessageResponse, error) {
    if len(to) <= multicastBatchSize {
        // 既存処理（単一リクエスト）
    }
    // 500件ずつバッチ送信
    var combined model.MessageResponse
    for i := 0; i < len(to); i += multicastBatchSize {
        end := i + multicastBatchSize
        if end > len(to) {
            end = len(to)
        }
        resp, err := a.multicastBatch(to[i:end], messages)
        if err != nil {
            return nil, fmt.Errorf("batch %d/%d failed: %w", i/multicastBatchSize+1, (len(to)+499)/500, err)
        }
        combined.SentMessages = append(combined.SentMessages, resp.SentMessages...)
    }
    return &combined, nil
}
```

**バッチ送信時の stderr 出力**（`--quiet` でない場合）:
```
Sending batch 1/20 (500 users)...
Sending batch 2/20 (500 users)...
...
Multicasted 1 message(s) to 10000 users (20 batches)
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/message.go` | `Multicast()` に 500件バッチ分割を実装 |
| `cmd/message/message.go` | バッチ数を stderr に表示 |
| `internal/api/message_test.go` | バッチ分割テスト追加（501件 → 2リクエスト確認） |

---

### 3. URL パラメータのエンコーディング修正

**Background**:
`insight`, `user followers` などの URL 構築で、ユーザー入力を直接文字列結合している。
`&`, `=`, `+`, スペースを含む値で URL が壊れる。`net/url.Values` を使用する。

**Current**:
```go
// internal/api/insight.go:16
url += "?date=" + date                           // injection 可能
url += fmt.Sprintf("&type=%s", msgType)          // injection 可能
```

**After**:
```go
import "net/url"

func (a *InsightAPI) GetDelivery(msgType, date string) (*model.DeliveryStats, error) {
    params := url.Values{}
    params.Set("type", msgType)
    if date != "" {
        params.Set("date", date)
    }
    endpoint := a.Client.BaseURL + "/v2/bot/insight/message/delivery?" + params.Encode()
    var resp model.DeliveryStats
    return &resp, a.Client.Get(endpoint, &resp)
}
```

同様の修正対象: `InsightAPI.GetFollowers`, `UserAPI.GetFollowers`.

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/insight.go` | `url.Values` でクエリ文字列構築 |
| `internal/api/user.go` | `url.Values` でクエリ文字列構築 |

---

### 4. Insight date フォーマット検証

**Background**:
LINE Insight API は日付を `YYYYMMDD` 形式で要求する。`2024-01-15` や `yesterday` のような入力では
LINE API が HTTP 400 を返すが、エラーメッセージは「日付フォーマットが違う」とは言ってくれない。
CLI 側で事前に検証し、わかりやすいエラーを返す。

**After**:
```go
// cmd/insight/insight.go 内のヘルパー
func validateDate(date string) error {
    if date == "" {
        return nil  // デフォルト（前日）をLINE APIに任せる
    }
    if len(date) != 8 {
        return &lmerrors.ValidationError{
            Field:   "date",
            Message: fmt.Sprintf("must be YYYYMMDD format, got %q", date),
        }
    }
    if _, err := time.Parse("20060102", date); err != nil {
        return &lmerrors.ValidationError{
            Field:   "date",
            Message: fmt.Sprintf("invalid date %q (expected YYYYMMDD)", date),
        }
    }
    return nil
}
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/insight/insight.go` | `validateDate()` ヘルパー追加、`followersCmd` / `deliveryCmd` で呼び出し |

---

### 5. Rich menu 画像: サイズ・フォーマット検証

**Background**:
LINE API は richmenu 画像に厳格なサイズ制約を課す。違反するとアップロードは HTTP 400 になるが
エラーメッセージが不親切。CLI 側でアップロード前にファイルを検査する。

**制約**:
- フォーマット: JPEG または PNG のみ
- サイズ: **2500×1686px** (フルサイズ) または **2500×843px** (ハーフサイズ)
- ファイルサイズ: **最大 1MB**

**After**:
```go
// internal/api/richmenu.go の UploadImage() に追加
const maxImageSize = 1 * 1024 * 1024  // 1MB

if len(data) > maxImageSize {
    return &lmerrors.ValidationError{
        Field:   "image",
        Message: fmt.Sprintf("file size %d bytes exceeds 1MB limit", len(data)),
    }
}
```

> **Note**: 画像の pixel サイズ検証は `image.DecodeConfig` で可能だが、JPEG/PNG デコード依存が増えるため
> v0.2.0 以降の追加機能とする。v0.1.2 ではファイルサイズのみ検証する。

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/richmenu.go` | `UploadImage()` に 1MB ファイルサイズ検証を追加 |

---

### 6. `auth login`: ログイン直後にトークンを検証

**Background**:
longterm トークンを入力した直後、そのトークンが有効かどうかは確認していない。
誤ったトークンをペーストした場合、次の API 呼び出しまでエラーに気づかない。
ログイン直後に `GET /v2/bot/info` を呼んで検証し、即座にフィードバックを与える。

**After**（longterm / stateless 共通）:
```go
// cmd/auth/auth.go の loginLongterm(), loginStateless() の末尾
fmt.Fprintln(os.Stderr, "Verifying token...")
client := api.NewClient(token)
botAPI := &api.BotAPI{Client: client}
info, err := botAPI.GetInfo()
if err != nil {
    // トークン保存は完了しているが、検証失敗を警告
    fmt.Fprintf(os.Stderr, "Warning: token saved but verification failed: %v\n", err)
    fmt.Fprintf(os.Stderr, "Run 'lm bot info' to test your token manually.\n")
    return nil  // exit 0 を維持（保存は成功しているため）
}
fmt.Fprintf(os.Stderr, "Verified as @%s (%s)\n", info.BasicID, info.DisplayName)
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/auth/auth.go` | `loginLongterm()` / `loginStateless()` 末尾に Bot 情報取得による検証追加 |

---

### 7. `Retry-After` ヘッダーのパース修正

**Background**:
現在の実装は `time.ParseDuration(retryAfter + "s")` を使っている。
HTTP の `Retry-After` ヘッダーは秒数の整数文字列（例: `"60"`）であり、Go の duration 文字列ではない。
`"60" + "s" = "60s"` は偶然動作するが、意図が不明瞭で RFC 7231 の日付形式に対応していない。

**After**:
```go
// internal/api/client.go
import "strconv"

func parseRetryAfter(header string) time.Duration {
    if n, err := strconv.Atoi(header); err == nil {
        return time.Duration(n) * time.Second
    }
    // RFC 7231 HTTP-date 形式のフォールバック
    if t, err := http.ParseTime(header); err == nil {
        d := time.Until(t)
        if d > 0 {
            return d
        }
    }
    return 0  // 0 = caller がデフォルト値を使用
}
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/client.go` | `parseRetryAfter()` ヘルパーを追加、retry loop で使用 |
| `internal/api/client_test.go` | `parseRetryAfter()` のテスト追加 |

---

### 8. `LM_ENDPOINT` への HTTP URL を拒否

**Background**:
`LM_ENDPOINT=http://...` を設定すると、Bearer Token が平文 HTTP で送信される。
意図しない設定ミスによるトークン漏洩を防ぐ。

**After**:
```go
// internal/api/client.go の NewClient()
if ep := os.Getenv(config.EnvEndpoint); ep != "" {
    if !strings.HasPrefix(ep, "https://") {
        // テスト環境 (httptest) でのみ http:// を許容するフラグ
        if os.Getenv("LM_ALLOW_HTTP") != "1" {
            panic(fmt.Sprintf("LM_ENDPOINT must start with https://, got: %s", ep))
        }
    }
    baseURL = ep
}
```

> `LM_ALLOW_HTTP=1` はテストコード用の非公式フラグ。README には記載しない。

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/client.go` | `NewClient()` で `LM_ENDPOINT` の https 強制チェックを追加 |
| `internal/api/client_test.go` | `LM_ALLOW_HTTP=1` でテスト通過できるよう `t.Setenv` を使用 |

---

## 0.2.0 Changes (Feature Additions)

### 1. Flex Message タイプのサポート

**Background**:
Flex Message は LINE Messaging API で最も広く使われているメッセージタイプ。
通知カード・注文確認・ダッシュボードなど、視覚的にリッチなメッセージを組める。
現在 `--type text|sticker|image` のみサポートしており、Flex を送るには `--file` 経由の JSON 直接指定が必要。
`--type flex --file bubble.json` という形で第一クラスサポートする。

**After**:
```bash
# bubble.json を Flex Message として push
lm message push <userId> --type flex --file bubble.json

# carousel (複数 bubble) も同様
lm message push <userId> --type flex --file carousel.json
```

**bubble.json の例**（LINE Flex Message Simulator でも生成可能）:
```json
{
  "type": "bubble",
  "body": {
    "type": "box",
    "layout": "vertical",
    "contents": [
      {"type": "text", "text": "Hello, World!"}
    ]
  }
}
```

**設計**:
- `--type flex` 指定時、`--file` が必須
- `--file` の中身を `{"type": "flex", "altText": "...", "contents": <file内容>}` にラップ
- `altText` は `--alt-text` フラグで指定（デフォルト: `"Flex Message"`）

```go
// cmd/message/message.go の buildMessages()
case "flex":
    if fileFlag == "" {
        return nil, &lmerrors.ValidationError{Field: "file", Message: "--file is required for --type flex"}
    }
    var contents any
    if err := api.ParseJSONFile(fileFlag, &contents); err != nil {
        return nil, err
    }
    altText, _ := cmd.Flags().GetString("alt-text")
    if altText == "" {
        altText = "Flex Message"
    }
    return []any{map[string]any{
        "type":     "flex",
        "altText":  altText,
        "contents": contents,
    }}, nil
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/message/message.go` | `buildMessages()` に `flex` ケース追加、`--alt-text` フラグ追加 |

---

### 2. `user followers --all`: 全フォロワーの自動ページネーション

**Background**:
LINE API の `GET /v2/bot/followers/ids` は最大 1000件を返し、`next` トークンで続きを取得する。
現在の CLI は 1ページのみ取得し、`(next: xxx)` と stderr に出すだけ。
全フォロワーを取得したい場合、スクリプトでループを組む必要がある。

**After**:
```bash
# 1ページのみ（既存動作）
lm user followers

# 全ページを自動ページネーション
lm user followers --all

# --all + --format json: {"userIds": [...全ID...]}
lm user followers --all --format json
```

**設計**:
```go
// cmd/user/user.go
followersCmd.Flags().Bool("all", false, "fetch all followers with auto-pagination")
```

```go
if allFlag {
    var allIDs []string
    cursor := start
    for {
        resp, err := userAPI.GetFollowers(0, cursor)
        if err != nil {
            return err
        }
        allIDs = append(allIDs, resp.UserIDs...)
        if resp.Next == "" {
            break
        }
        cursor = resp.Next
        if !isQuiet(cmd) {
            fmt.Fprintf(os.Stderr, "Fetched %d followers...\r", len(allIDs))
        }
    }
    fmt.Fprintln(os.Stderr)  // 改行
    // allIDs を出力
}
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/user/user.go` | `--all` フラグ追加、自動ページネーションループ実装 |

---

### 3. `lm status`: API 接続確認コマンドの追加

**Background**:
CI/CD パイプラインや監視スクリプトで、LINE API への疎通確認をしたい。
`lm bot info` でも代替可能だが、意図が明確な専用コマンドがあるべき。
`lm status` は bot info を呼んで成功/失敗を返す。

**After**:
```bash
$ lm status
API:     ok
Bot:     @mybotid (My Bot Name)
Token:   valid (expires in 29d 23h)

$ lm status --format json
{
  "api": "ok",
  "bot_id": "@mybotid",
  "display_name": "My Bot Name",
  "token_type": "stateless",
  "token_expires_at": "2026-04-19T10:00:00Z"
}
```

**エラー時** (exit 2):
```bash
$ lm status
Error: auth error: request failed with status 401
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/status/status.go` | **New** — `lm status` コマンド実装 |
| `cmd/root.go` | `status.Cmd` を `AddCommand` |

---

### 4. `bot consumption` の quota との比較表示

**Background**:
現在 `lm bot consumption` は `totalUsage: 42` のような数値だけ返す。
`lm bot quota` の上限値と合わせてみないと「どのくらい使ったか」がわからない。
2つを組み合わせた表示を `bot usage` サブコマンドとして提供する。

**After**:
```bash
$ lm bot usage
TYPE     LIMIT  USED   REMAINING  USAGE
limited  500    42     458        8.4%

$ lm bot usage --format json
{
  "type": "limited",
  "limit": 500,
  "used": 42,
  "remaining": 458,
  "usage_pct": 8.4
}
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/bot/bot.go` | `usageCmd` サブコマンド追加 |
| `internal/model/bot.go` | `BotUsageRow` 型追加 |

---

### 5. `audience list`: ページネーション対応

**Background**:
LINE API の `GET /v2/bot/audienceGroup/list` はページネーションをサポートし、
`hasNextPage: true` の場合は `page` パラメータで続きを取得できる。
現在は 1ページ目のみ返す。

**After**:
```bash
lm audience list                # 1ページ（デフォルト）
lm audience list --page 2       # 指定ページ
lm audience list --all          # 全ページ自動取得
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/audience/audience.go` | `--page`, `--all` フラグ追加 |
| `internal/api/audience.go` | `List()` に `page int` パラメータ追加 |

---

## 0.3.0 Changes (Advanced Features)

### 1. `auth login --type v2`: JWT アサーションによる v2.1 チャネルアクセストークン

**Background**:
LINE の v2.1 チャネルアクセストークン（JWT assertion 方式）はプロダクション推奨の認証方式。
RSA 秘密鍵で署名した JWT を発行し、それをシークレットの代わりに使う。
secrets の漏洩リスクが stateless よりも低く、発行数を LINE コンソールで管理できる。

**設定**:
```yaml
# credentials.yaml
profiles:
  prod:
    private_key_file: ~/.config/lm/prod_private.pem
```

**After**:
```bash
lm auth login --type v2
# → Channel ID プロンプト
# → Private Key ファイルパスプロンプト
# → JWT assertion 生成 → POST /oauth2/v2.1/token
```

**JWT assertion 形式**:
```json
{
  "iss": "<Channel ID>",
  "sub": "<Channel ID>",
  "aud": "https://api.line.biz/",
  "exp": <now + 30min>,
  "token_exp": 2592000
}
```

**Modified files**:

| File | Change |
|------|--------|
| `internal/api/auth.go` | `IssueV2Token(channelID, privateKeyPath string)` 関数追加 |
| `cmd/auth/auth.go` | `loginV2()` 実装、`--type v2` 対応 |
| `internal/config/credentials.go` | `PrivateKeyFile` フィールド使用 |
| `go.mod` | RSA/JWT ライブラリ追加（`github.com/golang-jwt/jwt/v5`） |

---

### 2. `auth token --check`: 有効性確認（標準出力なし）

**Background**:
スクリプトでトークンが有効かどうかを確認したい場合、`lm auth token` は値を stdout に出力してしまう。
有効性のみを exit code で返すコマンドが必要。

**After**:
```bash
# 有効なら exit 0、無効なら exit 2
if lm auth token --check; then
    echo "Token is valid"
else
    lm auth login
fi
```

**実装**:
```go
tokenCmd.Flags().Bool("check", false, "exit 0 if token is valid, exit 2 if not (no output)")
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/auth/auth.go` | `tokenCmd` に `--check` フラグ追加 |

---

### 3. `config validate`: 設定ファイルの整合性チェック

**Background**:
`config.yaml` / `credentials.yaml` / `tokens.yaml` が手動編集で壊れた場合、
最初の API 呼び出し時まで気づかない。整合性を事前確認できるコマンドを提供する。

**After**:
```bash
$ lm config validate
config.yaml:        ok (1 profile)
credentials.yaml:   ok
tokens.yaml:        ok (default: valid, staging: expired)
Active profile:     default

$ lm config validate
config.yaml:        ok
credentials.yaml:   error: YAML parse failed at line 3: mapping key is not a string
```

**Modified files**:

| File | Change |
|------|--------|
| `cmd/config/config.go` | `validateCmd` サブコマンド追加 |

---

## Appendix A: 既知の制限事項

| 制限 | 説明 | 回避策 |
|------|------|--------|
| Webhook 受信 | CLI は Webhook を受信できない。送信のみ | ngrok + webhook receiver スクリプトを別途用意 |
| LINE Notify | 別 API ・別トークン。本ツールでは非対応 | curl で直接呼ぶ |
| 予約送信 | LINE の scheduled message API は非対応 | LINE Official Account Manager を使用 |
| Flex Message Builder | GUI ビルダーなし | LINE Flex Message Simulator を使用 |
| 既読取り消し | LINE API 自体が非対応 | — |

---

## Appendix B: テスト戦略

### Unit Tests
- table-driven, `t.Run()` で分割
- HTTP: `httptest.NewServer()` でモックサーバー
- ファイル: `t.TempDir()` で一時ファイル

### 必須テスト一覧

| テスト | ファイル | 確認内容 |
|--------|---------|---------|
| GroupAPI.Leave が POST を送ること | `group_test.go` | HTTP メソッド |
| RichMenuAPI.CreateAlias が空 body でエラーにならないこと | `richmenu_test.go` | EOF エラーなし |
| EnsureToken の優先順位 | `auth_test.go` | env > cache > 再発行 > error |
| multicast 501件が2バッチに分割されること | `message_test.go` | リクエスト回数 |
| buildMessages --type flex が正しい構造を返すこと | `message_test.go` | JSON 構造 |
| parseRetryAfter("60") → 60s | `client_test.go` | duration 値 |
| LM_ENDPOINT に http:// を拒否すること | `client_test.go` | panic/error |
| insight date "2024-01-01" が ValidationError | `insight_test.go` | エラー型 |
| richmenu upload 2MB ファイルが ValidationError | `richmenu_test.go` | エラー型 |

### Integration Tests
```go
//go:build integration
// 実 LINE API に対して実行。LM_TOKEN + LM_CHANNEL_ID が必要。
```

---

## Appendix C: LINE API レート制限リファレンス

| API | 上限 | 超過時 |
|-----|------|--------|
| push | 500 req/s | 429 |
| reply | 500 req/s | 429 |
| multicast | 500 req/s | 429 |
| broadcast | **1 req/s** | 429 |
| narrowcast | 1 req/s | 429 |
| bot info | 制限なし（事実上） | — |
| follower IDs | 制限なし（事実上） | — |
| richmenu list | 制限なし（事実上） | — |

> **broadcast の 1 req/s 制限**は見落とされやすい。CI/CD で複数環境に同時に broadcast しようとするとレート制限に引っかかる。
> `lm message broadcast` を複数プロファイルで並列実行する場合は `sleep 1` を挟むこと。
