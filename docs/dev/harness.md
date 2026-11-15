# 開発ハーネスの一覧

`make check` と CI が行う検査の全リスト（ADR-0004）。**検査を足すときは必ずここに 1 行足す。** 長いあいだ何も捕まえず費用だけがある検査は、削除の候補にする。

CI の `required (make check)` は `make check` をそのまま呼ぶ。必須チェックはこの 1 つ。

## 文書（ADR-0002）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| リンク・パス・索引・競合マーカー | 存在しないものを指す文書、索引にない文書、マージで残った `<<<<<<<` | `make docs-check` | `make check` |
| `AGENTS.md` / `CLAUDE.md` の行数予算（150） | 読まれない指示書 | `make docs-check` | `make check` |

## 知識の鮮度（ADR-0007）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| プラクティスの索引の項目の形（欄の欠け、出どころ・ソース・対象・確認の書き方、40 項目の上限） | 出どころや確認日のない項目、増えすぎて引かれない索引 | `make practices-check` | `make check` |

## 構造と規約

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| shellcheck（`scripts/*.sh`） | 失敗を見ない `cd`、引用符のない展開など、シェルの典型的な誤り | `make shell-lint` | `make check` |
| actionlint（`.github/workflows/*.yml`。`run:` のスクリプトは shellcheck に渡す。版は `.tool-versions`。手元に actionlint がなければ、同じ版のコンテナイメージで走る） | ワークフローの書き間違い（式、イベント、ジョブの依存、権限）と、`run:` の中のシェルの誤り | `make actions-lint` | `make check` |

## 依存

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| Dependabot（週次、GitHub Actions） | SHA で固定したアクションの陳腐化 | — | GitHub |

## 手順（機械化されていない）

| 手順 | 目的 | いつ |
| --- | --- | --- |
| `docs-sync` 自己点検 | 差分で言及が変わる文書の修正 | PR 前（`docs/dev/docs-hygiene.md`） |
| 自己レビュー | 人間のレビュー負荷の低減 | PR 前 |
