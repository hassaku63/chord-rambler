# 業界標準の索引

使い方と項目の形は `SKILL.md`。

## コンテナとビルド

### Dockerfile（BuildKit の Dockerfile フロントエンドの構文）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Docker, Inc. と Moby プロジェクト（BuildKit。フロントエンドのイメージは `docker/dockerfile`） |
| 何を決めているか | コンテナイメージを組み立てる命令の書き方。パーサーの指令（`syntax`、`escape`、`check`）、命令、BuildKit の拡張（`RUN --mount`、`COPY --link`・`--parents`・`--exclude` など）と、それぞれが要るフロントエンドの最小の版 |
| 一次情報 | https://docs.docker.com/reference/dockerfile/ （正典。登録なしで読める）。版ごとの変更は https://github.com/moby/buildkit/releases の `dockerfile/` で始まるタグ。書き方の指針は https://docs.docker.com/build/building/best-practices/ |
| 状態 | 現行。`syntax` の指令は、多くの利用者には `docker/dockerfile:1` を勧めている。確かめた時点の最新の安定版は `dockerfile/1.29.0`。`-labs` の系列は別にある |
| 当たるべき場面 | `Dockerfile` を書く・変えるとき。BuildKit の拡張を使う前に、要る最小の版を確かめるとき。ベースイメージの固定（タグか digest か）を決めるとき |
| chord-rambler との接点 | `Dockerfile`（`# syntax=docker/dockerfile:1`、多段のビルド、`RUN --mount=type=cache`、`--target` の `api` と `web`、実行用のベースはタグと digest で固定。ADR-0010）。hadolint で検査する（`docs/dev/harness.md`） |
| 確認 | 2026-11-15 / 間隔 90 日 |

### Build checks（BuildKit の組み込みの検査）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Docker, Inc.（BuildKit） |
| 何を決めているか | Dockerfile とビルドの設定を、決まった規則（`UndefinedVar`、`SecretsUsedInArgOrEnv`、`CopyIgnoredFile`、`JSONArgsRecommended` など）で検査する仕組み。`docker build --check` は、ビルドを走らせずに検査だけを行い、違反があれば 0 でない終了コードで終わる。ふつうのビルドでは、違反は警告で、終了コードは 0。`check=error=true` の指令で、ふつうのビルドでも失敗にできる |
| 一次情報 | 規則の一覧は https://docs.docker.com/reference/build-checks/ 。使い方は https://docs.docker.com/build/checks/ 。どちらも登録なしで読める |
| 状態 | 現行。Dockerfile 1.8 と Buildx 0.15.0 から。規則は増えていく（experimental の印が付くものがある） |
| 当たるべき場面 | Dockerfile の検査を足す・変えるとき。hadolint と役割が重なるか確かめるとき。検査の違反を失敗にするかを決めるとき |
| chord-rambler との接点 | CI の `image` のジョブが、`api` と `web` に `docker buildx build --check` を走らせる（`docs/dev/harness.md`）。`check=error=true` の指令は置いていない |
| 確認 | 2026-11-15 / 間隔 90 日 |

### Compose Specification

| 欄 | 内容 |
| --- | --- |
| 発行元 | Compose Specification のプロジェクト（compose-spec。Docker Compose がその実装） |
| 何を決めているか | 複数のコンテナからなるアプリを 1 つのファイルで定める形式。サービス、ネットワーク、ボリューム、設定、秘密情報の書き方 |
| 一次情報 | https://github.com/compose-spec/compose-spec の `spec.md`。Docker Compose の実装としての参照は https://docs.docker.com/reference/compose-file/ 。どちらも登録なしで読める |
| 状態 | 現行。最上位の `version` は後方互換のためだけにあり、書くと廃止の警告が出る。既定のファイル名は `compose.yaml`（`compose.yml` も可。`docker-compose.yaml` は後方互換） |
| 当たるべき場面 | `compose.yaml` を書く・変えるとき。手元でコンテナを動かす構成（`make up`）を変えるとき |
| chord-rambler との接点 | `compose.yaml`（`web` と `api`。`api` は公開しない。ADR-0010）。`version` は書いていない |
| 確認 | 2026-11-15 / 間隔 180 日 |
