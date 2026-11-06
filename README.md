# golang-api-server

Goを用いたTask CRUD APIサーバーのサンプルです。Echo、pgx、PostgreSQL + pgvectorを利用し、クリーンアーキテクチャを意識して外部技術とアプリケーションロジックを分離しています。

## 起動方法

DockerとDocker Composeが必要です。

```sh
docker compose up --build
```

起動後、`http://localhost:8080/health` が `{"status":"ok"}` を返します。DB MigrationはAPI起動前に自動実行されます。

ローカルでGoプロセスだけを動かす場合は、DBとMigrationを起動してから実行します。

```sh
cp .env.example .env
docker compose up -d db migrate
set -a; source .env; set +a
go run ./cmd/api
```

## API

| Method | Path | 説明 |
| --- | --- | --- |
| `POST` | `/api/v1/tasks` | Task作成 |
| `GET` | `/api/v1/tasks` | Task一覧（新しい順） |
| `GET` | `/api/v1/tasks/:id` | Task取得 |
| `PATCH` | `/api/v1/tasks/:id` | Task更新 |
| `DELETE` | `/api/v1/tasks/:id` | Task削除 |

作成例:

```sh
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"READMEを書く","description":"利用方法を追加する"}'
```

更新例（指定したフィールドだけを更新）:

```sh
curl -X PATCH http://localhost:8080/api/v1/tasks/<id> \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'
```

`title` は1〜200文字、`description` は最大2000文字です。エラーは `{"error":"..."}` の形式で返します。

## ディレクトリ構成

```text
cmd/api/                          エントリーポイントとDI
internal/domain/                  Entity、Repository interface、ドメインエラー
internal/usecase/                 アプリケーションロジック
internal/presentation/http/       Echo HandlerとHTTP DTO
internal/infrastructure/postgres/ pgxによるRepository実装とSQL
db/migrations/                    golang-migrate用Migration
```

依存方向は `Presentation / Infrastructure -> UseCase -> Domain` です。Repository interfaceはDomainに置き、PostgreSQL実装との差し替えを可能にしています。

## テスト

単体テストは実DBを使いません。

```sh
go test ./...
```

Repository統合テストはMigration済みのPostgreSQL + pgvectorを使います。

```sh
docker compose up -d db migrate
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/app?sslmode=disable' \
  go test -tags=integration ./internal/infrastructure/postgres/...
```

## アーキテクチャ要件

本プロジェクトは、Goのクリーンアーキテクチャを意識して実装すること。

### 基本方針

- Domain / UseCase / Presentation / Infrastructure の4層に分離する。
- 依存関係は外側から内側へ向ける。
- DomainおよびUseCaseはEcho、pgx、Supabaseなどの外部技術に依存させない。
- Repositoryはインターフェースを定義し、PostgreSQL実装と分離する。
- DIは`cmd/api/main.go`で明示的に行い、DIフレームワークは使用しない。
- HTTPリクエスト・レスポンス用DTOとDomain Entityを分離する。
- `context.Context`をHandlerからUseCase、Repositoryまで伝播する。
- DB固有のエラーはRepository層で変換し、HTTPステータスへの変換はPresentation層で行う。
- SQLはInfrastructure層に配置する。
- Migrationは`db/migrations/`で管理し、golang-migrateを使用する。
- DomainとUseCaseの単体テストでは、実DBを使わずRepositoryのFakeを利用する。
- Repositoryの統合テストでは、DockerのPostgreSQL + pgvectorを使用する。
- 循環依存を禁止する。
- プロトタイプのため、不要なインターフェースや抽象化は増やさない。

### 実装上の注意

将来RenderからAWS ECSへ、SupabaseからAmazon RDSへ移行することを想定し、アプリケーションロジックがインフラ固有の実装に依存しない構造とすること。

既存のCRUD API仕様、Docker Compose、Migration、GitHub Actions、テスト要件は維持すること。
