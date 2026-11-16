# 採っている流儀と出どころ

エンジンの「読み」と「表記」（`docs/architecture/engine.md` のまとまり 4 と 2）が従う規則。和声の教材の間で立場が割れる点について、どの立場を採り、どこに書いてあるかを記す。規則を足す・変えるときは、ここに出どころを書く。

## 1. 機能

- **採る立場**: トニック（T）、サブドミナント（S）、ドミナント（D）の 3 つに分ける。長調では I・iii・vi が T、ii・IV が S、V・vii° が D。iii と vi は「トニックの仲間」と呼ぶ。IV → I は S から T への動きとして述べる。
- ドミナントの強さは、導音を含むか（属七では三全音を含むか）で述べる。導音のない和音を、V と同じ強さのドミナントとは言わない。
- 機能を付けるのは、ダイアトニックの和音だけ。ダイアトニックの和音は、和音を作る音（根音と、2〜7 度の構成音。sus2・sus4 の 2 度・4 度と 6 の和音の 6 度を含む）と 7 度のバス（根音の短 7 度上の音のバス。構成音でなくてもよい）の音名が、綴り（幹音と変化）で調の音階の音名に含まれる和音。テンション（9・11・13 度）と、それ以外のバスは見ない（C で `G7(♭9)` は V、`C(♯11)` と `C/F♯` は I、`Fsus4` は B♭ を持つのでダイアトニックでない）。異名同音の綴りの和音（C で B♯ の長三和音は ♯VII、A♯ の長三和音は ♯VI）は、どの同主旋法の音階にも綴りで含まれないので、ダイアトニックとも借用和音とも呼ばず、事実（位置と構成音）だけを述べる（#101 の所有者の判断 D1、#102 の所有者の判断）。
- 借用和音には機能を付けない。種別と借り先だけを述べる。セカンダリー・ドミナントは「向かう先のドミナント」とだけ述べ、T・S・D を付けない（#101 の所有者の判断 A1、#102 の受け入れ条件）。
- **別の立場**: 4 つに分け、vi・iii と IV → I の IV を「トニックの延長」とする（Hutchinson）。iii を前後関係で D と読む（世界大百科事典、SoundQuest）。S を、ドミナントの前に置く和音として「プレドミナント」と呼ぶ（英語圏の教科書）。
- **出どころ**:
  - ヤマハ Music Pal: https://jp.yamaha.com/services/music-pal/study/chord/cadence/index.html
  - 洗足オンラインスクール: https://www.senzoku-online.jp/theory/chord/ch01_03.html
  - 世界大百科事典「和声」: https://kotobank.jp/word/%E5%92%8C%E5%A3%B0-154108
  - Open Music Theory 第 1 版: https://openmusictheory.github.io/harmonicFunctions.html
  - Hutchinson: https://musictheory.pugetsound.edu/mt21c/HarmonicFunction.html
  - SoundQuest: https://soundquest.jp/quest/chord/chord-mv1/chord-functions/2/

## 2. セカンダリー・ドミナント

- **採る立場**: 形で判定する。調の外の音を含む長三和音か属七で、根音の完全 5 度下が、調の長三和音か短三和音（I を除く）の根音であるもの。ローマ数字は「V7/V」のように、向かう先を `/` のあとに書く。
- 期待した先へ進まなくても、名前を保つ。解決の仕方を「期待どおり」「変則」として別に述べる。
- 解決は、セカンダリー・ドミナントと V（5 度の長三和音か属七。バスとテンションは問わない）について述べる。次の和音が期待した先なら「期待どおり」、ほかの和音なら「変則」。期待した先は、次の和音の読みがダイアトニックで、向かう先の度数（V では I）にあり、3 度のある三和音か七の和音であること。C で D7 → G/B は期待どおり、D7 → Gm（借用和音）と A7 → D7 は変則（#103 の所有者の判断）。
- 次の和音が sus の和音や 6 の和音なら、解決を述べない（G7 → Csus4）。同じドミナントが続くとき（V → V、向かう先が同じセカンダリー・ドミナントどうし）は、続いた最後の和音でだけ述べる（G → G7 → C は G7 だけ。#103 の所有者の判断）。
- **別の立場**: 5 度下の和音へ進んだときだけセカンダリー・ドミナントと呼び、進まなければ別の名前（クオリティ・チェンジ）で呼ぶ（SoundQuest）。V/V → IV のように、向かう先の代わりになる和音（ポップスの IV や ♭VII）へ進むことを、偽終止になぞらえて「偽の解決（resolve deceptively）」と呼び分ける（Hutchinson の変則的な解決）。
- **出どころ**:
  - Hutchinson（判定の手順）: https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html
  - Hutchinson（減三和音はトニック化しない）: https://musictheory.pugetsound.edu/mt21c/Tonicization.html
  - Hutchinson（変則的な解決。偽の解決）: https://musictheory.pugetsound.edu/mt21c/IrregularResoltionsOfSecondaryDominants.html
  - SoundQuest: https://soundquest.jp/quest/chord/chord-mv2/secondary-dominant/3/

