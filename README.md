# Gate

汎用認証・認可サーバー。Clean Architecture で構築された Go 製の認証基盤。

## 機能

- **メール + パスワード認証** — 登録・ログイン・ログアウト・メール検証・パスワードリセット
- **JWT トークン管理** — ES256/RS256、リフレッシュトークンローテーション、ファミリー検出
- **MFA** — TOTP (Google Authenticator 等)、リカバリーコード
- **OAuth 2.0** — Authorization Code + PKCE、Client Credentials、イントロスペクション
- **OpenID Connect** — ID Token、UserInfo、Discovery (`/.well-known/openid-configuration`)、JWKS
- **RBAC** — 階層ロール・パーミッション、Redis キャッシュ、認可ミドルウェア
- **ソーシャルログイン** — Google、GitHub (アカウント自動リンク)
- **Admin API** — ユーザー管理、クライアント管理、ロール/パーミッション管理、監査ログ
- **セキュリティ** — レートリミット (Sliding Window)、HTTPS 強制、JTI リプレイ防止

## 技術スタック

| カテゴリ | ライブラリ |
|---|---|
| HTTP | [Chi](https://github.com/go-chi/chi) |
| DB | PostgreSQL (pgx/v5 + sqlx) |
| キャッシュ | Redis (go-redis/v9) |
| JWT | golang-jwt/jwt/v5 (ES256) |
| パスワード | argon2id (PHC 形式) |
| MFA | pquerna/otp |
| マイグレーション | golang-migrate/migrate/v4 |
| 設定 | caarlos0/env/v11 |
| ログ | log/slog |
| メトリクス | prometheus/client_golang |

## 起動

### 前提条件

- Go 1.25+
- Docker / Docker Compose

### 開発環境

```bash
# リポジトリ取得
git clone https://github.com/HMasataka/gate
cd gate

# 環境変数設定
cp .env.example .env

# PostgreSQL + Redis + Gate を起動
docker compose up
```

### ローカルビルド

```bash
go mod tidy
go build -o gate ./cmd/gate
./gate
```

## 設定

環境変数で設定します。`.env.example` を参照してください。

主要な設定項目:

| 環境変数 | 説明 | デフォルト |
|---|---|---|
| `DATABASE_URL` | PostgreSQL 接続 URL | — |
| `REDIS_URL` | Redis 接続 URL | — |
| `JWT_ALGORITHM` | JWT 署名アルゴリズム | `ES256` |
| `JWT_PRIVATE_KEY_PATH` | PEM 秘密鍵パス (省略時は開発用鍵を自動生成) | — |
| `ACCESS_TOKEN_EXPIRY` | アクセストークン有効期限 | `15m` |
| `REFRESH_TOKEN_EXPIRY` | リフレッシュトークン有効期限 | `168h` |
| `SESSION_MAX_CONCURRENT` | 同時セッション数上限 | `5` |
| `MFA_RECOVERY_CODE_COUNT` | リカバリーコード生成数 | `8` |
| `GOOGLE_CLIENT_ID` | Google OAuth クライアント ID | — |
| `GITHUB_CLIENT_ID` | GitHub OAuth クライアント ID | — |
| `HTTPS_REDIRECT` | HTTPS リダイレクト有効化 | `false` |

## API エンドポイント

### 認証

| メソッド | パス | 説明 |
|---|---|---|
| POST | `/api/v1/auth/register` | ユーザー登録 |
| POST | `/api/v1/auth/login` | ログイン (JWT + セッション発行) |
| POST | `/api/v1/auth/logout` | ログアウト |
| POST | `/api/v1/auth/verify-email` | メール検証 |
| POST | `/api/v1/auth/forgot-password` | パスワードリセットメール送信 |
| POST | `/api/v1/auth/reset-password` | パスワードリセット |
| POST | `/api/v1/auth/change-password` | パスワード変更 (JWT 必須) |

### MFA

| メソッド | パス | 説明 |
|---|---|---|
| POST | `/api/v1/mfa/totp/setup` | TOTP セットアップ |
| POST | `/api/v1/mfa/totp/confirm` | TOTP 有効化 |
| DELETE | `/api/v1/mfa/totp` | TOTP 無効化 |
| POST | `/api/v1/mfa/recovery-codes/regenerate` | リカバリーコード再生成 |

### OAuth 2.0 / OIDC

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/api/v1/oauth/authorize` | 認可エンドポイント |
| POST | `/oauth/token` | トークン取得 (code / refresh_token / client_credentials) |
| POST | `/oauth/revoke` | トークン失効 |
| POST | `/oauth/introspect` | トークンイントロスペクション |
| GET | `/oauth/userinfo` | UserInfo (JWT 必須) |
| GET | `/.well-known/openid-configuration` | OIDC Discovery |
| GET | `/.well-known/jwks.json` | JWKS |

### ソーシャルログイン

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/api/v1/auth/social/{provider}/authorize` | 認可 URL へリダイレクト |
| GET | `/api/v1/auth/social/{provider}/callback` | コールバック処理 |

`{provider}`: `google` / `github`

### ヘルスチェック

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/health` | Liveness |
| GET | `/ready` | Readiness (DB + Redis 接続確認) |
| GET | `/metrics` | Prometheus メトリクス |

詳細は [`api/openapi.yaml`](api/openapi.yaml) を参照してください。

## アーキテクチャ

Clean Architecture に基づく 4 レイヤー構成。依存方向は外側から内側のみ。

```
cmd/gate/main.go          エントリポイント・手動 DI
internal/domain/          エンティティ・インターフェース (外部依存なし)
internal/usecase/         ビジネスロジック
internal/handler/         HTTP ハンドラ
internal/middleware/       ミドルウェア
internal/infra/postgres/  PostgreSQL 実装
internal/infra/redis/     Redis 実装
internal/infra/crypto/    argon2id・JWT・乱数
internal/infra/social/    ソーシャルログインプロバイダ
internal/infra/mailer/    メーラー (stdout 実装)
internal/config/          設定構造体
migrations/               SQL マイグレーション
api/                      OpenAPI 仕様
```

## License

MIT
