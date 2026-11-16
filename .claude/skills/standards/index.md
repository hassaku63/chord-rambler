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

## 音楽の表記

### MusicXML の音高（`pitch` の `step` と `alter`）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Music Notation Community Group（MusicXML 4.0。W3C の標準ではなく、Community Group の Final Report） |
| 何を決めているか | 楽譜を交換する形式。音高を、全音階の幹音（`step`。A〜G）、半音の数で書く変化（`alter`）、オクターブの組で表す |
| 一次情報 | https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/pitch/ 、`step` は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/step/ 、`alter` は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/alter/ 。どれも登録なしで読める。編集中の版は https://w3c.github.io/musicxml/ |
| 状態 | 現行。確かめた版は 4.0（Final Community Group Report、2021-06-01） |
| 当たるべき場面 | 音名の綴りの持ち方を決める・変えるとき。楽譜の形式を読み書きするとき。表記のまとまりで、和音の根音やバスの綴りを書くとき |
| chord-rambler との接点 | `backend/internal/material` の音名（`NoteName`）が、幹音と変化の組で綴りを持つ（`docs/architecture/domain-model.md`）。食い違い: MusicXML の `alter` は小数（微分音）を許すが、音名の変化は −2〜+2 の整数に限る。音名はオクターブを持たない |
| 確認 | 2026-11-15 / 間隔 365 日 |

### MusicXML の和音の種類（`harmony` の `kind`）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Music Notation Community Group（MusicXML 4.0。W3C の標準ではなく、Community Group の Final Report） |
| 何を決めているか | 和音記号を、根音（`root`）、種類（`kind`）、バス（`bass`）、種類からの足し引きと変化（`degree`）で表す。`kind` の値は、三和音、七の和音、6・9・11・13 の和音、sus、パワーコード、機能和声の和音（ナポリ、増六）などの決まった語彙で、多くは根音からの音程で定義される |
| 一次情報 | `kind` の値は https://www.w3.org/2021/06/musicxml40/musicxml-reference/data-types/kind-value/ 、要素は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/harmony/ 、`kind` は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/kind/ 、`degree` は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/degree/ 、`bass` は https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/bass/ 。どれも登録なしで読める |
| 状態 | 現行。確かめた版は 4.0（Final Community Group Report、2021-06-01）。`function` の要素は 4.0 で非推奨 |
| 当たるべき場面 | 和音名の読み書きを作る・変えるとき（表記のまとまり）。和音の種類の語彙や、テンションと変化の書き方を決めるとき。楽譜の形式と和音を交換するとき |
| chord-rambler との接点 | `backend/internal/material` の和音（`Chord`）は、種類の名前を持たず、根音からの音程の集合で持つ（`docs/architecture/domain-model.md`）。種類は見方として導く。和音名の 6 と m6 は `major-sixth`・`minor-sixth`（三和音に足した 6 度）に当たる。食い違い: MusicXML のバスは音名（`bass-step` と `bass-alter`）で、和音のバスは根音からの音程。add9 と m(add9) に当たる `kind` はなく、MusicXML では三和音の `kind` に `degree` で 9 度を足して表す |
| 確認 | 2026-11-16 / 間隔 365 日 |

### Harte の和音の記法（Harte、Sandler、Abdallah、Gómez、ISMIR 2005）

| 欄 | 内容 |
| --- | --- |
| 発行元 | C. Harte、M. Sandler、S. Abdallah（Queen Mary, University of London）、E. Gómez（Universitat Pompeu Fabra）。ISMIR 2005 の論文「Symbolic representation of musical chords: a proposed syntax for text annotations」 |
| 何を決めているか | 和音を調によらず、平文で曖昧なく書く記法。根音（幹音と `#`・`b`）、`:`、括弧の中にコンマで並べた度（`b3` のように変化を前に置く）、`/` のあとの度で書くバス。よく使う種類の短縮名（`maj`、`min`、`hdim7` など）と、足す度の括弧、省く度の `*`。文法を BNF で示す |
| 一次情報 | https://archives.ismir.net/ismir2005/paper/000080.pdf （登録なしで読める PDF） |
| 状態 | 2005 年の論文。開いたのはこの版だけで、のちの改訂は確かめていない |
| 当たるべき場面 | 和音名の読み書きを作る・変えるとき（表記のまとまり）。和音名の文法や、足す音の書き方、バスの書き方を決めるとき。和音の注釈のデータと和音を交換するとき |
| chord-rambler との接点 | `backend/internal/notation` の和音名（`docs/architecture/theory-conventions.md` の 6）。同じところ: 根音と、根音からの音程で和音を持つ（`backend/internal/material`）。足す音を括弧の中にコンマで並べる。`B` と `b` を大文字・小文字で分ける。6 と m6 の音程（`maj6 (3,5,6)`、`min6 (b3,5,6)`）。食い違い: 和音名に `:` を書かない。バスは音名で書く（Harte は度）。変化記号に `♯`・`♭` も使う（Harte は `#`・`b` だけ）。種類は短い名前（`m`、`dim`）で書く（Harte は打ち間違いを見つけやすいように長い名前を選ぶ）。Harte の短縮名には sus2 と add9 がない（m(add9) は、Harte の文法では `min(9)` と書ける） |
| 確認 | 2026-11-16 / 間隔 365 日 |

### RomanText（Tymoczko、Gotham、Cuthbert、Ariza、ISMIR 2019）

