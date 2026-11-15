# 0010. コンテナは、distroless の API と nginx の画面の 2 つのイメージ

- 日付: 2026-11-15
- 状態: Accepted

## 文脈

ADR-0008 は、手元のコマンドとコンテナの両方で動かせるようにし、CI でイメージのビルドとスモークを走らせると決めた。どのイメージで、どう版を固定し、どう確かめるかは決めていなかった（#28）。

- Docker の公式の文書（https://docs.docker.com/build/building/best-practices/ 、2026-11-15 に確認）は、マルチステージのビルド、`.dockerignore`、root 以外のユーザー、ベースイメージの版の固定を勧めている。タグは付け替えられるので、供給網を固めるには digest で固定し、更新は自動化の道具に任せる、としている。
- Go と Node の版の出どころは `.tool-versions` だけにしている（ADR-0004）。
- 開発に使うこの環境には Docker のデーモンがなく、イメージは CI でしかビルドできない。

## 決定

1. **1 つの Dockerfile から 2 つのイメージを作る。** `api` は Go のサーバーを distroless の `static` の `nonroot` に載せる（Debian の版と digest は `Dockerfile`）。`web` は nginx（`nginx-unprivileged`）で、ビルドした画面を配信し、`/api` を `api` へ中継する。どちらも root 以外のユーザーで動く。
2. **ビルドに使う Go と Node の版は、`.tool-versions` から build arg で渡す。** Dockerfile に版を書かない。`make image` と `make up` が渡す。渡さずにビルドすると、イメージが見つからずに止まる。
3. **実行用のベースイメージ（nginx と distroless）は、タグと digest で固定する。** Dependabot の docker が更新する。
4. **HEALTHCHECK は置かない。** distroless にはシェルがなく、置くにはサーバーに確かめるためのフラグを足す必要がある。起動の確認は、スモークの検査が `health` を決まった回数だけ待って行う。
5. **画面は、存在しないファイルを 404 にする。** クライアント側の経路がないので、ページへのフォールバックは置かない。
6. **compose は、`api` を公開せず、`web` だけを公開する。** 中継先は `API_UPSTREAM`（開発サーバーと同じ名前）。nginx はリクエストごとに名前を引くので、`api` より先に起動してよい。
7. **確かめ方。**
   - `make dockerfile-lint`（hadolint）は、hadolint か動いている Docker が要るので `make check` には入れず、CI の必須のジョブで走らせる。
   - CI の `image` のジョブが、BuildKit のビルドチェック、`make up`、`make smoke`、ログ、`make down` を走らせる。必須のチェックにはしない。

## 検討した選択肢

- **画面も Go のサーバーから配信する（イメージ 1 つ）**: イメージと中継の設定が減る。ADR-0008 が画面を配信するイメージを別に置くと決めており、配信の規則（キャッシュ、ヘッダ）をサーバーのコードに持ち込むことになる。
- **HEALTHCHECK を置く**: compose で `api` の準備ができてから `web` を起動できる。サーバーに確かめるためのフラグが要り、振る舞いの変更になる。今は利点が小さい。
- **ビルドに使うイメージも digest で固定する**: 固定は強くなる。版の出どころが `.tool-versions` と Dockerfile の 2 つになり、Dependabot の更新が `.tool-versions` と食い違う。
- **タグだけで固定する**: 更新の手間がない。同じタグで中身が変わっても気づけない。

## 結果

- 手元に Docker があれば、`make up` で本番に近い形で動かせる。この環境では動かせないので、イメージのビルドと起動は CI で確かめる。
- nginx の設定とスモークの検査は、手元の nginx と Go のサーバーでも確かめられる。
- 実行用のイメージの更新は、Dependabot の PR と `image` のジョブで確かめてからマージする。
- 再検討の条件: 公開先に配備するとき（ヘルスチェックの経路、セキュリティのヘッダ、キャッシュの規則を、配備先に合わせて決める）。
