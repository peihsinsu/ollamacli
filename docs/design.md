# ollamacli 設計文件

## 背景
- 使用者希望透過指令列直接與本機或遠端的 Ollama server 溝通，以簡化模型管理與對話流程。
- 目前缺乏統一的 CLI 工具，導致需要手動呼叫 REST API 或透過其他 UI 工具，效率較低。
- 工具需兼容既有的 Ollama API 規格，並對未來新增的模型操作保持彈性。

## 目標
- 提供簡潔一致的 CLI 介面，涵蓋常見的 Ollama 操作（如模型列表、模型拉取、對話、推論結果串流）。
- 支援互動式與一次性呼叫兩種模式，以滿足腳本化與人工探索需求。
- 透過設定檔、環境變數與旗標讓使用者可定義伺服器端點、預設模型與輸出格式。
- 提供良好錯誤訊息與非零結束碼，讓工具易於整合至自動化流程。

## 非目標
- 不計畫取代 Ollama server 自身的模型管理能力，只做客戶端交互。
- 不負責模型部署、版本控管或資源監控。
- 不提供 GUI 或瀏覽器介面。

## 典型使用情境
1. 顯示目前伺服器可用模型：`ollamacli list`
2. 從遠端 Registry 拉取模型：`ollamacli pull llama2`
3. 使用指定模型產生回應：`ollamacli chat llama2 --prompt "Hello"`
4. 透過管線串流提示：`cat prompt.txt | ollamacli chat llama2`
5. 將 CLI 一次性呼叫嵌入腳本並讀取 JSON 結果：`ollamacli run llama2 --prompt "data" --format json`
6. 進入互動模式持續對話：`ollamacli chat llama2 --interactive` (使用 Ctrl+C 離開)

## 功能需求
- **指令解析**：支援子指令（如 `list`、`pull`、`show`、`chat`、`run`、`serve`）。
- **伺服器連線設定**：可透過旗標 `--host`、`--port`、環境變數 `OLLAMA_HOST`，或設定檔自動載入。
- **認證**：若伺服器設定 Token，CLI 需支援 `--token` 與環境變數 `OLLAMA_TOKEN`。
- **輸入模式**：接受標準輸入（stdin）、`--prompt` 旗標、或互動式結構化對話模式。
- **互動模式**：支援 `--interactive` 旗標進入持續對話 session，保持上下文歷史，使用 Ctrl+C 優雅離開。
- **輸出模式**：預設為純文字；可選擇 `--format json` 取得原始事件或結構化資料；支援逐行/串流輸出。
- **錯誤/重試**：遇到網路錯誤可選擇重試（`--retry` 次數、`--retry-delay`）。
- **日誌**：`--verbose` 顯示 HTTP 要求/回應摘要，`--quiet` 只輸出必要資訊。

## 系統架構
CLI 由三層組成：
1. **Interface Layer**：使用 `cobra` 或等效 CLI parsing library 處理指令與旗標。
2. **Service Layer**：封裝 Ollama REST API 呼叫，提供模型管理、對話、執行等函式。
3. **IO Layer**：處理輸入/輸出與串流，維持互動模式狀態，將 API 回應格式化後輸出。

```
使用者 CLI -> 指令解析 -> Service Client -> HTTP(S) -> Ollama Server
            ↘ 設定管理 ↙            ↘ 輸出格式化 ↙
```

## 主要模組與責任
- `cmd/`：各子指令定義與旗標處理，負責呼叫對應的 service 函式。
- `internal/config`：讀取設定檔（預設 `~/.ollamacli/config.yaml`）、環境變數與旗標合併。
- `internal/client`：包裝 HTTP Client，管理 Base URL、認證與重試策略。
- `internal/chat`：處理互動式對話 loop，包括 Prompt 歷史紀錄與串流輸出，實作 signal handling (Ctrl+C)。
- `internal/output`：根據 `--format` 與 `--quiet` 等旗標格式化輸出（純文字、JSON、event stream）。
- `internal/log`：統一的 logging 介面，支援 debug、info、error 等層級。