## 3. 借用和音

- **採る立場**: 語は「借用和音」を使う。同じ主音の別の旋法にある和音を、長調で使うこと。セカンダリー・ドミナントは含めない。
- 借り先は、和音を含む同主旋法の集合として計算する。含むかは、ダイアトニックと同じ音を綴りで見る（1。#102 の所有者の判断）。自然短音階が含むなら「同主短調から」、含まなければ、含む旋法の名前（ミクソリディアンなど）を示す。
- **別の立場**: 借用を同主短調からに限る（Hutchinson、洗足）。「借用和音」をセカンダリー・ドミナントにも使い、同主短調からの借用を「準固有和音」と呼ぶ（ヤマハ）。♭VII を旋法ではなく IV の IV として説明する（Open Music Theory 第 1 版）。
- **出どころ**:
  - Hutchinson: https://musictheory.pugetsound.edu/mt21c/ModeMixtureSection.html
  - 洗足: https://www.senzoku-online.jp/theory/chord/ch04.html
  - Berklee の用語集（Modal Interchange）: https://college.berklee.edu/core/glossary.html
  - SoundQuest（借り先が 4 音では決まらないこと）: https://soundquest.jp/quest/chord/chord-mv6/ttj-modal-interchange/
  - Open Music Theory 第 1 版: https://openmusictheory.github.io/popRockHarmony-plagal.html

## 4. 終止

- **採る立場**: フレーズの終わりに来たときだけ終止と呼ぶ。正格終止（V → I）、半終止（V で終わる）、変格終止（IV → I）、偽終止（V から I 以外へ）。フレーズの途中の同じ動きは、終止と呼ばない。
- フレーズが終止で終わらなくてもよい。
- 終止を付けるのは、満ちたフレーズ（4 小節）の最後の枠だけ。作りかけの末尾のフレーズには、半終止も付けない。あとで取り消す名前を付けないため（ADR-0012。#103 の所有者の判断）。
- I・IV・V は、ダイアトニックの和音で、その度数にあり、3 度のある三和音か七の和音であるもの。バスとテンションは問わない（G/B、G7/F、G7(9) は V、C/E は I）。sus の和音と 6 の和音は含めない（Gsus4 と G6 は V でなく、Fsus2 → C は変格終止でない）。借用和音は含めない（G7 → Cm は偽終止、Fm → C は名前なし。#103 の所有者の判断）。
- 旋律を持たないので、完全・不完全の正格終止は分けない（#103 の所有者の判断）。
- 語は「終止」を使い、「カデンツ」は使わない。
- **別の立場**: 「カデンツ」を、T–S–D–T のような機能の型の意味で使う（ヤマハ、洗足、世界大百科事典）。両方の和音が基本形で、主音が最上声にあるものだけを完全正格終止とし、ほかと vii° → I を不完全正格終止とする（Hutchinson の完全正格終止）。I の第 2 転回形（終止の 6/4。C/G）を、続く V と同じドミナントの機能とみる（Hutchinson の終止の 6/4）。フレーズの長さを 4 小節に決めず、3〜8 小節のフレーズを旋律や歌詞の区切りで認める（Hutchinson のポップスのフレーズ）。
- **出どころ**:
  - Hutchinson（定義）: https://musictheory.pugetsound.edu/mt21c/cadences.html
  - Hutchinson（ポップスのフレーズ。3〜8 小節のフレーズ）: https://musictheory.pugetsound.edu/mt21c/PhraseSection.html
  - Hutchinson（完全正格終止）: https://musictheory.pugetsound.edu/mt21c/PerfectAuthenticCadence.html
  - Hutchinson（終止の 6/4）: https://musictheory.pugetsound.edu/mt21c/TheCadentialSixFourChord.html
  - Open Music Theory 第 1 版: https://openmusictheory.github.io/cadenceTypes.html
  - 大辞泉「終止」: https://kotobank.jp/word/%E7%B5%82%E6%AD%A2-526784

