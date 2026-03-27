# ロードマップ: Gate - 汎用認証認可サーバー (v1)

## アーキテクチャ方針

Clean Architecture に基づく 4 レイヤー構成。依存方向は外側から内側へのみ。

```
cmd/gate/main.go           エントリポイント、手動 DI
internal/domain/            エンティティ、リポジトリインターフェース (外部依存なし)
internal/usecase/           ビジネスロジック (domain にのみ依存)
internal/handler/           HTTP ハンドラ (usecase, domain に依存)
internal/middleware/         ミドルウェア
internal/infra/postgres/    PostgreSQL リポジトリ実装 (domain インターフェースを実装)
internal/infra/redis/       Redis セッション/キャッシュ/レートリミット
internal/infra/crypto/      argon2id、JWT、乱数生成
internal/infra/mailer/      メーラーインターフェース + stdout 実装
internal/infra/social/      ソーシャルログインプロバイダ
internal/config/            設定構造体 (環境変数バインド)
migrations/                 SQL マイグレーション
```

技術スタック: Chi, PostgreSQL (pgx/v5 + sqlx), Redis (go-redis/v9), golang-jwt/jwt/v5 (ES256), argon2id (PHC 形式), pquerna/otp, golang-migrate, caarlos0/env, log/slog, prometheus/client_golang

## マイルストーン依存関係

```
v0.1 (Bootstrap)
  │
  v
v0.2 (Core Auth)
  │
  ├──→ v0.3 (JWT & Token) ──→ v0.5 (OAuth 2.0) ──→ v0.6 (OIDC)
  │
  ├──→ v0.4 (MFA)
  │
  ├──→ v0.7 (RBAC)
  │
  ├──→ v0.8 (Social Login)
  │
  v
v0.9 (Admin & Audit) ※ v0.3, v0.4, v0.5, v0.6, v0.7, v0.8 の全てが完了後に着手
  │
  v
v0.10 (Security & Ops) ※ v0.9 完了後
```

並行実装可能なグループ (v0.2 完了後):

- Group A: v0.3 → v0.5 → v0.6 (トークン / OAuth / OIDC)
- Group B: v0.4 (MFA)
- Group C: v0.7 (RBAC)
- Group D: v0.8 (Social Login)

**v0.9 は Group A-D の全てが完了してから着手すること。**

---

## v0.1 - Project Bootstrap

**ゴール**: プロジェクト骨格を構築し、設定読み込み・DB/Redis 接続・マイグレーション・ヘルスチェックが動作する最小サーバーを起動する
**完動品としての価値**: `compose.yml` で PostgreSQL + Redis + Gate を起動し、`GET /health` と `GET /ready` が 200 OK を返す

- [x] プロジェクト構造の整理 (ルート `main.go` を `cmd/gate/main.go` に移行、`Taskfile.yml` の `MAIN_FILE` パスを更新)
- [x] Go モジュール初期化と依存ライブラリの追加 (`go get` で Chi, pgx/v5, sqlx, go-redis/v9, caarlos0/env/v11, golang-migrate/migrate/v4, go-playground/validator/v10, prometheus/client_golang, golang-jwt/jwt/v5, pquerna/otp, golang.org/x/crypto を追加)
- [x] 設定構造体の実装 (`internal/config/config.go`: 全環境変数バインド、`Validate()` メソッドでハードリミット検証)
- [x] ドメイン層のエンティティ定義 (`internal/domain/`: User, OAuthClient, Role, Permission, RefreshToken, AuthorizationCode, Session, SocialConnection, AuditLog エンティティ、UserStatus/ClientType/AuditAction 列挙型)
- [x] ドメイン層のインターフェース定義 (`internal/domain/`: 全リポジトリインターフェース、PasswordHasher, JWTManager, Mailer, SessionStore, RateLimiter, PermissionResolver, CaptchaVerifier インターフェース、ドメインエラー定義)
- [x] インフラ基盤: DB 接続・Redis 接続・乱数生成 (`internal/infra/postgres/db.go`, `internal/infra/redis/client.go`, `internal/infra/crypto/rand.go`: 接続プーリング設定含む)
- [x] DB マイグレーション初期スキーマ (`migrations/000001_initial_schema.up.sql`, `down.sql`: 全テーブル、部分ユニークインデックス含む)
- [x] トランザクションヘルパー (`internal/infra/postgres/tx.go`)
- [x] 基本ミドルウェア (`internal/middleware/`: Recovery, RequestID, Logging, Metrics, CORS)
- [x] ヘルスチェックハンドラ + レスポンス/リクエストヘルパー (`internal/handler/health.go`, `response.go`, `request.go`: 統一エラーレスポンス形式含む)
- [x] Chi ルーター骨格 + エントリポイント (`internal/handler/router.go`, `cmd/gate/main.go`: 設定読み込み → DB/Redis → マイグレーション → ミドルウェア → ルーター → HTTP サーバー → グレースフルシャットダウン)
- [x] compose.yml + Dockerfile + .env.example (`compose.yml`: PostgreSQL + Redis + Gate、マルチステージ Dockerfile、`.env.example` に必須環境変数の雛形)

