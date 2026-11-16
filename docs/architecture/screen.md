# 画面の構成

画面は React と TypeScript で作り、エンジンの API を呼ぶ（ADR-0008）。ここは今の画面の構成。

## 仮の配置

今の配置は仮のもの（#148 の A1）。2 つの体験を 1 つの画面にまとめるか、分けるかは、構成の企画（`docs/product/concept.md` の 8 章）で決め、配置もそこで作り直す。

画面 1 は、進行を組み、読みの札を見る。画面 2 は、それを鳴らす。上から次の順に並ぶ。

1. 調の選択（12 の長調。既定は C）。再生の操作（下の「鳴らし方」）。和音名の入力と、枠を足す操作。
2. 枠の帯。枠を左から右に並べ、折り返す。カードは、和音名・ローマ数字・札の短い印（種別、機能、解決・終止）と、長さ（1〜16 拍）・消す操作を持つ。鳴っている枠は、カードに「鳴っています」の文字と破線の枠を出し、`aria-current` を付ける。
3. 選んだ枠の詳しい札。種別、機能、借り先、向かう先、解決、終止、テンションの注意、役割に合うかを、語と根拠の種類の印で出す。文にはしない（#110 の B1）。

## 置き場所

| 場所 | 中身 |
| --- | --- |
| `frontend/src/features/progression/useProgression.ts` | 進行の状態と操作。`localStorage` に API の `Progression` と同じ形で残し、読み出した値は境界で検証する（壊れていれば空の進行） |
| `frontend/src/features/progression/useProgressionReading.ts` | 進行が変わるたびに `POST /progressions/readings` を呼ぶ。古い応答は捨てる |
| `frontend/src/features/progression/usePlayback.ts` | 進行が変わるたびに `POST /progressions/notes` を呼び（古い応答は捨てる）、テンポを持ち、進行の全体か 1 枠を鳴らす |
| `frontend/src/features/progression/schedule.ts` | 音符を発振器の予定（周波数、始まりの秒、長さ、音量）に直す純粋な関数 |
| `frontend/src/features/progression/usePlayer.ts` | Web Audio で予定を鳴らす。`AudioContext` を作る関数を引数で受ける（テストは `frontend/src/test/fakeAudioContext.ts` を渡す） |
| `frontend/src/features/progression/terms.ts` | 札の識別子を語にする表。ここ 1 か所に置く |
| `frontend/src/shared/api/client.ts` | API の呼び出しと、応答の検証 |

## 札の語

- 語は `docs/glossary.md` の語だけを使う。
- 契約が「名付けるものがない」とする値（`none`、`facts-only`）は、札を出さない。
- 表にない識別子は、表示を落とさずに識別子のまま出す。
- テンションの注意の音程は、契約の書き方（`P11`、`M3`）のまま出す。

## 読みの失敗

API の problem の `detail` を、`role="alert"` でそのまま出す。`detail` が枠を名指すとき（`slot 2: ...`。0 から数える）は、その枠のカードにも印を付ける。

## 鳴らし方

音は Web Audio API で直接鳴らす（ADR-0008 の 7）。エンジンは拍で数えた音符を返し、拍を秒に直すのは画面の側。

- 操作は「進行を鳴らす」（先頭から最後まで）、「この枠だけ鳴らす」（選んだ枠の音符だけを、その枠の長さで）、「止める」、テンポのスライダー。
- 再生中のボタンは `aria-pressed` で示す。押されているボタンをもう一度押すと止まる。
- テンポは 40〜200 BPM、既定は 90。変えたテンポは次の再生から効く。進行が変わると、再生は止まる。
- 音の高さは、MIDI のノート番号から平均律で求める（A4 = 69 = 440 Hz）。強さ（0〜127）は音量に写す。
- 音色は 1 つ。三角波に、短い立ち上がりと終わりまでの減衰を付ける。和音の声部が重なっても割れないよう、全体の音量を下げて出す。数値は `usePlayer.ts` の定数。
- `AudioContext` は最初の操作で作り、止まっていれば再開する（ブラウザの自動再生の制限のため）。
- 音符を得られない（拒まれた、つながらない）あいだ、和音のない枠を選んだときは、鳴らすボタンを押せない。理由は読みの失敗と同じなので、読みの側で出す。