## 5. テンション

- **採る立場**: 調の音階の音のうち、和音の音の半音上にある音は、和音に重ねると濁りやすい、と注意として示す。禁止はしない。属七では 11 度だけを注意とする。IIm7 の 13 度は注意にしない。調の外のテンションは注意の対象にしない（#108）。注意がないことは、濁らないという意味ではない。
- **別の立場**: これらを「避ける音（アヴォイド・ノート）」として、和音に使わない規則にする。IIm7 の 13 度も避ける（Berklee 系の一部）。
- **出どころ**:
  - SoundQuest（半音上の原則、属七の扱い）: https://soundquest.jp/quest/chord/chord-mv4/tension-chord-avoid/2/
  - SoundQuest（IIm7 の 13 度）: https://soundquest.jp/quest/chord/chord-mv6/cst-major-chordscales/
  - Berklee Online（教え方の用語であること）: https://online.berklee.edu/takenote/simplifying-jazz-harmonic-theory-an-interview-with-suzanna-sifter/
  - Music Theory Online の書評（Levine の「handle with care」）: https://mtosmt.org/issues/mto.00.6.1/mto.00.6.1.rawlins.html
  - 洗足（和音の種類ごとのテンション）: https://www.senzoku-online.jp/theory/chord/ch07_01.html

## 6. 表記

- **採る立場**:
  - sus は、3 度を 4 度か 2 度に置き換えた和音。ほかの三和音と同じく完全 5 度を持つ（5 度のない C–F–B♭ は sus と呼ばない）。4 度と 2 度の両方を持つものは sus4 でも sus2 でもない。3 度があれば、4 度は 11 度（7 度がなければ add4。add4 は表示の形の一覧にまだなく、この一覧が最初の語彙）と書く。
  - 7 度のない 9 度は add9。「9」と書いたら 7 度を含む。
  - テンションは括弧で書く（Dm7(9)）。
  - 和音名は、根音、種類、テンションの括弧、`/` とバスの音名の順に書く。変化記号は `♯` と `♭`、重変化は `𝄪` と `𝄫`。種類の表示の形は `C`、`Cm`、`Cdim`、`Caug`、`C7`、`Cmaj7`、`Cm7`、`Cm7(♭5)`、`Cdim7`、`Csus2`、`Csus4`、`C7sus4`、`Cadd9`。9 の和音も括弧で書く（`C7(9)`、`Cmaj7(9)`、`Cm7(9)`。#78 の判断 B1）。テンションは 9、♭9、♯9、11、♯11、13、♭13 で、`7`・`maj7`・`m7` のあとに、音程の数の小さい順にコンマで並べる（`C7(9,13)`）。
  - 読むときは、表示の形に加えて、よく使われる別名だけを受ける（#78 の判断 A1）。長 7 の `M7`・`Δ7`（U+0394）、短の `min`、減の `°`（U+00B0）、増の `+`、半減 7 の `m7♭5`・`ø7`（U+00F8）、変化記号の `#`・`b`、重変化の `x`・`##`・`♯♯`・`bb`・`♭♭`。
  - 受けない書き方は、読めなかった所を示すエラーを返す。曖昧な書き方（`C-`）、表示の形にない書き方（`C9`、`CM`、`Csus`）、空白、順序違いや同じ音程の数の重複のテンション（`C7(13,9)`、`C7(9,♭9)`）。
  - ローマ数字は、変化記号、度数の数字、種類とテンション、`/` と向かう先の順に書く。大文字は長・増、小文字は短・減。減は `°`（U+00B0。`vii°`、`vii°7`）、半減 7 は `ø7`（U+00F8。`viiø7`）、増は `+`（`♭III+`）。調の音でない根音は、長音階の度数に変化記号を前に付ける（`♭VII`、`♭VI`、`♭III`）。七度とテンションは和音名と同じ書き方を後ろに付け、短の `m` は小文字が表すので書かない（`ii7`、`Imaj7`、`V7(9)`）。セカンダリー・ドミナントは `V7/V`（2）。3 度のない和音は度数に接尾辞を付ける（`Vsus4`。大文字・小文字は下の独自の規約）。転回形とバスは書かない（#78 の判断 C1、#80 の受け入れ条件）。
  - 実装の側で決めた点（#81 で決めた。所有者の判断ではない。#78 の判断の範囲で、文面が決めていなかった所を埋めた）:
    - テンションの括弧は、`7`・`maj7`・`m7` のあとだけに付く。
    - 別名は種類の中の語にも効く（`Cmin7`、`C°7`、`CM7(9)`、`Cm7b5`）。
    - 根音の幹音は大文字だけ（`b` は変化記号）。
    - 表示のときの読み替えは、変化の向きに隣の幹音へ移す（B𝄪 は C♯、F♭♭ は E♭）。根音を読み替えたら、バスは読み替えた根音から同じ音程で書く（B♭♭/D♭ は A/C♯）。根音を読み替えないときだけ、バスを単独で同じ規則で読み替える。
  - 実装の側で決めた点（#80 で決めた。所有者の判断ではない。#78 の判断 C1 の範囲で、文面が決めていなかった所を、最も小さい解釈で埋めた）:
    - 向かう先は 1 つだけで、度数 2〜6 の `ii`、`iii`、`IV`、`V`、`vi`。大文字・小文字はその度数の長調の三和音に合わせ、変化記号を付けない。`/I`、`/vii`、`/II`、`V7/V/V` は受けない。これはローマ数字で書ける範囲の決めで、セカンダリー・ドミナントの判定（2）ではない。
    - `/` の前の度数は、向かう先の根音の長音階で数える（C で `V/ii` は、D の長音階の 5 度の A）。短調の主音から見た位置は、まだ扱わない。
    - 読むのは表示の形だけ。和音名の別名（`M7`、`Δ7`、テンションの `b9`・`#11`）も、RomanText の書き方（変化記号の `b`・`#`、減の `o`、半減の `/o`、転回形の数字）も受けない。
