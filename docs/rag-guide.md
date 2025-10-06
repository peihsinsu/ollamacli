# RAG (Retrieval Augmented Generation) 使用指南

## 概述

ollamacli 的 RAG 功能讓你能夠基於特定文件集合進行問答，而不是僅依賴模型的一般知識。這對於以下場景特別有用：

- 根據內部文檔回答問題
- 基於代碼庫進行技術問答
- 使用特定領域知識進行推理

## 工作原理

RAG 分為兩個階段：

### 1. 索引階段（Indexing）

將文檔分塊、轉換為向量嵌入，並存儲在本地知識庫中。

```bash
# 索引單個文件
ollamacli embed-files --files doc1.md --files doc2.md

# 索引整個目錄
ollamacli embed-files --dir ./docs --patterns "*.md" --patterns "*.txt"

# 索引代碼庫
ollamacli embed-files --dir ./src --patterns "*.go" --patterns "*.py"
```

### 2. 檢索階段（Retrieval）

查詢時自動檢索相關文檔片段，並將其作為上下文提供給模型。

```bash
# 基於知識庫進行問答
ollamacli rag-chat --prompt "如何設置配置文件？"

# 使用不同的模型
ollamacli rag-chat llama3.2 --prompt "代碼中的錯誤處理是如何實現的？"

# 調整檢索的文檔數量
ollamacli rag-chat --prompt "什麼是 RAG？" --top-k 5
```

## 完整使用流程

### 步驟 1: 準備文檔

確保你的文檔是文本格式（`.txt`, `.md`, `.go`, `.py` 等）。

```bash
mkdir my-docs
echo "ollamacli 是一個 Ollama 的命令行工具。" > my-docs/intro.md
echo "RAG 通過檢索相關文檔增強生成質量。" > my-docs/rag-info.md
```

### 步驟 2: 索引文檔

```bash
# 方法 1: 指定單個文件
ollamacli embed-files \
  --files my-docs/intro.md \
  --files my-docs/rag-info.md \
  --model mxbai-embed-large

# 方法 2: 索引整個目錄
ollamacli embed-files \
  --dir my-docs \
  --patterns "*.md" \
  --model mxbai-embed-large
```

### 步驟 3: 基於知識庫問答

```bash
# 直接問問題
ollamacli rag-chat --prompt "ollamacli 是什麼？"

# 使用管道
echo "RAG 的工作原理是什麼？" | ollamacli rag-chat
```

## 配置選項

### 全局配置 (`~/.ollamacli/config.yaml`)

```yaml
rag:
  knowledge_base: ~/.ollamacli/knowledge.db  # 知識庫路徑
  embed_model: mxbai-embed-large              # 嵌入模型
  chunk_size: 500                             # 文本塊大小
  chunk_overlap: 50                           # 塊重疊大小
  allowed_files:                              # 允許的文件模式
    - "*.md"
    - "*.txt"
    - "*.go"
```

### 命令行參數

#### `embed-files` 參數

```bash
ollamacli embed-files [FLAGS]

Flags:
  -f, --files strings        要嵌入的文件列表
  -d, --dir string           要掃描的目錄
      --patterns strings     文件匹配模式 (default: *.txt,*.md,*.go)
  -m, --model string         嵌入模型 (default: 配置文件中的設定)
      --db string            知識庫路徑 (default: ~/.ollamacli/knowledge.db)
      --chunk-size int       文本塊大小 (default: 500)
      --chunk-overlap int    塊重疊大小 (default: 50)
```

#### `rag-chat` 參數

```bash
ollamacli rag-chat [MODEL] [FLAGS]

Args:
  MODEL                      聊天模型 (default: llama3.1:8b)

Flags:
      --prompt string        問題內容 (必需)
      --embed-model string   嵌入模型 (default: 配置文件中的設定)
      --db string            知識庫路徑 (default: ~/.ollamacli/knowledge.db)
  -k, --top-k int            檢索的文檔數量 (default: 3)
  -f, --format string        輸出格式 (text, json) (default: text)
```

## 進階用法

### 使用不同的嵌入模型

```bash
# 使用 nomic-embed-text 模型
ollamacli embed-files \
  --dir ./docs \
  --model nomic-embed-text

# 確保 rag-chat 使用相同的模型
ollamacli rag-chat \
  --prompt "你的問題" \
  --embed-model nomic-embed-text
```

### 管理多個知識庫

