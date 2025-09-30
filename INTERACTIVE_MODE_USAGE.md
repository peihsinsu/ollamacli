# Interactive Mode 使用說明

## 功能特性

### ✅ Readline 支援
- **上下鍵**瀏覽歷史指令
- 自動保存指令歷史到記憶體
- Ctrl+C 優雅退出

### ✅ 彩色輸出
- **青色 (Cyan)**: 標題和系統訊息
- **黃色 (Yellow)**: 指令名稱和標籤
- **綠色 (Green)**: 提示符 `>` 和列表符號
- **紅色 (Red)**: 錯誤訊息

### ✅ 多種退出方式
- `/exit` - 標準指令
- `exit` - 簡化指令（不需斜線）
- `quit` - 簡化指令
- `Ctrl+C` - 信號處理

## 使用方法

### 啟動 Interactive Mode

```bash
# 基本用法
ollamacli chat llama3.2:1b --interactive

# 或簡寫
ollamacli chat llama3.2:1b -i
```

### 互動範例

```
Interactive mode with llama3.2:1b (type /help for commands, Ctrl+C to exit)

> 你好
你好！很高興見到你。有什麼我可以幫助你的嗎？

> 2+2等於多少？
2+2等於4。

> /status
Session Status:
  Model: llama3.2:1b
  Total messages: 4
  User messages: 2
  Assistant messages: 2

> /model list
Available models:
  • llama3.2:1b
  • llama2
  • codellama

> exit
Goodbye! Session ended.
```

### 可用指令

| 指令 | 說明 |
|------|------|
| `/help` | 顯示幫助訊息 |
| `/clear` | 清除對話歷史 |
| `/model list` | 列出可用模型 |
| `/model pull <name>` | 拉取模型 |
| `/model show <name>` | 顯示模型資訊 |
| `/status` | 顯示 session 狀態 |
| `/save [file]` | 保存對話歷史 |
| `/load [file]` | 載入對話歷史 |
| `/exit` | 退出 |
| `exit` / `quit` | 退出（簡化版） |

### 歷史記錄導航

- **Up Arrow (↑)**: 前一個指令
- **Down Arrow (↓)**: 下一個指令
- **Ctrl+R**: 搜尋歷史（標準 readline 功能）

## 注意事項

### TTY 需求

Interactive mode 需要在**真正的終端（TTY）**環境中運行：

✅ **支援環境：**
- macOS Terminal.app
- iTerm2
- Linux 終端（gnome-terminal, konsole, xterm 等）
- Windows Terminal / PowerShell

❌ **不支援環境：**
- CI/CD 管線
- 腳本執行（非互動式）
- 某些 IDE 內建終端（視 IDE 而定）

### 測試方法

在真正的終端中執行：

```bash
# 確保在實際終端中運行
cd /Users/cx009/project/claudews/ollamacli
./build/ollamacli chat llama3.2:1b --interactive
```

### 顏色支援

如果終端不支援 ANSI 顏色，將顯示純文字但功能正常。

大多數現代終端都支援 256 色：
- macOS Terminal.app ✅
- iTerm2 ✅
- Windows Terminal ✅
- VSCode 終端 ✅

## 疑難排解

### 問題：顯示 "invalid prompt"

**原因：** 不是在真正的 TTY 環境中運行

**解決：** 在實際的終端應用程式中執行，而不是透過管線或腳本

### 問題：顏色顯示為亂碼

**原因：** 終端不支援 ANSI 顏色碼

**解決：** 更新終端或使用支援顏色的終端模擬器

### 問題：上下鍵不工作

**原因：** readline 功能未正確初始化

**檢查：**
```bash
# 確認 liner 依賴已安裝
go mod verify

# 重新編譯
make build-dev
```

## 開發測試

在真正的終端中測試：

```bash
# 1. 編譯
make build-dev

# 2. 在真正的 Terminal.app 或 iTerm2 中運行
./build/ollamacli chat llama3.2:1b --interactive

# 3. 測試功能
> /help           # 測試幫助
> /model list     # 測試模型列表
> /status         # 測試狀態
> 你好            # 測試對話
> (按上鍵)       # 測試歷史
> exit            # 測試退出
```

## 顏色代碼參考

如果需要自定義顏色：

```go
// ANSI 顏色碼
\033[1;31m  // 紅色粗體
\033[1;32m  // 綠色粗體
\033[1;33m  // 黃色粗體
\033[1;36m  // 青色粗體
\033[0m     // 重置
```