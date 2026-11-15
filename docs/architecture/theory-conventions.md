# 採っている流儀と出どころ

エンジンの「読み」と「表記」（`docs/architecture/engine.md` のまとまり 4 と 2）が従う規則。和声の教材の間で立場が割れる点について、どの立場を採り、どこに書いてあるかを記す。規則を足す・変えるときは、ここに出どころを書く。

## 1. 機能

- **採る立場**: トニック（T）、サブドミナント（S）、ドミナント（D）の 3 つに分ける。長調では I・iii・vi が T、ii・IV が S、V・vii° が D。iii と vi は「トニックの仲間」と呼ぶ。IV → I は S から T への動きとして述べる。
- ドミナントの強さは、導音を含むか（属七では三全音を含むか）で述べる。導音のない和音を、V と同じ強さのドミナントとは言わない。
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
- **別の立場**: 5 度下の和音へ進んだときだけセカンダリー・ドミナントと呼び、進まなければ別の名前（クオリティ・チェンジ）で呼ぶ（SoundQuest）。
- **出どころ**:
  - Hutchinson（判定の手順）: https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html
  - Hutchinson（変則的な解決）: https://musictheory.pugetsound.edu/mt21c/IrregularResoltionsOfSecondaryDominants.html
  - SoundQuest: https://soundquest.jp/quest/chord/chord-mv2/secondary-dominant/3/

## 3. 借用和音

- **採る立場**: 語は「借用和音」を使う。同じ主音の別の旋法にある和音を、長調で使うこと。セカンダリー・ドミナントは含めない。
- 借り先は、和音を含む同主旋法の集合として計算する。自然短音階が含むなら「同主短調から」、含まなければ、含む旋法の名前（ミクソリディアンなど）を示す。
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
- 語は「終止」を使い、「カデンツ」は使わない。
- **別の立場**: 「カデンツ」を、T–S–D–T のような機能の型の意味で使う（ヤマハ、洗足、世界大百科事典）。
- **出どころ**:
  - Hutchinson（定義）: https://musictheory.pugetsound.edu/mt21c/cadences.html
  - Hutchinson（ポップスのフレーズ）: https://musictheory.pugetsound.edu/mt21c/PhraseSection.html
  - Open Music Theory 第 1 版: https://openmusictheory.github.io/cadenceTypes.html
  - 大辞泉「終止」: https://kotobank.jp/word/%E7%B5%82%E6%AD%A2-526784

## 5. テンション

- **採る立場**: 調の音階の音のうち、和音の音の半音上にある音は、和音に重ねると濁りやすい、と注意として示す。禁止はしない。属七では 11 度だけを注意とする。IIm7 の 13 度は注意にしない。
- **別の立場**: これらを「避ける音（アヴォイド・ノート）」として、和音に使わない規則にする。IIm7 の 13 度も避ける（Berklee 系の一部）。
- **出どころ**:
  - SoundQuest（半音上の原則、属七の扱い）: https://soundquest.jp/quest/chord/chord-mv4/tension-chord-avoid/2/
  - SoundQuest（IIm7 の 13 度）: https://soundquest.jp/quest/chord/chord-mv6/cst-major-chordscales/
  - Berklee Online（教え方の用語であること）: https://online.berklee.edu/takenote/simplifying-jazz-harmonic-theory-an-interview-with-suzanna-sifter/
  - Music Theory Online の書評（Levine の「handle with care」）: https://mtosmt.org/issues/mto.00.6.1/mto.00.6.1.rawlins.html
  - 洗足（和音の種類ごとのテンション）: https://www.senzoku-online.jp/theory/chord/ch07_01.html

## 6. 表記

- **採る立場**:
  - sus は、3 度を 4 度か 2 度に置き換えた和音。ほかの三和音と同じく完全 5 度を持つ（5 度のない C–F–B♭ は sus と呼ばない）。4 度と 2 度の両方を持つものは sus4 でも sus2 でもない。3 度があれば、4 度は 11 度（7 度がなければ add4）と書く。
  - 7 度のない 9 度は add9。「9」と書いたら 7 度を含む。
  - テンションは括弧で書く（Dm7(9)）。
- **出どころ**:
  - Hutchinson（sus）: https://musictheory.pugetsound.edu/mt21c/SimpleSusChords.html
  - Hutchinson（数字の規則）: https://musictheory.pugetsound.edu/mt21c/ChordSymbolSpecifics.html
  - 洗足（括弧）: https://www.senzoku-online.jp/theory/chord/ch07_02.html
- **独自の規約**（定めた規則が教材に見つからないもの）:
  - 3 度のない和音のローマ数字は、度数に接尾辞を付ける（Vsus4）。大文字・小文字は、その度数の調の三和音に合わせる。
  - 和音名の根音が重変化記号、C♭、F♭ になるときは、表示だけ異名同音に読み替える（B♭♭ は A、C♭ は B、F♭ は E）。内部の綴りは変えない。

## 確かめていないこと

- セカンダリー・ドミナントの向かう先から減三和音（vii°）を除くことの出どころ。
- Open Music Theory 第 2 版の記述（サイトが取得を拒む）。
- Levine、Nettles & Graf、Berklee Book of Jazz Harmony の原文（書評と教材サイトの引用で読んだ）。
