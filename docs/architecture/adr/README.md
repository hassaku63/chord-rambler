# ADR（設計判断の記録）

設計判断をしたら、[0000-template.md](0000-template.md) を写して次の番号で 1 枚足し、この索引に載せる。ADR は決定時点の記録で、あとから本文を書き換えない。決定を改めるときは新しい ADR を足し、古い方の状態を `Superseded by NNNN` にする（ADR-0001）。

| 番号 | 決定 |
| --- | --- |
| [0001](0001-record-architecture-decisions.md) | 設計判断を ADR に記録する |
| [0002](0002-documentation-freshness.md) | 文書は現在の状態だけを書き、機械的な検査で鮮度を保つ |
| [0003](0003-issues-and-pull-requests.md) | 作業は Issue と PR で回し、文書に作業の一覧を持たない |
| [0004](0004-one-gate-make-check.md) | 検査は `make check` に集め、CI は同じターゲットを呼ぶ |
| [0005](0005-engine-structure-by-kind-of-knowledge.md) | エンジンは知識の種類でまとまりを分ける |
| [0006](0006-progression-of-slots-with-absolute-chords.md) | 進行は、絶対の和音を持つ枠の列として保存する |
| [0007](0007-rules-in-three-tiers-and-a-dated-practice-index.md) | 規則は 3 段で持ち、索引は出どころと確認日で鮮度を見る |
| [0008](0008-tech-stack.md) | エンジンは Go の API、画面は React と TypeScript |
| [0009](0009-frontend-supporting-dependencies.md) | フロントエンドの道具を支える依存 |
| [0010](0010-container-images.md) | コンテナは、distroless の API と nginx の画面の 2 つのイメージ |
| [0011](0011-models-by-stage.md) | 段階ごとにモデルを使い分け、`ready` の Issue を引き継ぎの文書にする |
| [0012](0012-explanations-refine-without-retraction.md) | 説明の文は、取り消さずに詳しくできる単純化だけを許す |