---

## v0.2 - Core Authentication

**ゴール**: メール + パスワードによるユーザー登録・ログイン・ログアウト・メール検証・パスワード変更/リセットが動作する
**完動品としての価値**: ユーザーが登録・ログインでき、セッション ID ベースで認証状態を管理する。メール検証 (stdout) とパスワードリセットが動作する。JWT は v0.3 で追加、v0.2 ではセッション ID をレスポンスに返す

- [x] argon2id パスワードハッシュ実装 (`internal/infra/crypto/argon2.go`: PHC 形式、セマフォ同時実行制御、context キャンセル対応)
- [x] UserRepository PostgreSQL 実装 (`internal/infra/postgres/user.go`: CRUD、メール検索、楽観的ロック、部分ユニークインデックス活用)
- [x] Redis セッションストア実装 (`internal/infra/redis/session.go`: 作成/取得/削除、`session:{id}` Hash + `user:sessions:{user_id}` Set、同時セッション数制限)
- [x] Mailer インターフェース + stdout 実装 (`internal/infra/mailer/mailer.go`, `stdout.go`: 検証メール、パスワードリセットメール)
- [x] 認証ユースケース実装 (`internal/usecase/auth.go`: 登録、ログイン、ログアウト、メール検証、検証メール再送、パスワード変更、パスワードリセット、アカウントロック、ユーザー列挙防止)
- [x] 認証ハンドラ実装 (`internal/handler/auth.go`: `POST /api/v1/auth/register`, `login`, `logout`, `verify-email`, `resend-verification`, `forgot-password`, `reset-password`, `change-password`)
- [x] ルーターへの認証エンドポイント登録 + main.go ワイヤリング更新

---

## v0.3 - JWT & Token Management

**ゴール**: JWT アクセストークン発行・検証、リフレッシュトークンローテーション・ファミリー検出・失効が動作する
**完動品としての価値**: ログイン時に JWT + リフレッシュトークンが発行され、ローテーション・ファミリー検出・失効が機能する

- [x] JWT 署名・検証実装 (`internal/infra/crypto/jwt.go`: ES256/RS256、PEM 鍵読み込み、複数鍵対応、`kid` ヘッダー、JTI 生成、JWKS エクスポート)
- [x] TokenRepository PostgreSQL 実装 (`internal/infra/postgres/token.go`: リフレッシュトークン CRUD (SHA-256 ハッシュ保存)、ファミリー ID 検索・一括失効、`tokens_revoked_at` 判定、認可コード CRUD 構造)
- [x] トークンユースケース実装 (`internal/usecase/token.go`: ローテーション、ファミリー検出、グレースピリオド、個別/ユーザー単位失効、カスタムクレーム付与)
- [x] JWT 認証ミドルウェア (`internal/middleware/auth.go`: Bearer トークン検証、`kid` 鍵選択、`tokens_revoked_at` 判定、context 注入)
- [x] ログインフローの JWT 統合 (`internal/usecase/auth.go` 更新: ログイン成功時に JWT + リフレッシュトークン発行、ログアウト時にリフレッシュトークン失効追加。`internal/handler/oauth.go` 新規: `POST /oauth/token` (refresh_token)、`POST /oauth/revoke` 部分実装)
- [x] main.go ワイヤリング更新 (JWTManager, TokenRepository, TokenUsecase、JWT 認証ミドルウェア登録)

---

## v0.4 - Multi-Factor Authentication

**ゴール**: TOTP セットアップ・検証・無効化、リカバリーコード、ログインフローへの MFA 統合が動作する
**完動品としての価値**: ユーザーが TOTP を有効化でき、ログイン時に MFA が要求される。リカバリーコードによるフォールバックも機能する

- [ ] MFA ユースケース実装 (`internal/usecase/mfa.go`: TOTP セットアップ/確認/検証/無効化、リカバリーコード生成/再生成/検証、前後 1 ステップ時刻ずれ許容)
- [ ] MFA ハンドラ実装 (`internal/handler/mfa.go`: `POST /api/v1/mfa/totp/setup`, `confirm`, `DELETE /api/v1/mfa/totp`, `POST /api/v1/mfa/recovery-codes/regenerate`)
- [ ] ログインフローへの MFA 統合 (`internal/usecase/auth.go` 更新: MFA 有効ユーザーに `mfa_required` 中間レスポンス。`internal/handler/auth.go` 更新: `POST /api/v1/auth/mfa/verify`)
- [ ] ルーター + main.go ワイヤリング更新 (MFAUsecase, MFAHandler)

