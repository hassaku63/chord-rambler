# ADR（設計判断の記録）

設計判断をしたら、[0000-template.md](0000-template.md) を写して次の番号で 1 枚足し、この索引に載せる。ADR は決定時点の記録で、あとから本文を書き換えない。決定を改めるときは新しい ADR を足し、古い方の状態を `Superseded by NNNN` にする（ADR-0001）。

| 番号 | 決定 |
| --- | --- |
| [0001](0001-record-architecture-decisions.md) | 設計判断を ADR に記録する |
| [0002](0002-documentation-freshness.md) | 文書は現在の状態だけを書き、機械的な検査で鮮度を保つ |
| [0003](0003-issues-and-pull-requests.md) | 作業は Issue と PR で回し、文書に作業の一覧を持たない |
| [0004](0004-one-gate-make-check.md) | 検査は `make check` に集め、CI は同じターゲットを呼ぶ |
