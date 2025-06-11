# 📣 descord-srv-go

Minecraft サーバのログを監視し、Discord Webhook にログイン・ログアウトの通知を送る Go アプリケーションです。

## ✅ 構成

```
.
├── .devcontainer             # vscode DevContainer構成ファイル
├── build/
│   ├── Dockerfile.dev        # 開発用 Dockerfile
│   └── Dockerfile.prod       # 本番用 Dockerfile
├── docker-compose.yml        # コンテナ定義
├── go.mod                    # Go モジュール定義
├── pkg/
│   └── cmd/
│       ├── main.go           # 本体コード
│       └── main_test.go      # テストコード
└── scripts/
    └── install_go_tools.sh   # Go 開発ツールのインストーラ
```

## 🚀 セットアップ手順

### 2. DevContainer の起動

.devcontainerがあるルートフォルダに移動
`ctrl + shift + p`から`>Dev Containers: Rebuild and Reopen in Container`を開くと開発環境が立ち上がる

### 3. 環境変数の設定

`.env` または `docker-compose.yml` に以下のように指定してください：

```env
GOLANG_CONTAINER_NAME=discord-srv-go-container
GOLANG_ROOT_PATH=/go/src/github.com/taiki2523/app
MINECRAFT_DATA_PATH=/path/to/minecraft-data
LOG_FILE=/data/logs/latest.log
LOG_LEVEL=debug
MINECRAFT_LOG_PATH=/data/logs/latest.log
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxxxx/yyyyy
HEALTH_INTERVAL=1h
```
`GOLANG_ROOT_PATH`は、DevContainerのworkspaseFoldrerです。devcontainer.jsonに記載。

`MINECRAFT_DATA_PATH`は`/data`のマウントパスです。

`LOG_FILE` は、Minecraft サーバがログを出力するパスです。

## 🔎 開発・テスト

### 開発用起動

```devcontainer bash
go run ./pkg/cmd/...
```

### ユニットテストの実行

```devcontainer bash
go test -v ./pkg/cmd/...
```

※ `DISCORD_WEBHOOK_URL` を指定すると、統合テストで実際に通知が送信されます。

## 🔧 テスト内容

* `TestProcessLogLine`: ログ行の解析と通知内容の検証
* `TestRunWithNotifier_FileNotFound`: ログファイルが存在しない場合のリトライ検証
* `TestDiscordNotification_Integration`: Discord 通知の統合テスト（環境変数必須）

## 📌 注意点

* `MINECRAFT_LOG_PATH` に指定されたファイルが存在しない場合、10回（デフォルト）までリトライします。
* コンテナ内で `/data/logs/latest.log` を参照するため、Minecraft サーバの `/data` ボリュームと正しくマウント共有されている必要があります。

