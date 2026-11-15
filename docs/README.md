# 文書の索引

`docs/` の文書はすべてここに載せる（`make docs-check` が漏れを落とす）。書き方の規則は [dev/docs-hygiene.md](dev/docs-hygiene.md)。

## プロダクト

- [product/concept.md](product/concept.md) — コンセプト、体験の基本原則、扱う情報、未決定事項
- [glossary.md](glossary.md) — 画面と文書で使う語の定義

## 設計

- [architecture/engine.md](architecture/engine.md) — エンジンの構成（土台のデータ、まとまりと責務）
- [architecture/theory-conventions.md](architecture/theory-conventions.md) — 読みと表記が従う流儀と、その出どころ
- [architecture/explanations.md](architecture/explanations.md) — 説明の文の規則（許す単純化、許さない誤り、止め方）
- [api/openapi.yaml](api/openapi.yaml) — API の契約（OpenAPI）。サーバーのルートと、画面の型の正
- [architecture/adr/README.md](architecture/adr/README.md) — 設計判断の記録（ADR）の索引

## 開発

- [dev/workflow.md](dev/workflow.md) — Issue と PR で回す開発ループ、完了の定義
- [dev/docs-hygiene.md](dev/docs-hygiene.md) — 文書の鮮度を保つ規則と自己点検
- [dev/harness.md](dev/harness.md) — `make check` と CI が行う検査の一覧
- [dev/knowledge-review.md](dev/knowledge-review.md) — 手元の知識が古くなったことに気づき、確かめ直す手順（月次の報告、四半期の総点検）
