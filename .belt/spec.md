# 仕様: Gate - 汎用認証認可サーバー (v1)

## 技術コンテキスト

Go 言語で構築する汎用認証認可サーバー。Clean Architecture に基づく 4 レイヤー構成 (`domain` / `usecase` / `handler` / `infra`)。HTTP フレームワークは Chi (`net/http` 互換)、データストアは PostgreSQL (主ストア) + Redis (セッション/キャッシュ)。OAuth 2.0 / OpenID Connect 対応。v1 では WebAuthn と Device Authorization Grant を除外し、コア機能に集中する。

### 技術選定

| カテゴリ              | 選定                                                                  |
| --------------------- | --------------------------------------------------------------------- |
| HTTP ルーター         | Chi (`github.com/go-chi/chi/v5`)                                      |
| RDB                   | PostgreSQL (`pgx/v5` + `sqlx`)                                        |
| キャッシュ/セッション | Redis (`go-redis/v9`)                                                 |
| JWT                   | `github.com/golang-jwt/jwt/v5` (ES256 デフォルト、RS256 選択可)       |
| パスワードハッシュ    | argon2id (`golang.org/x/crypto/argon2`) + `crypto/rand`、PHC 形式保存 |
| TOTP                  | `github.com/pquerna/otp`                                              |
| マイグレーション      | `github.com/golang-migrate/migrate/v4`                                |
| バリデーション        | `github.com/go-playground/validator/v10`                              |
| 設定                  | `github.com/caarlos0/env/v11` (環境変数バインド)                      |
| ログ                  | `log/slog` (標準ライブラリ)                                           |
| メトリクス            | `github.com/prometheus/client_golang`                                 |

### アーキテクチャ

```
cmd/gate/main.go           エントリポイント、手動 DI
internal/domain/            エンティティ、リポジトリインターフェース
internal/usecase/           ビジネスロジック
internal/handler/           HTTP ハンドラ
internal/middleware/         ミドルウェア
internal/infra/postgres/    PostgreSQL リポジトリ実装
internal/infra/redis/       Redis セッション/キャッシュ/レートリミット
internal/infra/crypto/      argon2id、JWT、乱数生成
internal/infra/mailer/      メーラーインターフェース + stdout 実装
internal/config/            設定構造体
migrations/                 SQL マイグレーション
```

## 機能要件

### 認証 (Authentication)

- [x] メールアドレス + パスワードによるユーザー登録 (`POST /api/v1/auth/register`)
- [x] メールアドレスの確認 (検証トークン発行、有効期限 24 時間)
  - 有効期限は設定可能にしてほしい
- [x] 検証メール再送 (1 時間に 3 回まで)
  - 制限は設定可能にしてほしい
- [x] パスワードハッシュ (argon2id: time=1, memory=64MB, parallelism=4、PHC 形式)
- [x] パスワード同時ハッシュ計算のセマフォ制御 (デフォルト上限 16 並列 = 1GB)
- [x] パスワード変更 (認証必須、変更後に全リフレッシュトークン失効)
- [x] パスワードリセット (メールによるリセットトークン発行、有効期限 1 時間)
- [x] パスワード強度ポリシー (最小 8 文字、最大 128 文字)
  - 最小・最大は設定可能にしてほしい
- [x] メールアドレス + パスワードによるログイン (`POST /api/v1/auth/login`)
- [x] ログアウト (セッション破棄 + リフレッシュトークン失効)
- [x] セッション管理 (Redis サーバーサイドセッション)
- [x] 同時セッション数の制限 (デフォルト 5、超過時は最古を削除)
  - 制限は設定可能にしてほしい
- [x] アカウントの有効化・無効化 (status: `unverified`, `active`, `locked`, `deleted`)
- [x] アカウント論理削除 (`deleted_at` タイムスタンプ)
  - 論理削除 + 自動物理削除 (例: 30 日後に完全削除)
- [x] ユーザー列挙防止 (存在しないメールへのリセット要求でも 200 OK を返す)

### 多要素認証 (MFA)

- [x] TOTP セットアップ (`POST /api/v1/mfa/totp/setup` → QR コード用 URI 返却)
- [x] TOTP セットアップ確認 (`POST /api/v1/mfa/totp/confirm` → 確認コード入力で有効化)
- [x] TOTP 検証 (ログイン時、前後 1 ステップの時刻ずれを許容)
- [x] TOTP 無効化 (`DELETE /api/v1/mfa/totp`)
- [x] リカバリーコード生成 (8 個、8 文字英数字、ハッシュ化保存)
- [x] リカバリーコード再生成 (`POST /api/v1/mfa/recovery-codes/regenerate`)
- [x] MFA 必要時の中間レスポンス (`mfa_required` + `mfa_token`)

### ソーシャルログイン

- [x] Google OAuth2/OIDC 連携
- [x] GitHub OAuth2 連携
- [x] Apple Sign In 連携
- [x] 汎用 OIDC プロバイダコネクタ (プロバイダ固有コードを最小化)
  - 基本これをベースに各プロバイダの実装を行うことが出来る形にしたい。Google / GitHub / Apple は最初のサンプル実装として実装したい