- **出どころ**:
  - Hutchinson（sus）: https://musictheory.pugetsound.edu/mt21c/SimpleSusChords.html
  - Hutchinson（数字の規則）: https://musictheory.pugetsound.edu/mt21c/ChordSymbolSpecifics.html
  - 洗足（括弧）: https://www.senzoku-online.jp/theory/chord/ch07_02.html
  - Harte ほか（ISMIR 2005。足す音を括弧の中にコンマで並べること、`/` のあとのバス、`B` と `b` を大文字・小文字で分けること）: https://archives.ismir.net/ismir2005/paper/000080.pdf
  - Tymoczko、Gotham、Cuthbert、Ariza（RomanText。ISMIR 2019。大文字は長、小文字は短、増の `+`、根音を上げ下げする前置の変化記号、`/` のあとの向かう先）: https://archives.ismir.net/ismir2019/paper/000012.pdf 。減と半減の記号（RomanText は `o` と `/o`）は、この出どころと違う（`.claude/skills/standards/index.md`）
- **独自の規約**（定めた規則が教材に見つからないもの）:
  - 3 度のない和音のローマ数字は、度数に接尾辞を付ける（Vsus4）。大文字・小文字は、その度数の調の三和音に合わせる。変化した度数では大文字で書く。その度数に置く長三和音に合わせる（C で `♭VIIsus4`、`♭IIIsus2`。#82 の所有者の判断）。
  - 和音名の根音が重変化記号、C♭、F♭ になるときは、表示だけ異名同音に読み替える（B♭♭ は A、C♭ は B、F♭ は E）。内部の綴りは変えない。根音を読み替えたら、バスは読み替えた根音から同じ音程で書く。調のない表示だけ。調の中（ローマ数字と一緒に出すとき）は読み替えない（D♭ で ♭VII は C♭、♭VI は B𝄫。#82 の所有者の判断）。

## 確かめていないこと

- ローマ数字の減の `°` と半減 7 の `ø7` の出どころ（開いた RomanText の論文は `o` と `/o`）。
- Open Music Theory 第 2 版の記述（サイトが取得を拒む）。
- Levine、Nettles & Graf、Berklee Book of Jazz Harmony の原文（書評と教材サイトの引用で読んだ）。
