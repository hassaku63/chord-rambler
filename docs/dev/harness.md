# 開発ハーネスの一覧

`make check` と CI が行う検査の全リスト（ADR-0004）。**検査を足すときは必ずここに 1 行足す。** 長いあいだ何も捕まえず費用だけがある検査は、削除の候補にする。

CI の `required (make check)` は `make check` をそのまま呼ぶ。必須チェックはこの 1 つ。CI はそれに加えて、ネットワークが要る検査（`make backend-deps-check`、`make frontend-deps-check`）を走らせる。

## 文書（ADR-0002）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| リンク・パス・索引・競合マーカー | 存在しないものを指す文書、索引にない文書、マージで残った `<<<<<<<` | `make docs-check` | `make check` |
| `AGENTS.md` / `CLAUDE.md` の行数予算（150） | 読まれない指示書 | `make docs-check` | `make check` |

## コードの正しさ

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| Go のテスト（race、cover） | 振る舞いの退行 | `make backend-test`（手元の toolchain に coverage の道具がなければ `GOTEST_FLAGS=-race`） | `make check` |
| `go build` | ビルドできないコード | `make backend-build` | `make check` |
| ルートと契約の一致のテスト（`docs/api/openapi.yaml` の操作と、サーバーに登録したルート。契約の `paths` の形が崩れていても落ちる） | 契約にないルート、ルートのない操作（ADR-0008） | `cd backend && go test ./internal/httpapi/ -run TestRoutesMatchOpenAPI` | `make check` |
| Vitest と Testing Library | 画面の振る舞いの退行 | `make frontend-test` | `make check` |
| `tsc -b` | 型の矛盾 | `make frontend-typecheck` | `make check` |
| 生成した API の型の鮮度（`docs/api/openapi.yaml` から `frontend/src/shared/api/schema.gen.ts`） | 画面の型と契約のずれ。YAML や OpenAPI として壊れた契約 | `bash scripts/check-api-types.sh`（直すときは `cd frontend && npm run gen:api`） | `make check` |
| `vite build` | ビルドできない画面 | `make frontend-build` | `make check` |

## 構造と規約（Go）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| gofmt | 書式のばらつき | `make fmt-check` | `make check` |
| `go vet` と golangci-lint（標準のセットに、misspell、revive、gocritic、bodyclose、contextcheck、copyloopvar、dupl、errname、exhaustive、forcetypeassert、gosec、intrange、nakedret、nilerr、noctx、nolintlint、perfsprint、sloglint、unparam、usestdlibvars。設定は `backend/.golangci.yml`） | 典型的なバグと不統一、握りつぶしたエラー、注入されない logger、理由のない `//nolint` | `make backend-lint` | `make check` |
| `depguard`（エンジンの純粋さ） | エンジンのパッケージが、標準ライブラリと自分自身以外に依存すること。`os`・`net`・`io/fs`・`time`・`math/rand`・`log` を読み込むこと | `make backend-lint` | `make check` |
| `depguard`（まとまりの向き） | 下のまとまりが上を読み込むこと（material、notation、progression、relation、reading、catalog、query、text の順）。performance が読みを読み込むこと | `make backend-lint` | `make check` |
| `funlen` 80 行、`gocognit` 20、`gocyclo` 15 | 肥大した関数（テストは除く） | `make backend-lint` | `make check` |

## 構造と規約（TypeScript）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| Prettier | 書式のばらつき | `make fmt-check` | `make check` |
| ESLint（typescript-eslint の型つきの strict と stylistic、react-hooks。設定は `frontend/eslint.config.js`） | 典型的なバグ、要らない条件、非推奨の API、書き方の不統一 | `make frontend-lint` | `make check` |
| ESLint の機能の境界 | 機能どうしの読み込み、`../` での越境 | `make frontend-lint` | `make check` |
| ESLint のコンポーネントの純粋さ | `*.tsx` での API の呼び出し、`fetch`・`AudioContext`・保存・タイマー | `make frontend-lint` | `make check` |
| ESLint の Testing Library と Vitest の規則、スナップショットの禁止 | 実装の細部に結びついたテスト、待たない非同期、消し忘れの `.only` | `make frontend-lint` | `make check` |
| ESLint の `max-lines` 400、`max-lines-per-function` 120、`complexity` 15 | 肥大したファイルと関数（テストの関数の長さは除く） | `make frontend-lint` | `make check` |

## 依存

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `go mod tidy -diff`（ネットワークが要るので `make check` には入れない） | `go.mod` と `go.sum` の不整合、要らない依存 | `make backend-deps-check` | CI |
| `npm audit --omit=dev --audit-level=high`（ネットワークが要るので `make check` には入れない） | 本番の依存の既知の脆弱性 | `make frontend-deps-check` | CI |
| Dependabot（週次。GitHub Actions、Go のモジュール、npm。`typescript` と `@types/node` の大きな版、`@vitest/eslint-plugin` の 1.6.28 以降は上げない。理由は ADR-0009） | 依存と、SHA で固定したアクションの陳腐化 | — | GitHub |

## 道具の版

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `.tool-versions` を CI の setup-go、setup-node、golangci-lint の導入が読む | CI と手元の版ずれ | — | CI |
| `make doctor` | 手元の Go、Node、golangci-lint、actionlint の major.minor のずれ | `make doctor` | CI、手元 |

## 知識の鮮度（ADR-0007）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| プラクティスの索引の項目の形（欄の欠け、出どころ・ソース・対象・確認の書き方、40 項目の上限） | 出どころや確認日のない項目、増えすぎて引かれない索引 | `make practices-check` | `make check` |
| 知識の鮮度の報告（プラクティスの索引の見直しの期限、出どころのリンク切れと移動、記録したコミットのあとのソースの変更。`docs/architecture/theory-conventions.md` の出どころのリンクも見る。日付とネットワークに依るので `make check` には入れない） | 手元に持った知識が古くなったことに気づけない | `bash scripts/knowledge-staleness.sh`（期限だけなら `--offline`） | 月次 `knowledge review`（指摘があれば `knowledge-review` の Issue。スクリプトかワークフローを変えた PR では、報告だけを走らせる） |

## 構造と規約（スクリプトとワークフロー）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| shellcheck（`scripts/*.sh`） | 失敗を見ない `cd`、引用符のない展開など、シェルの典型的な誤り | `make shell-lint` | `make check` |
| actionlint（`.github/workflows/*.yml`。`run:` のスクリプトは shellcheck に渡す。版は `.tool-versions`。手元に actionlint がなければ、同じ版のコンテナイメージで走る） | ワークフローの書き間違い（式、イベント、ジョブの依存、権限）と、`run:` の中のシェルの誤り | `make actions-lint` | `make check` |

## 手順（機械化されていない）

| 手順 | 目的 | いつ |
| --- | --- | --- |
| `docs-sync` 自己点検 | 差分で言及が変わる文書の修正 | PR 前（`docs/dev/docs-hygiene.md`） |
| 自己レビュー | 人間のレビュー負荷の低減 | PR 前 |
| 知識の総点検（出どころを開き直す、当たられなかった項目を削る、足りない項目を拾う） | 出どころが変わらないまま古くなった知識、手元に無い知識 | 四半期。所有者が始める（`docs/dev/knowledge-review.md`） |