---

## v0.5 - OAuth 2.0

**ゴール**: Authorization Code + PKCE、Client Credentials、クライアント管理、トークンイントロスペクションが動作する
**完動品としての価値**: 外部アプリが OAuth 2.0 でアクセストークンを取得できる。サーバー間通信用 Client Credentials も利用可能

- [ ] ClientRepository PostgreSQL 実装 (`internal/infra/postgres/client.go`: OAuthClient CRUD、リダイレクト URI 管理 (最大 10、厳密一致、スキーム制限)、許可スコープ、クライアント認証、`tokens_revoked_at`)
- [ ] OAuth ユースケース実装 (`internal/usecase/oauth.go`: 認可コード発行/検証 (PKCE S256、有効期限設定可能)、再利用検出、Client Credentials、イントロスペクション、スコープ-パーミッション連携)
- [ ] Admin クライアント管理ユースケース (`internal/usecase/client.go`: 登録/更新/削除、secret ローテーション、クライアント単位全トークン失効)
- [ ] OAuth ハンドラ実装 (`internal/handler/oauth.go`: `GET /oauth/authorize`, `POST /oauth/token` (全 grant_type), `POST /oauth/revoke`, `POST /oauth/introspect`、RFC 6749 準拠エラー)
- [ ] Admin クライアント管理ハンドラ (`internal/handler/admin_client.go`: CRUD エンドポイント)
- [ ] ルーター + main.go ワイヤリング更新 (ClientRepository, OAuthUsecase, ClientUsecase, OAuthHandler, AdminClientHandler)

---

## v0.6 - OpenID Connect

**ゴール**: ID Token 発行、UserInfo、ディスカバリ、JWKS が動作する
**完動品としての価値**: Gate が OIDC Provider として機能し、クライアントは `/.well-known/openid-configuration` から設定を自動取得できる

- [ ] OIDC ユースケース実装 (`internal/usecase/oidc.go`: ID Token 発行 (必須クレーム + 標準クレーム)、UserInfo レスポンス、ディスカバリメタデータ、鍵ローテーション (旧鍵保持 = 最大 TTL + 10 分))
- [ ] OIDC ハンドラ実装 (`internal/handler/oidc.go`: `GET /.well-known/openid-configuration`, `GET /.well-known/jwks.json`, `GET /oauth/userinfo`)
- [ ] OAuth 認可フローへの OIDC 統合 (`internal/usecase/oauth.go` 更新: `scope` に `openid` 含む場合に ID Token 発行、`nonce` 伝搬)
- [ ] ルーター + main.go ワイヤリング更新 (OIDCUsecase, OIDCHandler)

---

## v0.7 - RBAC

**ゴール**: ロール/パーミッション CRUD、階層継承、循環参照検出、パーミッション解決 (Redis キャッシュ)、認可ミドルウェアが動作する
**完動品としての価値**: 管理者がロール/パーミッションを定義・割り当てし、API エンドポイントごとのアクセス制御が機能する

- [ ] RoleRepository PostgreSQL 実装 (`internal/infra/postgres/role.go`: ロール/パーミッション CRUD、階層 (parent_id)、循環参照検出 (再帰 CTE、最大深度 10)、カスケード除去)
- [ ] Redis パーミッションキャッシュ実装 (`internal/infra/redis/cache.go`: `permissions:{user_id}` Set、TTL 5 分、キャッシュ無効化)
- [ ] ロール/パーミッション管理ユースケース + パーミッション解決 (`internal/usecase/role.go`, `permission.go`: CRUD、割り当て、パーミッション解決 (直接 + 階層展開の和集合)、Redis キャッシュ連携)
- [ ] パーミッション認可ミドルウェア (`internal/middleware/permission.go`: `RequirePermission("...")` 形式、403 Forbidden)
- [ ] Admin ロール/パーミッション管理ハンドラ (`internal/handler/admin_role.go`: ロール CRUD、パーミッション CRUD、割り当て)
- [ ] ルーター + main.go ワイヤリング更新 (Admin エンドポイントに `RequirePermission("admin:access")` 適用)

---

## v0.8 - Social Login

**ゴール**: 汎用 OIDC コネクタベースで Google/GitHub/Apple ソーシャルログインが動作する。同一メール自動リンクが行われる
**完動品としての価値**: ユーザーが Google/GitHub/Apple アカウントでログインでき、既存アカウントと自動リンクされる

