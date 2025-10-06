# ADR 003: RAG (Retrieval Augmented Generation) 實作

## 狀態

已接受 (Accepted) - 2025-01-15

## 背景

用戶希望能夠限制 AI 模型回答問題時使用的文件範圍，實現基於特定文檔的問答。這需要實現 RAG (Retrieval Augmented Generation) 功能。

### 需求

1. 支援從文件建立向量索引
2. 基於查詢檢索相關文檔片段
3. 將檢索到的內容作為上下文提供給 LLM
4. 本地存儲，不依賴外部服務

## 決策

我們決定採用以下技術方案實作 RAG：

### 1. 向量數據庫：SQLite + Pure Go

**選擇 `modernc.org/sqlite` 的原因：**

- **純 Go 實現**：不需要 CGO，簡化跨平台編譯
- **零依賴**：用戶不需要安裝額外的數據庫
- **輕量級**：適合 CLI 工具的定位
- **熟悉的 SQL 接口**：開發和調試方便

**替代方案及拒絕原因：**

- ❌ **專用向量數據庫（Qdrant, Milvus）**：過於複雜，需要額外服務
- ❌ **Pinecone, Weaviate 等雲服務**：違反本地優先原則
- ❌ **sqlite-vec C 擴展**：需要 CGO，增加編譯複雜度

### 2. 相似度計算：餘弦相似度

**選擇原因：**

- 業界標準，效果驗證
- 實現簡單，純 Go 計算
- 適合文本嵌入向量

**實現：**

```go
func cosineSimilarity(a, b []float64) float64 {
    dotProduct := 0.0
    normA, normB := 0.0, 0.0

    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

**性能考量：**

- 對於小到中等規模的知識庫（< 10,000 文檔），線性掃描可接受
- 未來可優化：HNSW、LSH 等近似搜索算法

### 3. 文本分塊策略

**默認配置：**

- **chunk_size**: 500 字符
- **chunk_overlap**: 50 字符

**分塊算法：**

1. 優先在句子邊界分割（`.`, `!`, `?`）
2. 次選在詞邊界分割（空格）
3. 最後按字符數硬切割

**可配置性：**

```bash
ollamacli embed-files --chunk-size 800 --chunk-overlap 80
```

### 4. 模塊架構

```
internal/rag/
├── store.go        # 向量存儲接口定義
├── vectordb.go     # SQLite 實現
├── chunker.go      # 文本分塊器
├── retriever.go    # 檢索器（整合組件）
└── *_test.go       # 單元測試
```

**設計原則：**

- **接口驅動**：Store 接口便於未來替換實現
- **職責分離**：chunker, store, retriever 各司其職
- **依賴注入**：便於測試和擴展

### 5. CLI 指令設計

#### `embed-files` - 建立知識庫

```bash
ollamacli embed-files \
  --files file1.md --files file2.txt  # 或
  --dir ./docs --patterns "*.md"      # 批量索引
  --model mxbai-embed-large           # 嵌入模型
  --db ~/.ollamacli/knowledge.db      # 數據庫路徑
  --chunk-size 500                    # 可選配置
```

#### `rag-chat` - 基於知識庫問答

```bash
ollamacli rag-chat [MODEL] \
  --prompt "你的問題" \
  --top-k 3                           # 檢索文檔數
  --embed-model mxbai-embed-large     # 嵌入模型
  --db ~/.ollamacli/knowledge.db      # 知識庫路徑
```

### 6. 配置文件支持

擴展 `~/.ollamacli/config.yaml`：

```yaml
rag:
  knowledge_base: ~/.ollamacli/knowledge.db
  embed_model: mxbai-embed-large
  chunk_size: 500
  chunk_overlap: 50
  allowed_files:
    - "*.md"
    - "*.txt"
    - "*.go"
```

## 實現細節

### 數據庫 Schema

```sql
CREATE TABLE documents (
    id TEXT PRIMARY KEY,            -- SHA256 hash
    content TEXT NOT NULL,          -- 文本內容
    source TEXT NOT NULL,           -- 源文件路徑
    embedding TEXT NOT NULL,        -- JSON 格式的向量
    metadata TEXT,                  -- JSON 格式的元數據
    created_at DATETIME NOT NULL
);

CREATE INDEX idx_documents_source ON documents(source);
```

### 工作流程

#### 索引階段

1. 讀取文件內容
2. 按配置分塊（chunker）
3. 調用 Ollama API 生成嵌入
4. 存儲到 SQLite（vectordb）

#### 檢索階段

1. 將用戶查詢轉為嵌入向量
2. 計算與所有文檔的餘弦相似度
3. 返回 top-k 個最相關的文檔
4. 組裝為 prompt 上下文
5. 調用 LLM 生成答案

## 後果

### 優點

✅ **簡單易用**：兩條命令即可實現 RAG
✅ **本地優先**：無需外部依賴或雲服務
✅ **可擴展**：接口設計便於未來優化
✅ **跨平台**：純 Go，無 CGO 依賴

### 缺點

❌ **性能限制**：線性搜索不適合超大規模知識庫
❌ **功能簡化**：相比專業向量數據庫功能較少
❌ **內存占用**：大量文檔需要載入內存計算

### 風險與緩解

| 風險 | 緩解措施 |
|------|----------|
| 性能瓶頸 | 文檔建議限制在 10,000 以內 |
| 嵌入模型變更 | 記錄使用的模型，提示不一致 |
| 數據庫損壞 | 提供導出/導入功能 |

## 未來改進

### 短期（v1.1）

- [ ] 添加 `rag list-sources` 指令查看已索引文件
- [ ] 支援刪除特定源的文檔
- [ ] 添加 `rag stats` 顯示知識庫統計

### 中期（v1.2）

- [ ] 實現基於 HNSW 的近似搜索優化
- [ ] 支援多模態嵌入（圖片、代碼）
- [ ] 增量更新（只重新索引變更的文件）

### 長期（v2.0）

- [ ] 支援其他向量數據庫後端（可選）
- [ ] 實現混合搜索（向量 + 關鍵詞）
- [ ] 分布式知識庫（多節點同步）

## 相關資源

- [Ollama Embeddings API](https://github.com/ollama/ollama/blob/main/docs/api.md#generate-embeddings)
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)
- [RAG 原理論文](https://arxiv.org/abs/2005.11401)
- [Cosine Similarity](https://en.wikipedia.org/wiki/Cosine_similarity)

## 變更歷史

- 2025-01-15: 初始版本，RAG 功能設計和實現
