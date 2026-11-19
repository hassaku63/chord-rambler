# 用語集

画面と文書で使う語の定義。語の範囲が教材によって違うものは、`docs/architecture/theory-conventions.md` に採った立場を書いている。

## 訳語の方針

- 英語の複合語を語ごとに直訳せず、日本語の楽典・和声の教材と辞典で定まった語を使う。triad quality は「三和音の質」ではなく「三和音の種類」（長・短・減・増）、interval quality は「音程の種類」（完全・長・短・増・減）、chord quality は「和音の種類」。
- 日本語と英語で取り違えやすい語は、下の「取り違えやすい語」に書く。
- 文書、Issue、PR で、英語を見失いやすい語を使うときは、最初に出たところで英語を括弧で添える（「同主短調（parallel minor）」）。添える英語は、表の「英語」の列に従う。
- 直訳で新しい訳語を作らない。新しい語が要るときは、先にこの用語集に足す。
- 表の「英語」の列は、音楽の語では英語の音楽理論の語。定まった英語の語が見当たらないものは、コードと契約が使う英語を「（コードと契約の語）」と添えて書く。このプロダクトの語では、コードと契約が使う英語で、定まった音楽の語ではない。
- 表の「コード・契約」の列は、Go の識別子（`reading.Kind`）、契約（`docs/api/openapi.yaml`）の欄と値（契約 `kind: secondary-dominant`）、画面の識別子（`frontend/src/features/progression/terms.ts` の `TERMS.kind` など）。識別子のない語は「—」。

### 取り違えやすい語

| 語 | 取り違え |
| --- | --- |
| 同主調と平行調 | 英語の parallel key は、主音が同じ長調と短調で、日本語では同主調（C major と C minor）。英語の relative key は、調号が同じ長調と短調で、日本語では平行調（C major と A minor）。「平行」を parallel の訳と思って取り違えない |
| 度数と度 | 度数は scale degree（主音から数えた音階上の位置。ローマ数字で書く）。度は音程の大きさ（interval number。3 度の 3）。度数を音程の大きさに使わない |
| 主音とトニック | どちらも英語は tonic。主音は音（`material.Scale.Tonic`）、トニックは機能（`reading.Tonic`） |
| 種類と種別 | 和音の種類（m7、sus4）と、読みの種別（ダイアトニック、借用和音など）は別のもの。コードはどちらも `kind` と呼ぶ（`notation` の `kind`、`reading.Kind`、契約 `kind`）。画面の `ROLE_GROUPS` の `kind` は、役割のまとまり（構造的・色合い）を指す |
| 解決 | 3 つを指す。和音の進み方（`reading.Resolution`）、その「期待どおり」の値の画面の語（`TERMS.resolution`）、構造的な役割の 1 つ（`progression.RoleResolve`、契約 `resolve`） |
| 形 | 和音から根音を除いたもの（`material.Shape`）、役割の条件の「3 度のある三和音か七の和音」（`reading.ConditionShape`、契約 `shape`）、セカンダリー・ドミナントの形（`relation.SecondaryDominantShape`） |
| 区間 | 曲の中の大きな部分（`progression.Section`）と、枠の定義の「1 つの和音が鳴る区間」（時間の幅） |
| モード | `docs/product/concept.md` では音楽の旋法（mode）、`docs/product/screen-promise.md` では画面の操作のモード。この用語集の音楽の語は「旋法」 |

### 英語との対の出どころ

- 辞典が日本語と英語の対を書いているもの（コトバンク。2026-11-19 に取得）:
  - 平行調 relative key: ブリタニカ国際大百科事典 小項目事典（「同じ調号をもつ長調と短調をいう。ハ長調とイ短調はその例」）。https://kotobank.jp/word/平行調
  - 導音 leading note, leading tone: ヤマハミュージックメディア 音楽用語ダス。https://kotobank.jp/word/導音
  - 三全音 tritone: 日本大百科全書（ニッポニカ）。https://kotobank.jp/word/三全音
  - 変格終止 plagal cadence: ブリタニカ国際大百科事典 小項目事典。https://kotobank.jp/word/変格終止
  - 三和音 triad、転回形 inversion、属七の和音 dominant seventh chord、和声的短音階 harmonic minor scale、教会旋法 church mode: 音楽用語ダス。https://kotobank.jp/word/三和音 、https://kotobank.jp/word/転回形 、https://kotobank.jp/word/属七の和音 、https://kotobank.jp/word/和声的短音階 、https://kotobank.jp/word/教会旋法 。三和音に長三和音・短三和音・増三和音・減三和音の別があることは、同じ「三和音」の項の精選版 日本国語大辞典
  - 同主調（日本語の意味だけ。「ハ長調とハ短調（同主調）」）: 音楽用語ダス「近親調」。https://kotobank.jp/word/近親調
