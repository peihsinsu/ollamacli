# ADR-0001: 專案初始化與技術選型

## Status
Accepted

## Context
根據設計文件 docs/design.md 初始化 ollamacli 專案，需要確定基礎架構和技術選型。

### 專案需求
- 專案類型: CLI Application
- 技術棧: Golang
- 目標: 提供統一的 Ollama CLI 介面

## Decision
採用以下技術架構:

### 後端技術
- **語言**: Go 1.21+
- **CLI 框架**: Cobra (指令解析與子指令管理)
- **配置管理**: Viper (整合旗標、環境變數、設定檔)
- **HTTP 客戶端**: 標準庫 net/http + 自訂封裝
- **日誌**: Logrus 或 標準庫 slog
- **測試**: Go 內建測試 + httptest

### 架構設計
- **分層結構**: Interface Layer (cmd) → Service Layer (internal) → IO Layer (output)
- **模組職責明確**:
  - `cmd/`: 子指令定義 (list, pull, chat, run, etc.)
  - `internal/config`: 設定檔與環境變數管理
  - `internal/client`: Ollama API 客戶端封裝
  - `internal/chat`: 互動式對話處理 + Interactive Mode
  - `internal/output`: 多格式輸出 (text, json, stream)

### Interactive Mode 技術選型
- **Readline Library**: 使用 `github.com/chzyer/readline` 提供互動體驗
- **Signal Handling**: Go 標準庫 `os/signal` 處理 Ctrl+C
- **Session Management**: 記憶體內對話歷史 + 可選檔案持久化
- **Context Preservation**: 維護 Ollama API 所需的對話上下文

### 部署與發佈
- **建置工具**: GoReleaser (跨平台二進位檔)
- **CI/CD**: GitHub Actions
- **測試策略**: Unit + Integration + E2E
- **包管理**: Homebrew tap + 直接下載

### 設定與擴充
- **設定檔**: YAML 格式 (~/.ollamacli/config.yaml)
- **優先順序**: CLI flags > 環境變數 > 設定檔 > 預設值
- **擴充機制**: Plugin hooks (未來考慮)

## Consequences

### 優點
- Go 語言適合 CLI 工具開發，二進位檔案小且無依賴
- Cobra 是成熟的 CLI 框架，生態系統完整
- 架構清晰，職責分離，易於維護和測試
- 支援跨平台部署，使用便利

### 缺點
- 需要學習 Cobra 和 Viper 的使用方式
- 初期設定較複雜，需建立完整的專案結構
- 串流處理實作需要額外考慮

### 風險緩解
- 參考現有成熟 CLI 工具的設計模式
- 實施充分的測試策略 (unit, integration, e2e)
- 建立完整的錯誤處理機制
- 提供清晰的文件和使用範例

## Implementation Plan

### Phase 1: 基礎架構
- [x] 專案目錄結構建立
- [ ] Go module 初始化
- [ ] 基礎 Makefile 建立
- [ ] CI/CD pipeline 設定

### Phase 2: 核心功能
- [ ] Config 管理實作
- [ ] HTTP Client 封裝
- [ ] 基礎子指令 (list, show)
- [ ] 錯誤處理與日誌

### Phase 3: 進階功能
- [ ] 基礎 chat 指令 (一次性對話)
- [ ] 串流輸出處理
- [ ] 多格式輸出支援
- [ ] 重試機制

### Phase 3.5: Interactive Mode
- [ ] Readline 整合與基礎互動模式
- [ ] Signal handling (Ctrl+C 優雅離開)
- [ ] 對話歷史管理與上下文維護
- [ ] Session meta 指令 (/help, /clear, /save)
- [ ] 錯誤復原與重連機制

### Phase 4: 發佈準備
- [ ] 完整測試涵蓋
- [ ] 文件撰寫
- [ ] GoReleaser 設定
- [ ] 發佈流程驗證

## References
- [設計文件](../design.md)
- [專案配置](../../CLAUDE.md)
- [Cobra Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)

---
*建立時間: 2025-09-29 14:50:00*