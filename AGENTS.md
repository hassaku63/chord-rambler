# AGENTS.md — 開発ガイド

chord-rambler で作業する人間と AI コーディングエージェントの共通の正典。ツール固有の補足は `CLAUDE.md` に置き、規則の本体はここだけに書く。

## 1. このプロダクトは何か

音楽的なイメージを出発点に、ユーザーがコード進行を視覚的に探索・構築し、再生しながら学ぶ作曲支援アプリ。中心は、ユーザー自身がコードを選び、各選択の意味を理解しながら組み上げる体験。詳細は `docs/product/concept.md`。

エンジンは Go の API、画面は React と TypeScript（ADR-0008）。エンジンの構成は `docs/architecture/engine.md`。

## 2. 作業の進め方

1. 作業は GitHub の Issue から始める。`ready` ラベルの Issue を 1 つ選ぶ。Issue にないことは実装しない。思いついたことや見つけた問題は `observation` の Issue にする。
2. 何を作るかが決まっていないものは、`proposal` の Issue で選択肢と推奨を書き、所有者の判断を待つ。
3. 受け入れ条件が曖昧なら、実装せずに質問する。
4. テストを先に書く（または同じコミットで書く）。
5. `make check` を通す。通らないものは完成ではない。
6. 設計判断をしたら `docs/architecture/adr/` に 1 枚足す。影響する文書を同じ PR で直す（`docs/dev/docs-hygiene.md`）。
7. 1 PR = 1 Issue。PR テンプレートを埋め、スコープ外にしたことを書く。

詳細は `docs/dev/workflow.md`。

## 3. ガードレール

- **スコープを広げない。** 今の Issue に要らない機能・リファクタ・依存追加はしない。Issue にして次に回す。
- **テストを消す・スキップする・lint を無効化して通す** ことで緑にしない。根本原因を直す。
- **依存を足すときは ADR を書く。**
- **秘密情報を扱わない。** `.env` を読まない・コミットしない。鍵をコードに書かない。
- **`main` に直接 push しない。** `git push --force` をしない。
- **検査を足したら `docs/dev/harness.md` に 1 行足す。** `make check` にない検査を CI だけに置かない（ADR-0004）。
- **音楽理論の正しさを推測で埋めない。** 説明に教育上の単純化は許すが、明白な誤りや矛盾を含めない。根拠は文書化した定義とテストに置き、自信がなければ一次情報に当たる。
- 設計・実装・レビューで、機械で検出できない判断（テストの書き方、モックを使うか、抽象を足すか、型と境界、変更の分け方）をするときは、プラクティスの索引 `.claude/skills/practices/index.md` を引き、当たった項目を PR に書く（ADR-0007）。
- ファイル形式・記法・プロトコルを決めるときは、該当する業界標準の一次情報に当たる。採用する技術の書き方は、その公式の文書に従う。

## 4. 配置とコマンド

| 場所 | 役割 | 守ること |
| --- | --- | --- |
| `backend/cmd/server` | エントリポイント。配線だけ | ロジックを書かない |
| `backend/internal/config` | 環境変数からの設定 | 新しい設定は `Load` とテストに足す |
| `backend/internal/httpapi` | HTTP のハンドラ | decode、エンジンの呼び出し、encode だけ |
| `backend/internal/material`、`notation`、`progression`、`relation`、`reading`、`catalog`、`query`、`text`、`performance` | エンジン（`docs/architecture/engine.md`） | 標準ライブラリだけ。入出力・時刻・乱数を使わない。下のまとまりは上を読み込まない（lint で強制。向きは `backend/.golangci.yml`） |

- Go: 標準の `net/http`（メソッド付きのパターン）、`log/slog`。フレームワークは使わない。logger は引数で渡す。

```bash
make check        # すべての検査（CI は同じターゲットを呼ぶ）
make doctor       # 手元の道具の版が .tool-versions と合っているか
make fmt          # 自動整形
make dev-backend  # API サーバー（:8080。PORT で変える）
cd backend && go test ./internal/httpapi/ -run TestHealth -v   # 単体で回す例
```

## 5. コミットと PR

- Conventional Commits（`feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`, `ci:`）。
- コミットメッセージ・コード・コメントにモデル名や生成ツールの宣伝を書かない。

## 6. 文書の鮮度（ADR-0002）

- 文書は **現在の状態だけ** を書く。経緯は git log と ADR に、検討の過程は Issue に残す。
- 事実は 1 箇所にだけ書き、他は参照する。数値・日付・「予定」をコピーしない。
- `make docs-check` が、リンク切れ・存在しないパス・索引漏れを落とす。

## 7. 迷ったら

1. `docs/product/concept.md` の「体験の基本原則」に近い方を選ぶ。
2. 決定論的に解ける方を選ぶ。
3. 小さい方を選ぶ。
4. それでも決まらなければ、実装せずに質問する。