- 英語の語を、`docs/architecture/theory-conventions.md` が挙げる出どころで確かめたもの。日本語との対は、定義が一致することによる:
  - authentic cadence、half cadence、plagal cadence、deceptive cadence: Hutchinson（4 の「定義」）
  - mode mixture、parallel minor、borrowed chord: Hutchinson（3）
  - Modal Interchange（parallel を「同じ主音（pitch axis）の」の意味で使う）、chord tone、inversion: Berklee の用語集（3）
  - chord quality: Hutchinson（2 の「判定の手順」）
  - irregular resolution: Hutchinson（2 の「変則的な解決」）
  - sus（suspended）: Hutchinson（6）
  - chord symbol: Hutchinson（6 の「数字の規則」）
  - 和音の種類を `kind` と呼ぶこと: MusicXML 4.0（6）
- ほかの対は、出どころを確かめていない。同主調 parallel key の対を書いた辞典は見つけていない（英語の parallel の意味は、上の Berklee と Hutchinson）。

## 音楽の語

| 語 | 英語 | コード・契約 | 意味 |
| --- | --- | --- | --- |
| 主音 | tonic | `material.Scale.Tonic`。契約 `key`（長調の主音の音名） | 調の中心になる音 |
| 調 | key | `progression.Progression.Key`（`material.Scale`） | 主音と旋法の組（C major など） |
| 長調 | major key（旋法は Ionian） | `material.Ionian`。契約 `ionian`（画面の語は「長調」）。長調でない調は `relation.ErrNotMajor` | 長音階の調。旋法としてはイオニアン。読みは長調だけを扱う |
| 同主調 | parallel key | — | 主音が同じ長調と短調（C major と C minor） |
| 同主短調 | parallel minor | `reading.Reading.FromParallelMinor`。契約 `fromParallelMinor`。画面 `PARALLEL_MINOR` | 同じ主音の短調（C major に対する C minor） |
| 平行調 | relative key | — | 調号が同じ長調と短調（C major と A minor）。読みは使わない。英語の parallel key（同主調）と取り違えない |
| 旋法 | mode | `material.Mode`。契約 `Mode` | 音階の種類（長調、自然短音階、ミクソリディアン、ドリアンなど） |
| 教会旋法 | church modes（Ionian、Dorian、Phrygian、Lydian、Mixolydian、Aeolian、Locrian） | `material.Mode` の `Ionian`、`Dorian`、`Phrygian`、`Lydian`、`Mixolydian`、`Aeolian`、`Locrian`。契約 `ionian`、`dorian`、`phrygian`、`lydian`、`mixolydian`、`aeolian`、`locrian`。画面 `TERMS.mode` | 7 つの旋法。イオニアン（長調）、ドリアン、フリジアン、リディアン、ミクソリディアン、エオリアン（自然短音階）、ロクリアン。借用和音の借り先を言うときに使う |
| 自然短音階 | natural minor scale（旋法は Aeolian） | `material.Aeolian`。契約 `aeolian`（画面の語は「自然短音階」） | 旋法としてはエオリアン。借用和音の借り先がこれを含むとき、「同主短調から」と言う |
| 和声的短音階 | harmonic minor scale | — | 自然短音階の 7 度を半音上げた音階。エンジンは持たないので、借り先にならない（C で Bdim7 は事実だけ） |
| 同主旋法 | parallel modes | `relation.ModesHolding` | 主音が同じ旋法。借り先は、和音を綴りで含む同主旋法の集合 |
| 幹音 | natural note（letter） | `material.Letter` | 変化記号の付かない音名（C D E F G A B） |
| 綴り | spelling | `material.NoteName`。表示だけの読み替えは `notation.DisplaySpelling` | 音名を幹音と変化（♯・♭・𝄪・𝄫）で書いたもの。高さが同じでも、綴りの違う音（C♯ と D♭）は別の音名。調や旋法に含まれるかは綴りで比べる |
| 異名同音 | enharmonic | — | 綴りが違い、高さが同じ音（C♯ と D♭）。音高クラスに移すと同じになる |
| 音高クラス | pitch class | `material.PitchClass`、`material.PitchClassSet` | 綴りとオクターブを持たない音の高さ（0〜11。C = 0） |
| 音程 | interval | `material.Interval` | 2 つの音の隔たり。度と種類の組で書く（長 3 度） |
| 音程の種類 | interval quality | `material.Quality`（`Perfect`、`Major`、`Minor`、`Augmented`、`Diminished`。略号 P、M、m、A、d）。契約の音程は略号と度で書く（`P11`、`M3`） | 完全・長・短・増・減の 5 つ |
| 根音 | root | `material.Chord.Root` | 和音の 1 度にあたる構成音。和音は、根音と、根音からの音程で書く（`docs/architecture/domain-model.md` の「和音」） |
| 構成音 | members（コードと契約の語） | `material.Chord.Notes`、`query.MembersLabel`。契約 `members` | 和音を作る音。根音と、根音からの音程の音で、テンションを含む。構成音でないバス（C/D の D）は含まない |
| 和音の音 | chord tones | `reading.TensionCaution` の `ChordTone`。契約 `chordTone` | テンションの注意で見る、根音と 3 度・5 度・7 度の構成音 |
| 三和音 | triad | — | 根音に 3 度と 5 度の音を重ねた和音 |
| 三和音の種類 | triad quality | `material.Triad`（`TriadMajor`、`TriadMinor`、`TriadDiminished`、`TriadAugmented`。ほかに `TriadSus4`、`TriadSus2`、`TriadOther`）。`query.ErrNoTriadQuality` | 3 度と 5 度で決まる、長・短・減・増の 4 つ。エンジンの見方は、3 度を置き換えた sus4・sus2 と「その他」も同じ型で返す（`docs/architecture/domain-model.md` の「見方」） |
| 七の和音 | seventh chord | `material.Seventh`（七度の見方） | 三和音に、根音から 7 度の音を重ねた和音 |
| 主和音 | tonic chord | — | 主音を根音とする三和音。ローマ数字の I |
| 属七の和音（属七） | dominant seventh chord | — （`material.TriadMajor` と `material.SeventhMinor` の組） | 長三和音に短 7 度を重ねた和音（C で G7）。エンジンは度数によらず形で決める（`docs/architecture/domain-model.md` の「テンションの扱い」） |
| 和音の種類（コードタイプ） | chord quality、chord type | `notation` の公開しない型 `kind`、`notation.ErrChordKind`。MusicXML の `kind` | 和音名で根音のあとに書く部分が表すもの（m、7、maj7、sus4 など）。語彙と表示の形は `docs/architecture/theory-conventions.md` の 6。読みの種別とは別のもの |
| 和音名 | chord symbol | `notation.ParseChord`、`notation.FormatChord`。契約 `chord`、`ChordNameLabel` | 根音と種類で書いた和音の名前（Dm7）。調によらない |
| ローマ数字 | Roman numeral | `notation.ParseRoman`、`notation.FormatRoman`、`reading.RomanPosition`。契約 `roman`、`RomanLabel` | 和音を、主音から見た度数と種類で書いたもの（ii7、V7/V）。書き方は `docs/architecture/theory-conventions.md` の 6 |
| 度数 | scale degree | `material.Position.Degree` | 主音から数えた音階上の位置。ローマ数字で書く（ii7）。音程の大きさには使わない |
| 度 | interval number | `material.Interval.Number` | 音程の大きさを数える語（3 度、長 3 度）。文書でまとめて指すときは「音程の数」 |
| ダイアトニック | diatonic | `reading.Diatonic`。契約 `kind: diatonic` | 和音を作る音（根音と 2〜7 度の構成音）と 7 度のバスが、綴りで調の音階の音だけでできていること。テンションとほかのバスは見ない |
| 機能 | harmonic function | `reading.Function`。契約 `function`。画面 `TERMS.function`、`TERMS.functionMark` | 調の中での和音の働き。トニック（T）、サブドミナント（S）、ドミナント（D）の 3 つ |
| トニック | tonic | `reading.Tonic`。契約 `function: tonic` | 落ち着く働き。I と、その仲間の iii・vi |
| サブドミナント | subdominant | `reading.Subdominant`。契約 `function: subdominant` | トニックとドミナントのあいだで、動き出す働き。IV と ii |
| ドミナント | dominant | `reading.Dominant`。契約 `function: dominant` | トニックへ向かう緊張を作る働き。V と vii° |
| トニックの仲間 | — | — | I と 2 つの音を共有する iii と vi |
| 導音 | leading tone、leading note | `relation.HasLeadingTone` | 主音の半音下の音。主音へ進もうとする |
| 三全音 | tritone | `relation.HasTritone` | 全音 3 つ分の音程。属七の 3 度と 7 度が作る |
| 借用和音 | borrowed chord。この使い方は modal interchange（Berklee）、mode mixture（Hutchinson） | `reading.Borrowed`。契約 `kind: borrowed` | 同じ主音の別の旋法にある和音を、今の調で使うこと。セカンダリー・ドミナントは含めない |
| 借り先 | the parallel modes a borrowed chord is borrowed from（契約の説明の語） | `reading.Reading` の `Modes`。契約 `modes` | 借用和音を含む同主旋法の集合。自然短音階を含むなら「同主短調から」と言う |
| セカンダリー・ドミナント | secondary dominant | `reading.SecondaryDominant`、`relation.SecondaryDominantShape`。契約 `kind: secondary-dominant` | 主和音以外の調の和音に対して、ドミナントの形をした和音（V7/V など） |
| 向かう先 | target（コードと契約の語） | `reading.Reading` の `Target`、`material.Position.WithTarget`。契約 `target`。画面 `TARGET_NUMERALS` | セカンダリー・ドミナントが向かう和音の度数（2〜6。V7/V の /V） |
| 解決 | resolution | `reading.Resolution`。契約 `resolution`、`ResolutionLabel` | 緊張のある和音が、期待される和音へ進むこと。期待と違う先へ進むことを「変則」と呼ぶ |
| 期待どおり（の解決） | expected resolution | `reading.ResolutionExpected`。契約 `resolution: expected`（画面の語は「解決」） | 解決の値の 1 つ。次の和音が期待した先であること（`docs/architecture/theory-conventions.md` の 2） |
| 変則 | irregular resolution | `reading.ResolutionIrregular`。契約 `resolution: irregular` | 解決の値の 1 つ。次の和音が期待した先でないこと |
| 根音の動き | root motion | `relation.RootMotion`、`query.RootMotionLabel`。契約 `rootMotion` | 前の和音の根音から次の和音の根音への、上向きの音程。画面は契約の書き方（P4）のまま出す |
| 共通音 | common tones | `relation.CommonTones`、`query.CommonTonesLabel`。契約 `commonTones` | 前の和音と次の和音の両方で鳴る音。画面は、その数を出す |
| フレーズ | phrase | `progression.Phrase`、`relation.Place` | 進行のひと区切り。既定では 4 小節ごと |
| 終止 | cadence | `reading.Cadence`。契約 `cadence`、`CadenceLabel` | フレーズの終わりの和音の動き。正格終止、半終止、変格終止、偽終止 |
| 正格終止 | authentic cadence | `reading.AuthenticCadence`。契約 `cadence: authentic` | V → I で終わる終止 |
| 半終止 | half cadence | `reading.HalfCadence`。契約 `cadence: half` | V で終わる終止 |
| 変格終止 | plagal cadence | `reading.PlagalCadence`。契約 `cadence: plagal` | IV → I で終わる終止 |
| 偽終止 | deceptive cadence | `reading.DeceptiveCadence`。契約 `cadence: deceptive` | V から I 以外へ進んで終わる終止 |
| テンション | tension | `material.Tension`、`reading.TensionCautions`。契約 `tensionCautions` | 七の和音の上に重ねる 9 度・11 度・13 度の音 |
| sus | suspended（sus4、sus2） | `material.TriadSus4`、`material.TriadSus2` | 3 度を 4 度（sus4）か 2 度（sus2）に置き換えた和音 |
| add9 | add9 | — | 三和音に 9 度を足した和音。7 度を含まない。短三和音に足したものは m(add9) |
| 3 度のある三和音か七の和音 | — | `reading.Tertian`。契約の役割の条件 `shape` | 長・短・減・増の三和音で、2 度・4 度・6 度の構成音を持たない和音。7 度はあってもなくてもよい。役割の判定は、この形の和音でだけ機能を数える（`docs/architecture/domain-model.md` の 4、`docs/architecture/theory-conventions.md` の 7） |
| 6 の和音 | sixth chord（MusicXML の `major-sixth`、`minor-sixth`） | — | 長三和音か短三和音に長 6 度を足した和音（C6、Cm6）。ローマ数字の `I6` も 6 の和音で、転回形ではない（`docs/architecture/theory-conventions.md` の 6） |
| バス | bass | `material.Chord.Bass`、`material.Chord.WithBass` | 和音のいちばん低い音。指定がなければ根音。指定のバスは根音からの音程で持ち、構成音でなくてもよい（C/D） |
| 分数コード | slash chord | `material.Chord.WithBass` | バスの音を指定して書いた和音（C/E、F/G） |
| 転回形 | inversion | — | 分数コードのうち、バスが和音の音であるもの（C/E） |
| 声部 | voice | `performance` の公開しない関数 `voice` | 和音を鳴らすときの、一つひとつの音の線。演奏は、バス 1 音と上の声部に分けて置く（`docs/architecture/domain-model.md` の 9） |

