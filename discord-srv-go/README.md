# 📣 descord-srv-go

Minecraft サーバのログを監視し、Discord Webhook にログイン・ログアウトの通知を送る Go アプリケーションです。

## ✅ 構成

```
.
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

### 1. 必要な Go ツールのインストール

```bash
bash scripts/install_go_tools.sh
```

### 2. Docker コンテナの起動

```bash
docker-compose up -d
```

### 3. 環境変数の設定

`.env` または `docker-compose.yml` に以下のように指定してください：

```env
MINECRAFT_LOG_PATH=/data/logs/latest.log
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxxxx/yyyyy
```

`/data/logs/latest.log` は、Minecraft サーバがログを出力するパスです。

## 🔎 開発・テスト

### 開発用ビルドと起動

```bash
docker build -f build/Dockerfile.dev -t descord-srv-go-dev .
docker run --rm -it \
  -v "$(pwd)":/app \
  -v ${MINECRAFT_DATA_PATH}:/data:ro \
  -e MINECRAFT_LOG_PATH=/data/logs/latest.log \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxxxx/yyyyy \
  descord-srv-go-dev
```

### ユニットテストの実行

```bash
go test ./pkg/cmd/...
```

※ `DISCORD_WEBHOOK_URL` を指定すると、統合テストで実際に通知が送信されます。

## 🔧 テスト内容

* `TestProcessLogLine`: ログ行の解析と通知内容の検証
* `TestRunWithNotifier_FileNotFound`: ログファイルが存在しない場合のリトライ検証
* `TestDiscordNotification_Integration`: Discord 通知の統合テスト（環境変数必須）

## 📌 注意点

* `MINECRAFT_LOG_PATH` に指定されたファイルが存在しない場合、10回（デフォルト）までリトライします。
* コンテナ内で `/data/logs/latest.log` を参照するため、Minecraft サーバの `/data` ボリュームと正しくマウント共有されている必要があります。