```bash
# 創建特定主題的知識庫
ollamacli embed-files \
  --dir ./tech-docs \
  --db ~/.ollamacli/tech-kb.db

ollamacli embed-files \
  --dir ./business-docs \
  --db ~/.ollamacli/business-kb.db

# 查詢特定知識庫
ollamacli rag-chat \
  --prompt "技術問題" \
  --db ~/.ollamacli/tech-kb.db

ollamacli rag-chat \
  --prompt "業務問題" \
  --db ~/.ollamacli/business-kb.db
```

### 調整塊大小以優化性能

```bash
# 較小的塊（更精確但可能失去上下文）
ollamacli embed-files \
  --dir ./docs \
  --chunk-size 300 \
  --chunk-overlap 30

# 較大的塊（更多上下文但可能不夠精確）
ollamacli embed-files \
  --dir ./docs \
  --chunk-size 800 \
  --chunk-overlap 80
```

### 過濾文件類型

```bash
# 只索引 Go 代碼
ollamacli embed-files \
  --dir ./src \
  --patterns "*.go"

# 索引多種文件類型
ollamacli embed-files \
  --dir ./project \
  --patterns "*.go" \
  --patterns "*.md" \
  --patterns "*.yaml"
```

## 最佳實踐

### 1. 文檔準備

- **保持文檔結構清晰**：使用標題和段落
- **避免過多格式**：純文本效果最好
- **定期更新**：文檔變更後重新索引

### 2. 塊大小選擇

- **小文檔（< 1000 字）**：chunk_size: 200-300
- **中等文檔（1000-5000 字）**：chunk_size: 500-600
- **大文檔（> 5000 字）**：chunk_size: 800-1000

### 3. 查詢優化

- **明確的問題**：越具體的問題，檢索效果越好
- **調整 top-k**：如果答案不完整，增加 `--top-k` 值
- **使用正確的嵌入模型**：確保查詢和索引使用同一個模型

### 4. 性能考慮

- **批量索引**：使用 `--dir` 而不是多次 `--files`
- **知識庫大小**：單個知識庫建議不超過 10,000 個文檔
- **定期清理**：刪除過時的知識庫文件

## 故障排除

### 問題：檢索不到相關文檔

**解決方案**：
1. 確認文檔已正確索引：`ls -lh ~/.ollamacli/knowledge.db`
2. 嘗試增加 `--top-k` 值
3. 檢查嵌入模型是否一致

### 問題：響應質量不佳

**解決方案**：
1. 調整塊大小：較大的塊提供更多上下文
2. 增加 overlap：確保重要信息不會被分割
3. 使用更好的聊天模型（如 llama3.2 而不是 llama2）

### 問題：索引速度慢

**解決方案**：
1. 減少文件數量或使用更具體的 `--patterns`
2. 使用更快的嵌入模型
3. 考慮分批處理大型文檔集合

## 示例場景

### 場景 1: 技術文檔問答

```bash
# 索引技術文檔
ollamacli embed-files --dir ./docs/api --patterns "*.md"

# 查詢 API 用法
ollamacli rag-chat --prompt "如何使用 REST API 創建用戶？"
```

### 場景 2: 代碼庫分析

```bash
# 索引 Go 代碼
ollamacli embed-files --dir ./internal --patterns "*.go"

# 詢問代碼實現
ollamacli rag-chat --prompt "認證流程是如何實現的？"
```

### 場景 3: 多語言文檔

```bash
# 索引多種類型的文件
ollamacli embed-files \
  --dir ./project \
  --patterns "*.md" \
  --patterns "*.go" \
  --patterns "*.yaml" \
  --patterns "*.sh"

# 綜合性問題
ollamacli rag-chat --prompt "這個項目的部署流程是什麼？" --top-k 5
```

## 限制與注意事項

1. **嵌入模型依賴**: 必須確保 Ollama 服務器上已安裝嵌入模型（如 `mxbai-embed-large`）
2. **向量搜索精度**: 使用餘弦相似度，可能不如專用向量數據庫精確
3. **本地存儲**: 所有數據存儲在本地 SQLite 數據庫中
4. **語言支持**: 嵌入模型的語言支持取決於具體模型

## 相關資源

- [Ollama 嵌入模型列表](https://ollama.com/library)
- [向量相似度搜索原理](https://en.wikipedia.org/wiki/Cosine_similarity)
- [RAG 架構設計](../docs/decisions/rag-architecture.md)
