# RAG 演示範例

這個範例展示如何使用 ollamacli 的 RAG 功能來建立基於文檔的問答系統。

## 前置需求

1. 已安裝並運行 Ollama
2. 已下載嵌入模型：
   ```bash
   ollama pull mxbai-embed-large
   ```
3. 已下載聊天模型：
   ```bash
   ollama pull llama3.1:8b
   ```

## 範例文檔

本範例包含三個示範文檔：

1. **golang-basics.md** - Go 語言基礎知識
2. **web-development.md** - Web 開發指南
3. **testing.md** - 測試最佳實踐

## 使用步驟

### 1. 索引文檔

```bash
# 從專案根目錄執行
ollamacli embed-files \
  --dir examples/rag-demo/docs \
  --patterns "*.md" \
  --model mxbai-embed-large
```

### 2. 進行問答

```bash
# 問關於 Go 語言的問題
ollamacli rag-chat --prompt "Go 語言有什麼特點？"

# 問關於 Web 開發的問題
ollamacli rag-chat --prompt "如何設計 RESTful API？"

# 問關於測試的問題
ollamacli rag-chat --prompt "什麼是測試驅動開發？"
```

### 3. 調整檢索參數

```bash
# 增加檢索的文檔數量以獲得更全面的答案
ollamacli rag-chat --prompt "完整的開發流程包括哪些步驟？" --top-k 5

# 使用不同的聊天模型
ollamacli rag-chat llama3.2 --prompt "比較不同的測試方法" --top-k 3
```

## 範例腳本

提供了一個示範腳本 `demo.sh`，可以自動執行索引和問答流程：

```bash
cd examples/rag-demo
chmod +x demo.sh
./demo.sh
```

## 預期結果

運行 RAG 系統後，你應該能夠：

1. ✅ 根據提供的文檔回答具體問題
2. ✅ 獲得比通用模型更準確的答案
3. ✅ 看到系統引用相關文檔片段的內容

## 進階實驗

### 實驗 1: 比較有無 RAG 的差異

```bash
# 使用 RAG
ollamacli rag-chat --prompt "Go 的並發模型是什麼？"

# 不使用 RAG（普通聊天）
ollamacli chat llama3.1:8b --prompt "Go 的並發模型是什麼？"
```

觀察兩者的答案差異。

### 實驗 2: 調整塊大小

```bash
# 使用較小的塊
ollamacli embed-files \
  --dir examples/rag-demo/docs \
  --patterns "*.md" \
  --chunk-size 300 \
  --chunk-overlap 30 \
  --db ~/rag-small-chunks.db

ollamacli rag-chat --prompt "Go 的錯誤處理方式" --db ~/rag-small-chunks.db

# 使用較大的塊
ollamacli embed-files \
  --dir examples/rag-demo/docs \
  --patterns "*.md" \
  --chunk-size 800 \
  --chunk-overlap 80 \
  --db ~/rag-large-chunks.db

ollamacli rag-chat --prompt "Go 的錯誤處理方式" --db ~/rag-large-chunks.db
```

比較不同塊大小對檢索質量的影響。

### 實驗 3: 多個知識庫

```bash
# 創建技術知識庫
ollamacli embed-files --dir examples/rag-demo/docs --db ~/tech-kb.db

# 添加更多文檔到其他知識庫
ollamacli embed-files --dir ~/my-notes --db ~/notes-kb.db

# 分別查詢
ollamacli rag-chat --prompt "技術問題" --db ~/tech-kb.db
ollamacli rag-chat --prompt "個人筆記" --db ~/notes-kb.db
```

## 故障排除

### 問題：找不到相關文檔

確認文檔已正確索引：
```bash
ls -lh ~/.ollamacli/knowledge.db
```

### 問題：索引失敗

檢查 Ollama 是否運行並且嵌入模型已下載：
```bash
ollama list
curl http://localhost:11434/api/tags
```

### 問題：答案質量不佳

嘗試：
1. 增加 `--top-k` 值
2. 調整 `--chunk-size` 和 `--chunk-overlap`
3. 使用更好的聊天模型