- [x] ソーシャルログイン開始 (`GET /api/v1/auth/social/{provider}/authorize`)
- [x] コールバック処理 (`GET /api/v1/auth/social/{provider}/callback`)
- [x] プロバイダの `sub` (一意識別子) でアカウント紐付け

### 認可 (Authorization)

- [x] ロールの作成・更新・削除 (Admin API)
- [x] ユーザーへのロール割り当て
- [x] ロールの階層構造 (親ロール継承、最大深度 10)
- [x] 循環参照の検出と拒否 (再帰 CTE で検出、400 Bad Request)
- [x] パーミッションの定義 (例: `users:read`, `users:write`)
- [x] ロールへのパーミッション割り当て
- [x] ユーザー単位のパーミッション上書き (allow-only、v1)
- [x] パーミッション解決 (ユーザー直接 + ロール由来の和集合)
- [x] 解決済みパーミッションの Redis キャッシュ (TTL: 5 分)

### トークン管理

- [x] JWT アクセストークン発行 (ES256 デフォルト、RS256 設定可)
- [x] カスタムクレームの付与 (上限 20 個、値サイズ上限 1KB)
  - クレームのキーと値のサイズ上限は設定可能にしてほしい
- [x] アクセストークン有効期限 (デフォルト 15 分、最小 1 分、最大 1 時間)
- [x] リフレッシュトークン発行 (Opaque 形式、SHA-256 ハッシュで DB 保存)
- [x] リフレッシュトークンローテーション (使用時に新トークン発行、旧トークン失効)
- [x] リフレッシュトークン有効期限 (デフォルト 7 日、最小 1 時間、最大 90 日)
  - リフレッシュトークンの有効期限は設定可能にしてほしい
- [x] トークンファミリー検出 (失効済みトークンの再利用でファミリー全失効)
- [x] グレースピリオド (ローテーション後 10 秒間は旧トークンも許容)
  - グレースピリオドの長さは設定可能にしてほしい
- [x] 個別トークンの失効 (`POST /oauth/revoke`)
- [x] ユーザー単位の全トークン失効 (`tokens_revoked_at` タイムスタンプ比較)
- [x] クライアント単位の全トークン失効

### OAuth 2.0

- [x] Authorization Code + PKCE グラント (`GET /oauth/authorize`, `POST /oauth/token`)
- [x] PKCE 必須 (public クライアント)、`S256` のみ許可
- [x] Client Credentials グラント (`POST /oauth/token`)
- [x] 認可コード保存 (DB、有効期限 10 分、使い回し検出)
  - 認可コードの有効期限は設定可能にしてほしい
- [x] 認可コード再利用検出時に当該コードで発行された全トークンを失効
- [x] クライアント登録 (client_id / client_secret 発行、Admin API)
- [x] クライアント種別 (confidential / public)
- [x] クライアント認証 (`client_secret_basic`, `client_secret_post`)
- [x] リダイレクト URI の管理 (最大 10 個/クライアント)
- [x] リダイレクト URI の厳密一致検証
- [x] リダイレクト URI スキーム制限 (`https://` のみ、`http://localhost` は開発用に許可)
- [x] 許可スコープの設定 (クライアントごと)
- [x] スコープとパーミッションの連携 (スコープはパーミッションのサブセット)
- [x] トークンイントロスペクション (`POST /oauth/introspect`)

### OpenID Connect (OIDC)

- [x] ID Token 発行 (JWT、必須クレーム: `iss`, `sub`, `aud`, `exp`, `iat`, `nonce`)
- [x] ID Token 有効期限 (デフォルト 1 時間、最小 5 分、最大 24 時間)
- [x] UserInfo エンドポイント (`GET /oauth/userinfo`)
- [x] 標準クレーム (`sub`, `email`, `email_verified`, `name`, `picture`)
- [x] ディスカバリ (`GET /.well-known/openid-configuration`)
- [x] JWKS エンドポイント (`GET /.well-known/jwks.json`)
- [x] JWT 署名鍵ローテーション (JWKS に新旧複数鍵公開、`kid` で特定)
- [x] 旧鍵の保持期間 = アクセストークン最大有効期限 + 10 分バッファ

### 管理機能 (Admin API)

- [x] ユーザー一覧・詳細・更新・論理削除 (`/api/v1/admin/users`)
- [x] アカウントロック・ロック解除 (`POST /api/v1/admin/users/{id}/lock`, `/unlock`)
- [x] MFA リセット (`POST /api/v1/admin/users/{id}/reset-mfa`)
- [x] ロール CRUD (`/api/v1/admin/roles`)
- [x] パーミッション CRUD (`/api/v1/admin/permissions`)
- [x] クライアント CRUD (`/api/v1/admin/clients`)

### 監査ログ