| 欄 | 内容 |
| --- | --- |
| 発行元 | D. Tymoczko（Princeton）、M. Gotham（Cornell）、M. S. Cuthbert（MIT）、C. Ariza。ISMIR 2019 の論文「The RomanText format: a flexible and standard method for representing Roman numeral analyses」。読み書きの実装は music21 |
| 何を決めているか | ローマ数字の和声分析を平文で書く形式。小節と拍、調の切り替え（`C:`、`c:`）、和音の書き方。和音は、大文字が長・小文字が短、増の `+`、減の `o`、半減 7 の `/o`、根音を上げ下げする前置の `#`・`b`、転回形の数字（`6`、`6/4`、`6/5`、`4/3`、`2`）、角括弧の足し引きと変化（`[no5]`、`[add#6]`、`[b5]`）、`/` のあとの向かう先（`V/V`。`V7/V/V` のように重ねてよい）、増六とナポリの略記（`Ger`、`It`、`Fr`、`N`）。短調の 6・7 度の扱いを設定で変えられる |
| 一次情報 | https://archives.ismir.net/ismir2019/paper/000012.pdf （登録なしで読める PDF） |
| 状態 | 2019 年の論文。開いたのはこの版だけで、のちの改訂は確かめていない |
| 当たるべき場面 | ローマ数字の読み書きを作る・変えるとき（表記のまとまり）。転回形、足し引き、向かう先の書き方を決めるとき。分析のデータとローマ数字を交換するとき |
| chord-rambler との接点 | `backend/internal/notation` のローマ数字と主音から見た位置（`docs/architecture/theory-conventions.md` の 6）。同じところ: 大文字・小文字で長・短を分ける。増の `+`。調の音でない根音に前置の変化記号。`/` のあとの向かう先。食い違い: 減は `°`、半減 7 は `ø7`（RomanText は `o` と `/o`）。変化記号は `♭`・`♯`（RomanText は `b`・`#`）。七度とテンションは和音名の書き方（`Imaj7`、`V7(9)`）で、角括弧の足し引きを使わない。転回形と、重ねた向かう先を書かない。RomanText の `6` は転回形の数字（`I6` は第 1 転回形）だが、ここの `I6` は 6 の和音（`docs/architecture/theory-conventions.md` の 6）。調を書かない（位置は調を持たない）。sus の書き方は、論文の和音の一覧にない |
| 確認 | 2026-11-16 / 間隔 365 日 |

## MIDI

### MIDI のノート番号（MIDI 1.0 の Note On と Note Off）

| 欄 | 内容 |
| --- | --- |
| 発行元 | The MIDI Association |
| 何を決めているか | 楽器どうしが演奏を伝えるメッセージ。Note On と Note Off は、鍵盤の番号（key (note) number）と強さ（velocity）を、それぞれ 7 ビットのデータバイト（0〜127）で運ぶ |
| 一次情報 | メッセージの一覧は https://midi.org/summary-of-midi-1-0-messages （登録なしで読める。ページ自身が、正式の仕様に当たるよう勧めている）。正式の仕様は MIDI 1.0 Detailed Specification で、https://midi.org/midi-1-0-core-specifications から Google Drive の PDF を開く。PDF は開けず、ノート番号の定義の文言と、中央の C をどの番号に置くかは確かめていない |
| 状態 | 現行。core specifications のページは、Detailed Specification を 1996 年の改訂としている。MIDI 2.0（Universal MIDI Packet）の仕様が別にある（確かめていない） |
| 当たるべき場面 | 音の高さを数で表す・受け渡すとき（演奏のまとまりの音符、画面での周波数への換算、標準 MIDI ファイルの書き出し）。オクターブの番号の付け方を決めるとき。強さの値を決めるとき |
| chord-rambler との接点 | `backend/internal/performance` の音符（`Note`）が、ノート番号と強さを持つ（`docs/architecture/domain-model.md` の 9）。C4 = 60 は、このプロジェクトの規約として採る（オクターブの番号の付け方は流儀が割れ、ここでは一次情報で確かめていない）。強さの定数の値は、仕様から引いていない |
| 確認 | 2026-11-16 / 間隔 365 日 |

## Web と API

### Problem Details for HTTP APIs（RFC 9457）

| 欄 | 内容 |
| --- | --- |
| 発行元 | IETF（M. Nottingham、E. Wilde、S. Dalal。2023 年 7 月） |
| 何を決めているか | HTTP の API が失敗を返すときの、機械で読める本文の形。メディアタイプ `application/problem+json` と `application/problem+xml`、標準の欄（`type`、`status`、`title`、`detail`、`instance`）とそれぞれの意味、問題の種類ごとの拡張の欄 |
| 一次情報 | https://www.rfc-editor.org/rfc/rfc9457 （登録なしで読める） |
| 状態 | 現行。Proposed Standard。RFC 7807 を廃止して置き換えた |
| 当たるべき場面 | API の失敗の応答を足す・変えるとき。`detail` に何を書くか、`type` を定めるか、拡張の欄を足すかを決めるとき。画面が失敗の応答を読むところを作るとき |
| chord-rambler との接点 | `backend/internal/httpapi` の `writeProblem` と、契約（`docs/api/openapi.yaml`）の `Problem`。`type` を書かないので、RFC の定めにより `about:blank` とみなされ、`title` は状態の標準の句（`http.StatusText`）にしている（RFC はそれを SHOULD とする）。`status` は応答の状態と同じ値。`detail` は、直し方が分かるように、読めなかった枠と所を書く。RFC は、受け手が `detail` を解析しないよう求めるので、画面は `detail` から情報を取り出さない。食い違い: RFC はどの欄も必須にしないが、契約は `title` と `status` を必須にしている（このサーバーはいつも書く）。`instance` と拡張の欄は使っていない |
| 確認 | 2026-11-16 / 間隔 365 日 |