## 資料流程
1. 啟動時讀取設定與旗標，建立 `AppContext`（包含設定、logger、client）。
2. 解析子指令，對應呼叫 service 函式。
3. Service 函式向 `internal/client` 發出 HTTP 請求。
4. 回應資料送交 `internal/output` 轉換，若為串流則逐事件處理。
5. CLI 將結果寫出至 stdout 或檔案。

## Interactive Mode 設計

### 核心功能
- **持續對話**：使用 `--interactive` 旗標進入互動模式，保持與模型的持續對話 session。
- **上下文維護**：自動管理對話歷史，維持上下文連貫性。
- **優雅退出**：監聽 Ctrl+C (SIGINT) 信號，清理資源後優雅離開。

### 使用者體驗
```bash
$ ollamacli chat llama2 --interactive
Connected to llama2 model. Type your message or press Ctrl+C to exit.

> Hello, how are you?
I'm doing well, thank you for asking! How can I help you today?

> What's the weather like?
I don't have access to real-time weather data. You might want to check a weather app or website for current conditions in your area.

> ^C
Goodbye! Session ended.
```

### 技術實作
- **Readline 支援**：提供指令歷史、自動完成等功能。
- **Signal Handling**：使用 `os/signal` 套件監聽 SIGINT，實作優雅關閉。
- **Session State**：維護對話狀態，包括訊息歷史和模型設定。
- **錯誤復原**：網路中斷時提供重連選項，保持 session 狀態。

### 進階功能
- **多行輸入**：支援 `\` 續行符號或特殊指令模式。
- **Session 指令**：內建 `/help`、`/clear`、`/save`、`/load` 等 meta 指令。
- **輸出控制**：在互動模式中支援即時調整輸出格式。

## 設定與擴充性
- 支援 YAML 設定檔，包含預設伺服器、預設模型、輸出格式、重試策略。
- 使用 Viper（或等同 library）整合旗標、環境變數與設定檔。
- 透過 Plug-in Hook 預留空間：例如 `plugins` 目錄可放置自訂子指令，在 runtime 掃描載入。

## 錯誤處理與回傳碼
- 一般錯誤（網路、解析、API HTTP 4xx/5xx）以非零碼結束，並在 stderr 顯示對應訊息。
- 常見錯誤碼：
  - `1`：一般錯誤
  - `2`：使用者輸入錯誤（缺少必填參數）
  - `3`：無法連線至 Ollama server
  - `4`：認證失敗
- 串流模式若中斷，需回傳錯誤碼並提示是否自動重試。

## 安全性考量
- Sensitive Token 儲存在設定檔時應可選擇加密或以系統憑證庫存取。
- CLI 禁止在 verbose 模式下直接輸出完整 Token，可用遮罩 (`****`).
- 支援 HTTPS 連線與自訂 CA 憑證匯入。

## 測試策略
- Unit tests：覆蓋 config 合併、指令解析、HTTP 客戶端錯誤處理、輸出格式化。
- Integration tests：使用 mock server（如 httptest）模擬 Ollama API 回應。
- End-to-end smoke：對本機開啟的 Ollama server 執行核心指令（list/pull/chat）。
- Linting 與 CI：採用 GitHub Actions 執行 gofmt、golangci-lint、單元測試。

## 發佈與部署
- 透過 GoReleaser 產生各平台二進位檔（macOS、Linux、Windows）。
- 提供 Homebrew tap 與直接下載連結。
- 版本遵循 Semantic Versioning，透過 Git tag 觸發 CI 發佈。

## 未來工作項目
- 進階 `session` 管理：儲存對話上下文於檔案，支援 session 恢復。
- 支援 WebSocket 介面以取得更即時的串流事件。
- 提供 `--prompt-template` 功能，支援自訂樣板。
- 整合 API 使用統計與簡易分析。