- [x] ログイン・ログアウトの記録
- [x] 認証失敗の記録
- [x] 権限変更の記録
- [x] トークン発行・失効の記録
- [x] 管理操作の記録
- [x] 監査ログ保存先: PostgreSQL (`audit_logs` テーブル、`metadata` は JSONB)

## 非機能要件

### セキュリティ

- [x] ログイン試行のレートリミット (Sliding Window Counter、Redis Sorted Set)
- [x] エンドポイントごとのレートリミット設定 (IP ベース)
- [x] 連続失敗時のアカウントロック (デフォルト 5 回失敗で 30 分ロック)
  - アカウントロックの閾値とロック期間は設定可能にしてほしい
- [x] `Retry-After` ヘッダー付き 429 レスポンス
- [x] CORS 設定 (明示的オリジン指定のみ、ワイルドカード `*` 禁止)
- [x] CSRF 対策 (OAuth2 state パラメータ検証)
- [x] HTTPS 強制
- [x] JTI (JWT ID) によるリプレイ防止

### パフォーマンス・スケーラビリティ

- [x] アプリケーションサーバーのステートレス設計 (セッションは Redis に外部化)
- [x] 水平スケーリング対応
- [x] PostgreSQL 接続プーリング
- [x] 適切なインデックス設計 (部分ユニークインデックス等)

### 運用

- [x] ヘルスチェック (`GET /health` liveness)
- [x] レディネスチェック (`GET /ready` DB/Redis 接続確認)
- [x] Prometheus メトリクス (`GET /metrics`)
- [x] 構造化ログ (`log/slog`)
- [x] グレースフルシャットダウン (SIGINT/SIGTERM)
- [x] Docker イメージ提供
- [x] `compose.yml` (PostgreSQL + Redis + Gate)
  - docker-compose.yml は旧来の名称なので、`docker-compose.yml` ではなく `compose.yml` とする
- [x] DB マイグレーション (`golang-migrate`)
- [x] 環境変数による設定 (`caarlos0/env`)

### API 設計

- [x] REST (JSON) + OpenAPI 3.0 定義
- [x] `/api/v1/` プレフィックスによるバージョニング
- [x] 統一エラーレスポンス形式 (`{"error": {"code": "...", "message": "..."}}`)
- [x] OAuth 2.0 エラーは RFC 6749 Section 5.2 準拠

## エッジケース・リスク

### 認証系

- [x] 同時パスワード変更のレースコンディション (楽観的ロックで 409 Conflict)
- [x] メール未検証ユーザーのパスワードリセット (許可するが検証済みとはしない)
- [x] 論理削除後の同一メールでの再登録 (部分ユニークインデックスで許可)
- [x] TOTP の時刻ずれ (前後 1 ステップのみ許容)
- [x] リカバリーコード全消費 + TOTP デバイス紛失 (Admin API による MFA リセット)
- [x] ソーシャルログインプロバイダ側のメール変更 (`sub` で紐付け、メールは更新)

### トークン・OAuth2 系

- [x] リフレッシュトークン同時使用のレースコンディション (グレースピリオド 10 秒)
- [x] 認可コードのリプレイ攻撃 (再利用検出時に全関連トークン失効)
- [x] JWT 署名鍵ローテーション中のリクエスト (JWKS に新旧両鍵公開、`kid` で選択)
- [x] client_secret 漏洩 (クライアント単位の全トークン失効 + secret ローテーション API)

### データ整合性系

- [x] ロール削除時の既存ユーザーへの影響 (割り当て除去、`force=true` で確認)
- [x] パーミッション削除時のロールへの影響 (カスケード除去)
- [x] 大量トークン失効のパフォーマンス (`tokens_revoked_at` タイムスタンプ比較で即時対応)
- [x] argon2id 同時実行によるメモリ枯渇 (セマフォで 16 並列に制限)
  - 攻撃対策で対応するが、過度に厳しい制限は正当なユーザーの UX を損なう可能性があるため、デフォルトは 16 並列とし、必要に応じて調整できるようにする

## Open Questions

- [x] ソーシャルログインで同一メールアドレスの自動リンクを行うか (セキュリティ vs UX)
  - 行う
- [ ] パスワード辞書チェック (Have I Been Pwned API 連携) を v1 に含めるか
- [x] 監査ログの保持期間ポリシー (無制限 / 90 日 / 設定可能)
  - 設定可能にする
- [ ] CAPTCHA プロバイダの選定 (reCAPTCHA / hCaptcha / Turnstile、v1 はインターフェースのみ)
  - v2で選定・実装する

## v1 スコープ外 (v2 以降)

- WebAuthn / FIDO2 (パスキー)
- Device Authorization Grant (RFC 8628)
- deny パーミッション (allow + deny モデル)
- マルチテナント
- CAPTCHA 実装 (v1 はインターフェースのみ)
- SMTP メーラー実装 (v1 は stdout 実装のみ)
- Admin ダッシュボード UI
- 高度なユーザー検索 (全文検索、複合フィルタ)