- [ ] 汎用 OIDC プロバイダコネクタ実装 (`internal/infra/social/provider.go`, `oidc.go`: SocialProvider インターフェース、OIDC Discovery、認可 URL 生成、code→token 交換、ユーザー情報取得)
- [ ] Google/GitHub/Apple プロバイダ実装 (`internal/infra/social/google.go`, `github.go`, `apple.go`: 各プロバイダ固有のカスタマイズ)
- [ ] SocialConnectionRepository PostgreSQL 実装 (`internal/infra/postgres/social.go`: CRUD、プロバイダ+provider_user_id ユニーク検索)
- [ ] ソーシャルログインユースケース (`internal/usecase/social.go`: 認可 URL 生成、コールバック、アカウントリンク (sub 検索 → 同一メール自動リンク → 新規作成)、プロバイダ側メール変更対応)
- [ ] ソーシャルログインハンドラ (`internal/handler/social.go`: `GET /api/v1/auth/social/{provider}/authorize`, `callback`)
- [ ] ルーター + main.go ワイヤリング更新

---

## v0.9 - Admin & Audit

**ゴール**: Admin ユーザー管理、監査ログ (全イベント記録)、アカウント自動物理削除が動作する
**完動品としての価値**: 管理者がユーザーを管理でき、全認証・認可イベントが監査ログに記録される

**前提: v0.3〜v0.8 の全マイルストーンが完了していること**

- [ ] AuditLogRepository PostgreSQL 実装 (`internal/infra/postgres/audit.go`: 書き込み、検索、保持期間に基づく削除)
- [ ] 監査ログの auth.go への統合 (ログイン、ログアウト、認証失敗の記録)
- [ ] 監査ログの token.go への統合 (トークン発行、失効の記録)
- [ ] 監査ログの oauth.go への統合 (認可コード発行、トークン交換の記録)
- [ ] 監査ログの role.go への統合 (権限変更の記録)
- [ ] 監査ログの mfa.go への統合 (MFA 操作の記録)
- [ ] 監査ログの social.go への統合 (ソーシャルログインの記録)
- [ ] Admin ユーザー管理ユースケース (`internal/usecase/user.go`: 一覧/詳細/更新/論理削除/ロック/アンロック/MFA リセット)
- [ ] アカウント自動物理削除 (バックグラウンドゴルーチン、`ACCOUNT_PURGE_DAYS` 設定可能、関連データカスケード削除)
- [ ] 監査ログ自動クリーンアップ (バックグラウンドゴルーチン、`AUDIT_LOG_RETENTION_DAYS` 設定可能)
- [ ] Admin ユーザー管理ハンドラ (`internal/handler/admin_user.go`: 一覧/詳細/更新/削除/ロック/アンロック/MFA リセット)
- [ ] ルーター + main.go ワイヤリング更新 (バックグラウンドジョブのグレースフルシャットダウン統合含む)

---

## v0.10 - Security Hardening & Operations

**ゴール**: レートリミット、HTTPS 強制、JTI リプレイ防止、OpenAPI 定義が完成し、本番デプロイ可能な状態にする
**完動品としての価値**: セキュリティ要件を満たし、API ドキュメントが揃った本番対応サーバー

- [ ] Redis レートリミットストア実装 (`internal/infra/redis/ratelimit.go`: Sliding Window Counter、Redis Sorted Set + Lua スクリプト)
- [ ] レートリミットミドルウェア (`internal/middleware/ratelimit.go`: IP ベース、エンドポイントごと設定、`Retry-After` 付き 429)
- [ ] レートリミットのルーター統合 (各エンドポイントグループに適切なレートリミットを適用)
- [ ] HTTPS 強制ミドルウェア (X-Forwarded-Proto チェック、本番環境のみ有効)
- [ ] JTI リプレイ防止 (Redis に短期間保存で重複検出)
- [ ] OpenAPI 3.0 定義ファイル作成 (全エンドポイントの API 仕様書)
- [ ] Prometheus メトリクスの最終確認 (全ミドルウェア・エンドポイントとの連携確認)
- [ ] Dockerfile / compose.yml / .env.example の最終調整

---

## Critic レビュー留保事項

- v0.2 のログイン成功レスポンスはセッション ID ベース。v0.3 で JWT に切り替え後、レスポンス形式が変わる
- テスト戦略はロードマップに明示されていないが、各タスク内でユニット/統合テストを含む想定
- CaptchaVerifier は v0.1 でインターフェース定義 + NoOp 実装 (常に true) を含める。具体プロバイダ実装は v2
