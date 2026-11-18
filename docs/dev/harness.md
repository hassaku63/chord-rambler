# 開発ハーネスの一覧

`make check` と CI が行う検査の全リスト（ADR-0004）。**検査を足すときは必ずここに 1 行足す。** 長いあいだ何も捕まえず費用だけがある検査は、削除の候補にする（四半期の総点検で見る。`docs/dev/knowledge-review.md`）。

CI の `required (make check)` は、Chromium を取ってから `make check` をそのまま呼ぶ。必須チェックはこの 1 つ。CI はそれに加えて、ネットワークが要る検査（`make backend-deps-check`、`make frontend-deps-check`）と、Docker が要る検査（`make dockerfile-lint`）と、比べる先が要る PR の大きさの予算（`make pr-size`）を、必須のジョブで走らせる。必須でない `image` のジョブが、イメージのビルドとスモークを走らせる。リポジトリの設定の CodeQL（既定の設定）が、ワークフローを置かずに静的解析を走らせる。

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
| 列挙の識別子の全域のテスト（エンジンの列挙のどの値も、`backend/internal/httpapi/names.go` で識別子に写り、契約の `enum` に同じ並びがある） | 識別子のない値、契約にない識別子。定数を足したときの写し漏れは、lint（exhaustive）も落とす | `cd backend && go test ./internal/httpapi/ -run TestEveryValueHasAnIdentifier` | `make check` |
| 埋め込んだカタログの読み込みと検証のテスト（項目の欠けと重複、知らない値と項目名、空の出どころと様式の範囲。JSON のファイルの一覧と、登録した読み込む関数の一覧の一致） | 壊れたカタログや、検証されないカタログを埋め込んだまま出すこと（ADR-0008 の 6） | `cd backend && go test ./internal/catalog/` | `make check` |
| Vitest と Testing Library | 画面の振る舞いの退行 | `make frontend-test` | `make check` |
| `tsc -b` | 型の矛盾 | `make frontend-typecheck` | `make check` |
| 生成した API の型の鮮度（`docs/api/openapi.yaml` から `frontend/src/shared/api/schema.gen.ts`） | 画面の型と契約のずれ。YAML や OpenAPI として壊れた契約 | `bash scripts/check-api-types.sh`（直すときは `cd frontend && npm run gen:api`） | `make check` |
| `vite build` | ビルドできない画面 | `make frontend-build` | `make check` |

## 画面（実ブラウザ。ADR-0017）

`make e2e` は、API をビルドし、Playwright の `webServer` で API と画面の開発サーバーを専用のポート（`frontend/playwright.config.ts`）で起動して、Chromium で `frontend/e2e/` を走らせる。手元の実測（2026-11-16、4 コア。API のビルドが済んだあと）は、`make e2e` が約 12 秒、`make check` の全体が約 40〜70 秒（Go のテストのキャッシュの有無で変わる）。CI での時間はまだ測っていない。

Chromium が要る。手元にないときは `cd frontend && npx playwright install chromium`（Linux で OS のライブラリもないときは `--with-deps`。CI はこれを `make check` の前に走らせる）。取れないときは、手元の Chromium の実行ファイルを渡す: `PW_CHROMIUM_PATH=/path/to/chrome make check`。

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| 通しのシナリオ（和音名で C を足す、候補から G7 を足す、鳴らすと鳴っている枠の印が進む。ヘッドレスの Chromium で `AudioContext` が作れること） | 部品のテストでは見えない、つなぎ目の退行 | `cd frontend && npx playwright test e2e/progression.spec.ts`（先に `make backend-build`） | `make check` |
| axe-core（WCAG 2.0・2.1・2.2 の A と AA のタグ。空の進行、枠を選んだ状態、候補を開いた状態） | 機械で見えるアクセシビリティの違反（名前の食い違い、小さすぎる操作の的など） | `cd frontend && npx playwright test e2e/accessibility.spec.ts` | `make check` |
| 配置（幅 360・768・1280 で、ページが横にスクロールしない。調、和音名、再生、候補の「足す」が見えて幅に収まる） | 画面の幅による配置の崩れ | `cd frontend && npx playwright test e2e/layout.spec.ts` | `make check` |

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
| Dependabot（週次。GitHub Actions、Go のモジュール、npm、Dockerfile の実行用のイメージ（digest）。`typescript` と `@types/node` の大きな版、`@vitest/eslint-plugin` の 1.6.28 以降は上げない。理由は ADR-0009） | 依存と、SHA で固定したアクションの陳腐化 | — | GitHub |

