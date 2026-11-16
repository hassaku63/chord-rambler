# 0009. フロントエンドの道具を支える依存

- 日付: 2026-11-15
- 状態: Accepted

## 文脈

ADR-0008 は、画面の道具を React、TypeScript、Vite、Vitest、Testing Library、ESLint と typescript-eslint、Prettier、契約から型を生成する道具と決めた。これらを動かすには、名前を挙げていない依存がいくつか要る。依存を足すときは ADR を書く規則（`AGENTS.md`）なので、土台を入れる時点の一覧と、それぞれが要る理由を記録する。

## 決定

実行時の依存は `react` と `react-dom` だけにする。開発時の依存は、ADR-0008 の道具と、次の「支える依存」に限る。

| 依存 | 要る理由 |
| --- | --- |
| `@vitejs/plugin-react` | Vite で JSX を扱う |
| `@types/react`、`@types/react-dom`、`@types/node` | 型の定義 |
| `jsdom` | Vitest で DOM を使うテストを走らせる |
| `@testing-library/dom`、`@testing-library/jest-dom` | Testing Library の本体と、DOM 向けの照合 |
| `@eslint/js`、`globals` | ESLint の推奨の規則と、ブラウザと Node の大域の名前 |
| `eslint-plugin-react-hooks` | hook の規則 |
| `eslint-plugin-testing-library`、`@vitest/eslint-plugin` | テストの書き方の規則 |
| `eslint-config-prettier` | Prettier と重なる ESLint の規則を切る |
| `openapi-typescript` | 契約から型を生成する |
| `@playwright/test` | 画面を実ブラウザで通しで検査する（ADR-0017） |
| `@axe-core/playwright` | 実ブラウザの画面のアクセシビリティを axe-core で検査する（ADR-0017） |

- 版の固定: `typescript` は 5.9 系（ADR-0008）。`@vitest/eslint-plugin` は 1.6.27（それより新しい版は TypeScript 6 以上を求める）。Dependabot は、この 2 つと `@types/node` の大きな版を上げない。
- ここにない依存を足すときは、新しい ADR を書く。

## 検討した選択肢

- **テストと lint のプラグインを入れない**: 依存が減る。テストの書き方（role で引く、待つ、スナップショットを使わない）と hook の規則を、レビューだけで守ることになる。機械で検出できるものは機械で検出する方針（ADR-0004、ADR-0007）に合わない。
- **middy の画面と同じ一式を入れる**（未使用のコードの検出、PWA、通しのテストなど）: 使う時点で足す。今は使わない。

## 結果

- 依存の一覧が、理由と一緒に 1 か所で読める。
- プラクティスの索引に依存の網羅の検査を入れるとき（#16）、この一覧が出発点になる。
- 再検討の条件: TypeScript を上げられるようになったとき（型の生成の道具が新しい系に対応したとき）、版の固定を外す。
