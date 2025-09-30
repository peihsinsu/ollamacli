# ollamacli

一個功能強大的 Ollama CLI 客戶端，提供簡潔直觀的命令列介面來管理和使用 Ollama 模型。

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 功能特性

- 🚀 **簡潔的 CLI 介面** - 符合 Unix 工具慣例的命令設計
- 💬 **互動式對話模式** - 支援持續對話並維持上下文歷史
- 📦 **模型管理** - 列出、拉取、查看模型資訊
- 🔄 **串流支援** - 即時顯示模型回應
- 🧬 **向量嵌入** - 產生文件向量嵌入用於 RAG 應用
- 🛠️ **自定義模型** - 從 Modelfile 建立和量化自定義模型
- ⚙️ **靈活配置** - 支援環境變數、配置檔和命令列參數
- 🔌 **本地/遠端** - 支援連接本地或遠端 Ollama 伺服器
- 🔐 **安全認證** - 支援 Token 認證機制
- 📊 **多種輸出格式** - 支援純文字和 JSON 格式
- 🛠️ **腳本友善** - 完善的錯誤碼和安靜模式設計

## 目錄

- [安裝](#安裝)
  - [前置需求](#前置需求)
  - [從原始碼編譯](#從原始碼編譯)
  - [預編譯二進位檔](#預編譯二進位檔)
- [快速開始](#快速開始)
- [使用說明](#使用說明)
  - [基本指令](#基本指令)
  - [互動式對話](#互動式對話)
  - [配置管理](#配置管理)
- [進階用法](#進階用法)
- [開發指南](#開發指南)
- [疑難排解](#疑難排解)
- [授權](#授權)

## 安裝

### 前置需求

在使用 ollamacli 之前，請確保：

1. **Ollama 伺服器已安裝並運行**
   ```bash
   # 檢查 Ollama 是否運行
   curl http://localhost:11434/api/tags
   ```

2. **已安裝至少一個模型**
   ```bash
   # 使用 Ollama 安裝模型（如果尚未安裝）
   ollama pull llama2
   ```

3. **Go 1.21 或更新版本**（僅限從原始碼編譯）
   ```bash
   go version
   ```

### 從原始碼編譯

#### 1. 克隆專案

```bash
git clone https://github.com/yourusername/ollamacli.git
cd ollamacli
```

#### 2. 安裝依賴

```bash
make deps
```

或手動執行：

```bash
go mod download
go mod verify
```

#### 3. 編譯

**快速開發版本：**
```bash
make build-dev
```

**生產版本（包含優化和版本資訊）：**
```bash
make build
```

編譯完成的二進位檔將位於 `build/ollamacli`

#### 4. 安裝到系統路徑（可選）

```bash
make install
```

這會將 `ollamacli` 複製到 `$GOBIN` 目錄（通常是 `$GOPATH/bin`），使您可以在任何位置執行該命令。

**驗證安裝：**
```bash
ollamacli --help
```

#### 5. 解除安裝（如需要）

```bash
make uninstall
```

### 預編譯二進位檔

#### 跨平台編譯

您可以為多個平台編譯二進位檔：

```bash
make build-all
```

這將在 `dist/` 目錄下產生以下平台的二進位檔：
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

#### 手動下載（未來支援）

預編譯的二進位檔將在 GitHub Releases 頁面提供：

```bash
# macOS (arm64)
wget https://github.com/yourusername/ollamacli/releases/download/v0.1.0/ollamacli-darwin-arm64
chmod +x ollamacli-darwin-arm64
sudo mv ollamacli-darwin-arm64 /usr/local/bin/ollamacli

# Linux (amd64)
wget https://github.com/yourusername/ollamacli/releases/download/v0.1.0/ollamacli-linux-amd64
chmod +x ollamacli-linux-amd64
sudo mv ollamacli-linux-amd64 /usr/local/bin/ollamacli
```

## 快速開始

### 1. 列出可用模型

```bash
ollamacli list
```

### 2. 與模型對話（單次）

```bash
ollamacli chat llama2 --prompt "Hello, how are you?"
```

### 3. 啟動互動式對話

```bash
ollamacli chat llama2 --interactive
```

輸入您的訊息，按 `Ctrl+C` 或輸入 `/exit` 退出。

### 4. 從標準輸入讀取

```bash
echo "What is the capital of France?" | ollamacli chat llama2
```

### 5. 查看模型資訊

```bash
ollamacli show llama2
```

## 使用說明

### 基本指令

#### list - 列出可用模型

列出 Ollama 伺服器上所有可用的模型。

```bash
# 基本用法
ollamacli list

# JSON 格式輸出
ollamacli list --format json

# 安靜模式（僅顯示模型名稱）
ollamacli list --quiet
```

#### pull - 拉取模型

從遠端註冊表拉取模型到本地。

```bash
# 基本用法
ollamacli pull llama2

# 設定重試次數和延遲
ollamacli pull llama2 --retry 5 --retry-delay 10

# 允許不安全連接
ollamacli pull llama2 --insecure
```

#### show - 顯示模型詳細資訊

顯示指定模型的詳細資訊。

```bash
# 基本用法
ollamacli show llama2

# JSON 格式輸出
ollamacli show llama2 --format json

# 安靜模式（最少輸出）
ollamacli show llama2 --quiet
```

#### chat - 對話模式

與模型進行單次或互動式對話。

```bash
# 單次對話
ollamacli chat llama2 --prompt "Explain quantum computing"

# 從檔案讀取提示
ollamacli chat llama2 --prompt "$(cat prompt.txt)"

# 從標準輸入讀取
echo "Tell me a joke" | ollamacli chat llama2

# JSON 格式輸出
ollamacli chat llama2 --prompt "What is AI?" --format json
```

#### run - 執行單次生成

執行單次文字生成（與 chat 類似，但語義上更適合非對話場景）。

```bash
# 基本用法
ollamacli run llama2 --prompt "Write a haiku about programming"

# 管線處理
cat document.txt | ollamacli run llama2 --prompt "Summarize this:"
```

#### embed - 產生向量嵌入

使用嵌入模型產生文本的向量表示，用於語義搜索和 RAG（檢索增強生成）應用。

```bash
# 從命令列參數產生嵌入
ollamacli embed nomic-embed-text --input "Hello, world!"

# 從檔案產生嵌入
ollamacli embed nomic-embed-text --file document.txt

# 從標準輸入產生嵌入
echo "Sample text" | ollamacli embed nomic-embed-text

# 批次處理多個文本
ollamacli embed nomic-embed-text --input "First text" --input "Second text"

# JSON 格式輸出（用於保存向量）
ollamacli embed nomic-embed-text --input "Text" --format json > embeddings.json
```

**推薦的嵌入模型：**
- `nomic-embed-text` - 高質量通用嵌入模型
- `mxbai-embed-large` - 大型上下文長度模型
- `all-minilm` - 輕量級快速模型

#### build - 建立自定義模型

從 Modelfile 建立自定義模型，支援參數調整、系統提示詞和模型量化。

```bash
# 產生 Modelfile 範本
ollamacli build mymodel --sample

# 從 Modelfile 建立模型
ollamacli build mymodel --modelfile Modelfile

# 建立並量化模型
ollamacli build mymodel --modelfile Modelfile --quantize q4_K_M

# 從現有模型路徑建立
ollamacli build mymodel --path /path/to/model/weights
```

**Modelfile 範例：**

```dockerfile
# 基於現有模型
FROM llama3.2

# 設定參數
PARAMETER temperature 0.8
PARAMETER top_p 0.9
PARAMETER num_ctx 4096

# 設定系統提示詞
SYSTEM """
You are a helpful coding assistant specialized in Python.
"""
```

**量化等級選項：**
- `q4_0` - 4-bit 量化（最小）
- `q4_K_M` - 4-bit 中等質量（推薦）
- `q5_K_M` - 5-bit 中等質量
- `q8_0` - 8-bit 量化（高質量）

### 互動式對話

互動式模式允許您與模型進行持續對話，自動維持上下文歷史。

#### 啟動互動模式

```bash
ollamacli chat llama2 --interactive
```

#### 互動模式內建指令

在互動模式中，您可以使用以下特殊指令：

| 指令 | 說明 |
|------|------|
| `/help` | 顯示可用指令 |
| `/clear` | 清除對話歷史 |
| `/save [filename]` | 儲存對話歷史（計劃中） |
| `/load [filename]` | 載入對話歷史（計劃中） |
| `/exit` | 退出互動模式 |
| `Ctrl+C` | 優雅退出 |

#### 互動模式範例

```
$ ollamacli chat llama2 --interactive
Connected to llama2 model. Type your message or press Ctrl+C to exit.

> Hello! Can you help me with Python?
Of course! I'd be happy to help you with Python. What would you like to know?

> How do I read a file?
You can read a file in Python using the open() function. Here's an example:

with open('filename.txt', 'r') as file:
    content = file.read()
    print(content)

> Thanks!
You're welcome! Feel free to ask if you have more questions.

> /exit
Goodbye! Session ended.
```

### 配置管理

ollamacli 支援三種配置方式，優先順序由高到低：

1. **命令列旗標**（最高優先級）
2. **環境變數**
3. **配置檔**（最低優先級）

#### 1. 命令列旗標

所有指令都支援以下通用旗標：

```bash
ollamacli list \
  --host localhost \
  --port 11434 \
  --token your-auth-token \
  --verbose \
  --format json
```

**常用旗標：**

| 旗標 | 說明 | 預設值 |
|------|------|--------|
| `--host` | Ollama 伺服器主機 | localhost |
| `--port` | Ollama 伺服器埠號 | 11434 |
| `--token` | 認證 Token | - |
| `--format` | 輸出格式 (text/json) | text |
| `--verbose` | 詳細輸出（顯示 HTTP 請求） | false |
| `--quiet` | 安靜模式（最少輸出） | false |
| `--insecure` | 允許不安全的連接 | false |
| `--retry` | 重試次數 | 0 |
| `--retry-delay` | 重試延遲（秒） | 5 |

#### 2. 環境變數

```bash
# 設定伺服器位置
export OLLAMA_HOST=remote-server.example.com
export OLLAMA_PORT=11434

# 設定認證 Token
export OLLAMA_TOKEN=your-auth-token

# 設定日誌等級
export OLLAMA_LOG_LEVEL=debug
export OLLAMA_VERBOSE=true

# 使用環境變數配置
ollamacli list
```

**支援的環境變數：**

| 環境變數 | 說明 |
|----------|------|
| `OLLAMA_HOST` | 伺服器主機名稱或 IP |
| `OLLAMA_PORT` | 伺服器埠號 |
| `OLLAMA_TOKEN` | 認證 Token |
| `OLLAMA_LOG_LEVEL` | 日誌等級 (debug/info/warn/error) |
| `OLLAMA_VERBOSE` | 是否啟用詳細輸出 |

#### 3. 配置檔

在 `~/.ollamacli/config.yaml` 建立配置檔：

```yaml
# Ollama 伺服器配置
host: localhost
port: 11434
token: your-auth-token

# 日誌配置
log_level: info
verbose: false
quiet: false

# 預設設定
default_model: llama2
default_format: text

# 重試配置
retry: 3
retry_delay: 5

# 安全性
insecure: false
```

**配置檔位置：**
- Linux/macOS: `~/.ollamacli/config.yaml`
- Windows: `%USERPROFILE%\.ollamacli\config.yaml`

## 進階用法

### 腳本整合

ollamacli 設計為易於在 shell 腳本中使用。

#### 取得模型列表並迭代

```bash
#!/bin/bash

# 取得所有模型名稱
MODELS=$(ollamacli list --quiet)

# 迭代處理每個模型
for model in $MODELS; do
    echo "Testing model: $model"
    ollamacli chat $model --prompt "Say hello" --quiet
done
```

#### 批次處理提示

```bash
# 處理多個提示檔案
cat prompts.txt | while read prompt; do
    echo "Prompt: $prompt"
    ollamacli run llama2 --prompt "$prompt" >> responses.txt
done
```

#### 錯誤處理

```bash
#!/bin/bash

# 執行命令並檢查結果
ollamacli list --host remote-server

if [ $? -eq 0 ]; then
    echo "✓ Successfully connected"
else
    case $? in
        1) echo "✗ General error" ;;
        2) echo "✗ Invalid arguments" ;;
        3) echo "✗ Cannot connect to server" ;;
        4) echo "✗ Authentication failed" ;;
        *) echo "✗ Unknown error" ;;
    esac
    exit 1
fi
```

**錯誤碼說明：**

| 錯誤碼 | 說明 |
|--------|------|
| 0 | 成功 |
| 1 | 一般錯誤 |
| 2 | 使用者輸入錯誤 |
| 3 | 無法連線至伺服器 |
| 4 | 認證失敗 |

### JSON 輸出處理

使用 `jq` 處理 JSON 輸出：

```bash
# 提取模型名稱
ollamacli list --format json | jq -r '.models[].name'

# 取得模型大小
ollamacli list --format json | jq -r '.models[] | "\(.name): \(.size)"'

# 美化輸出模型資訊
ollamacli show llama2 --format json | jq .

# 提取特定欄位
ollamacli show llama2 --format json | jq -r '.modelfile'
```

### 遠端伺服器連接

```bash
# 連接遠端伺服器
ollamacli list --host remote.example.com --port 11434

# 使用 Token 認證
ollamacli list \
  --host remote.example.com \
  --token "your-secret-token"

# 設定環境變數後使用
export OLLAMA_HOST=remote.example.com
export OLLAMA_TOKEN=your-secret-token
ollamacli list
```

### 管線和重定向

```bash
# 從檔案讀取輸入
cat input.txt | ollamacli chat llama2

# 輸出到檔案
ollamacli chat llama2 --prompt "Explain AI" > output.txt

# 錯誤輸出重定向
ollamacli list --verbose 2> debug.log

# 組合使用
cat questions.txt | ollamacli chat llama2 --quiet > answers.txt 2> errors.log
```

## 開發指南

### 開發環境設定

```bash
# 克隆專案
git clone https://github.com/yourusername/ollamacli.git
cd ollamacli

# 安裝依賴
make deps

# 執行測試
make test

# 執行測試並產生覆蓋率報告
make test-coverage

# 格式化程式碼
make fmt

# 執行 linter
make lint

# 快速編譯和執行
make run ARGS="list"
```

### Makefile 目標

| 指令 | 說明 |
|------|------|
| `make build` | 編譯生產版本 |
| `make build-dev` | 快速編譯開發版本 |
| `make run ARGS="..."` | 編譯並執行 |
| `make run-direct ARGS="..."` | 直接執行（不編譯） |
| `make test` | 執行測試 |
| `make test-coverage` | 測試覆蓋率報告 |
| `make benchmark` | 執行基準測試 |
| `make fmt` | 格式化程式碼 |
| `make vet` | 執行 go vet |
| `make lint` | 執行 golangci-lint |
| `make clean` | 清理編譯產物 |
| `make install` | 安裝到系統 |
| `make uninstall` | 從系統移除 |
| `make build-all` | 跨平台編譯 |
| `make package` | 建立發布套件 |
| `make info` | 顯示專案資訊 |
| `make help` | 顯示所有可用指令 |

### 專案結構

```
ollamacli/
├── cmd/                    # CLI 入口點和子指令
│   └── ollamacli/
│       └── main.go
├── internal/               # 內部套件（不對外公開）
│   ├── config/            # 配置管理
│   ├── client/            # Ollama API 客戶端
│   ├── chat/              # 互動式對話處理
│   ├── output/            # 輸出格式化
│   └── log/               # 日誌管理
├── pkg/                    # 可重用的公開套件
├── docs/                   # 文件
│   ├── design.md          # 設計文件
│   └── decisions/         # 架構決策記錄（ADR）
├── examples/               # 使用範例
├── test/                   # 測試資料和配置
├── scripts/                # 自動化腳本
├── build/                  # 編譯輸出目錄
├── dist/                   # 發布套件目錄
├── Makefile               # 建置工具
├── go.mod                 # Go 模組定義
└── README.md              # 本文件
```

### 執行測試

```bash
# 執行所有測試
make test

# 執行特定套件的測試
go test ./internal/config -v

# 產生覆蓋率報告
make test-coverage
# 然後在瀏覽器開啟 build/coverage.html

# 執行基準測試
make benchmark

# 執行競態檢測
go test -race ./...
```

### 貢獻指南

1. Fork 本專案
2. 建立您的功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的變更 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

**請確保：**
- 遵循 Go 程式碼規範
- 撰寫單元測試（覆蓋率 > 80%）
- 更新相關文件
- 所有測試通過
- Linter 無警告

## 疑難排解

### 無法連接到 Ollama 伺服器

```bash
# 檢查 Ollama 是否運行
curl http://localhost:11434/api/tags

# 如果不通，啟動 Ollama 服務
ollama serve

# 嘗試使用不同的主機/埠
ollamacli list --host 127.0.0.1 --port 11434

# 啟用詳細輸出查看連接詳情
ollamacli list --verbose
```

### 模型不存在

```bash
# 列出所有可用模型
ollamacli list

# 拉取所需的模型
ollamacli pull llama2

# 檢查模型是否成功安裝
ollamacli show llama2
```

### 認證失敗

```bash
# 確認 Token 設定正確
ollamacli list --token your-correct-token

# 或透過環境變數
export OLLAMA_TOKEN=your-correct-token
ollamacli list
```

### 啟用除錯模式

```bash
# 透過環境變數
export OLLAMA_LOG_LEVEL=debug
export OLLAMA_VERBOSE=true
ollamacli chat llama2 --prompt "test"

# 或使用旗標
ollamacli chat llama2 --prompt "test" --verbose
```

### 效能問題

1. **使用 `--quiet` 模式** - 在腳本中使用以減少輸出開銷
2. **選擇適當的模型** - 較小的模型回應更快
3. **重用互動式 session** - 對同一模型的多次查詢使用互動模式
4. **調整重試設定** - 根據網路狀況調整 `--retry` 和 `--retry-delay`

### 常見問題

**Q: ollamacli 和 ollama 有什麼不同？**

A: `ollama` 是官方的 CLI 工具和伺服器，`ollamacli` 是一個專注於客戶端互動的替代工具，提供更豐富的腳本整合和互動功能。

**Q: 可以同時連接多個 Ollama 伺服器嗎？**

A: 是的，使用 `--host` 和 `--port` 旗標可以連接不同的伺服器。

**Q: 互動模式如何保存對話歷史？**

A: 目前對話歷史在 session 期間保存在記憶體中，持久化儲存功能正在開發中。

**Q: 支援哪些輸出格式？**

A: 目前支援純文字（text）和 JSON 兩種格式，可透過 `--format` 旗標指定。

## 相關資源

- [Ollama 官方網站](https://ollama.ai/)
- [Ollama GitHub](https://github.com/ollama/ollama)
- [設計文件](docs/design.md)
- [範例文件](examples/basic_usage.md)

## 授權

本專案採用 MIT 授權 - 詳見 [LICENSE](LICENSE) 檔案。

## 致謝

- 感謝 [Ollama](https://ollama.ai/) 提供強大的本地 LLM 運行平台
- 感謝 [Cobra](https://github.com/spf13/cobra) 提供優秀的 CLI 框架

---

如有問題或建議，歡迎提交 [Issue](https://github.com/yourusername/ollamacli/issues) 或 [Pull Request](https://github.com/yourusername/ollamacli/pulls)。