## 静的解析

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| CodeQL の既定の設定（Go、JavaScript/TypeScript、GitHub Actions。リポジトリの設定で有効にし、ワークフローのファイルは置かない。手元では走らせられない） | 入力の検証漏れや、ワークフローの危うい書き方のような、既知の型のセキュリティの欠陥 | — | GitHub（main への push と PR。結果は Security の画面） |

## 道具の版

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `.tool-versions` を CI の setup-go、setup-node、golangci-lint の導入が読む | CI と手元の版ずれ | — | CI |
| `make doctor` | 手元の Go、Node、golangci-lint、actionlint の major.minor のずれ。hadolint と Docker は、あるかどうかを知らせる | `make doctor` | CI、手元 |

## コンテナ（ADR-0010）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| hadolint（`Dockerfile`。版は `.tool-versions`。hadolint か動いている Docker が要るので `make check` には入れない） | Dockerfile の典型的な誤りと、`RUN` の中のシェルの誤り | `make dockerfile-lint` | CI の必須のジョブ |
| BuildKit のビルドチェック（`docker buildx build --check`、`api` と `web`） | hadolint にない規則: 定義されていない変数、`ARG`・`ENV` の秘密情報など | （CI の手順を参照） | CI の `image` |
| イメージのビルドと、スモークの検査（`scripts/smoke.sh`: `health` が渡した版を返す、`/` が画面の HTML を返す、存在しないファイルが 404、API の存在しないパスが problem の文書の 404） | Dockerfile と compose の腐敗、配信と中継の退行 | `make up`、`make smoke`、`make down`（Docker が要る） | CI の `image`（必須ではない。ログを必ず出す） |

## 変更の大きさ

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| PR の大きさの予算（足した行と消した行の和。lockfile、`*.gen.ts`、バイナリは数えない。600 行を超えると警告、1000 行を超えると失敗。比べる先が要るので `make check` には入れない） | 読みきれない大きさの PR | `make pr-size`（比べる先は `PR_BASE`、既定は `origin/main`） | CI の必須のジョブ（PR のときだけ。マージの結果のコミットの 1 つ目の親と比べる） |

## 知識の鮮度（ADR-0007、ADR-0015）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| プラクティスの索引の項目の形（欄の欠け、出どころ・ソース・対象・確認の書き方、40 項目の上限） | 出どころや確認日のない項目、増えすぎて引かれない索引 | `make practices-check` | `make check` |
| プラクティスの索引の版と依存の対応（項目の「対象」の版と、`frontend/package-lock.json` の大きな版・`.tool-versions` の Go の major.minor と Node の大きな版の食い違い。`frontend/package.json` の直接の依存と Go・Node が、どれかの項目の「対象」か「対象外の依存」の表のどちらか一方にあること。引けない名前は落とす。node が要る） | 依存の大きな版を上げても、関係する項目が古いまま残ること。依存を足したのに、当たるべき知識を持っているかを問わないこと | `make practices-check` | `make check` |
| 標準の索引の項目の形（欄の欠けと空の欄、一次情報の URL、確認の書き方） | 一次情報や確認日のない項目 | `make standards-check` | `make check` |
| 知識の鮮度の報告（プラクティスの索引と標準の索引の見直しの期限、出どころのリンク切れと移動、記録したコミットのあとのソースの変更。`docs/architecture/theory-conventions.md` の出どころのリンクも見る。日付とネットワークに依るので `make check` には入れない） | 手元に持った知識が古くなったことに気づけない | `bash scripts/knowledge-staleness.sh`（期限だけなら `--offline`） | 月次 `knowledge review`（指摘があれば `knowledge-review` の Issue。スクリプトかワークフローを変えた PR では、報告だけを走らせる） |

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
| 画面レビュー（`npm run ui:shots` で撮った画像とアクセシビリティツリーを、文脈を共有しないサブエージェントに見せる。指摘は仮説なので、検査ではなく手元の道具。`make check` と CI には入れない。`.claude/skills/ui-review/SKILL.md`） | 作った文脈を知らない目での、画面の分かりにくさと約束とのずれの発見 | 画面の構成・配置・状態の見せ方を変える PR の前 |
| 知識の総点検（出どころを開き直す、当たられなかった項目を削る、足りない項目を拾う） | 出どころが変わらないまま古くなった知識、手元に無い知識 | 四半期。所有者が始める（`docs/dev/knowledge-review.md`） |