## このプロダクトの語

| 語 | 英語 | コード・契約 | 意味 |
| --- | --- | --- | --- |
| 進行 | progression | `progression.Progression`。契約 `Progression` | 枠を並べたもの |
| 枠 | slot | `progression.Slot`。契約 `Slot`、`slots` | 1 つの和音が鳴る区間。和音、長さ（拍）、役割を持つ |
| 和音なし、まだ選んでいない | no chord、unchosen | `progression.NoChord`、`progression.Unchosen`。契約 `content: none`、`content: unchosen` | 和音を持たない枠の 2 つの中身。和音なしは鳴らさない枠、まだ選んでいないは和音をまだ決めていない枠 |
| 区間 | section | `progression.Section` | フレーズをまとめた、曲の中の大きな部分（A メロ、サビなど） |
| 役割 | role | `progression.Role`。契約 `Role`、`role`。画面 `TERMS.role`、`ROLE_GROUPS` | 枠が進行の中で果たす働き。構造的な役割と、色合いの役割がある |
| 構造的な役割 | structural roles | `progression.RoleStable`、`RoleLink`、`RolePrepare`、`RoleRise`、`RoleResolve`、`RoleClose`。契約 `stable`（安定）、`link`（つなぎ）、`prepare`（準備）、`rise`（高まり）、`resolve`（解決）、`close`（締め） | 機能とフレーズの位置から決まる役割。安定、つなぎ、準備、高まり、解決、締め |
| 色合いの役割 | colour roles | `progression.RoleSpread`、`RoleFloat`、`RoleWistful`、`RoleLinger`。契約 `spread`（広がり）、`float`（浮遊）、`wistful`（切なさ）、`linger`（余韻） | 印象で決める役割。広がり、浮遊、切なさ、余韻 |
| 役割に合うか | role fit | `reading.RoleFit`、`reading.ReadRoleFits`。契約 `roleFit`、`fit: fits`（合う）、`misfits`（合わない）、`undecided`（決められない） | 枠の読みが、置いた構造的な役割に合うかの判定。合う、合わない、決められない の 3 つ（`docs/architecture/theory-conventions.md` の 7） |
| 当たらなかった条件 | unmet conditions | `reading.RoleCondition`、`reading.ReadRoleVerdicts`。契約 `unmet`、`RoleCondition`。画面 `TERMS.unmet` | 「合わない」ときに、構造的な役割の合う条件のうち満たさなかったもの（`docs/architecture/theory-conventions.md` の 7） |
| 候補 | candidate | `query.Candidate`、`query.SlotCandidate`。契約 `Candidate`、`SlotCandidate` | ある枠に置ける和音の提案。理由を伴う |
| 変形 | variation | `query.Variation`。契約 `Variation` | 同じ根音のまま、6 度・7 度・3 度の置き換え（sus）・テンションを変えた和音。長・短の入れ替えや、減・増への変更は含めない（C から Csus4 は変形、C から Cm は変形でない） |
| 読み | reading | `reading.Reading`、`reading.ChordReading`、`reading.SlotReading`。契約 `reading`、`ProgressionReading` | 和音を調と前後の和音の中に置いて付ける名前（機能、借用、終止など） |
| 主な読み | main reading | `reading.ChordReading` の `Main`（別の読みは `Others`） | 画面に出す読み。別の読みは、開いて見る |
| 種別 | kind（of a reading） | `reading.Kind`。契約 `kind`、`CandidateKind`。画面 `TERMS.kind` | 読みが和音に付ける名前の分け方。ダイアトニック、借用和音、セカンダリー・ドミナント、事実だけの 4 つ。和音の種類（m7、sus4）とは別のもの |
| 事実だけ | facts-only | `reading.FactsOnly`。契約 `kind: facts-only`（札を出さない） | 種別の 1 つ。名前を付けず、位置と構成音の事実だけを述べる（C で B♯ の長三和音） |
| 札 | label | `query` の `…Label` の型（`query.ReadingLabel` など）。契約の `…Label` の型 | 読みや問い合わせが返す事実の 1 項目。根拠の種類を持つ（和音名、ローマ数字、読み、解決、終止など） |
| 根拠の種類 | basis | `query.Basis`。契約 `Basis`、`basis`。画面 `TERMS.basis` | 説明の項目が何に基づくか。定義、計算、流儀、慣習、連想の 5 つ |
| 定義 | definition | `query.Definition`。契約 `basis: definition` | 根拠の種類の 1 つ。音の素材が定めること（和音の構成音、和音名） |
| 計算 | computation | `query.Computation`。契約 `basis: computation` | 根拠の種類の 1 つ。流儀によらず計算できる事実（調の外の音、主音から見た位置） |
| 流儀 | school | `query.School`。契約 `basis: school` | 読みが従う、特定の教え方の規則の組 |
| 慣習 | convention | `query.Convention`、`catalog.Convention`。契約 `basis: convention` | よく使われる進行や、移り変わりの好み |
| 移り変わりの好み | preference | `catalog.FunctionTransitions`、`catalog.Preference`（`Common`、`LessCommon`、`Unstated`）、`query.PreferenceLabel`。契約 `preference: common`、`less-common`、`unstated` と `style: classical`。カタログの JSON は `common`、`less_common`、`unstated` | 機能から機能への動きが、どれだけ使われるか。段階は「よく使われる」「より少ない」「出どころに言明なし」。様式の範囲（いまはクラシック）と一緒に言う（`docs/architecture/domain-model.md` の 5） |
| 連想 | association | —（識別子はまだない） | 印象と和声の手段との結びつき |
| 体験 A | — | — | グラフ型。末尾に枠を足して、候補から和音を選ぶ |
| 体験 B | — | — | 計画型。役割つきの枠を先に並べて、埋める |
| キャンバス | canvas | 画面 `Canvas` | 画面の区画。進行を、和音のノードと、それをつなぐ線で見せる（`docs/product/screen-promise.md`） |
| 帯 | strip | 画面の区画の class `strip` | 画面の区画。枠を時間の順に並べ、小節とフレーズの区切りを見せ、進行を鳴らす |
| 文脈の欄 | context panel | 画面 `ContextPanel` | 画面の区画。選んでいるもの（枝、本線のノード、枠）の理由と、それに対してできることを出す |
| 探索のタブ | explore tab | 画面のタブ `explore` | 体験 A を行うタブ。計画のタブと、同じ進行を扱う |
| 計画のタブ | plan tab | 画面のタブ `plan` | 体験 B を行うタブ。探索のタブと、同じ進行を扱う |
| ノード | node | — | キャンバスの上の和音の印 |
| 本線 | main line | — | キャンバスの上で、確定した枠のノードを順につないだもの。進行と同じ並び |
| 枝 | branch | 画面 `Branches`、`useBranches` | 探索のタブで、本線の末尾から伸びる候補のノード |
| 探索の足跡 | footprint | 画面 `Footprint` | 1 歩前の分岐で、確定しなかった枝。薄く残し、選び直せる |
