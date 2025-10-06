# Web 開發指南

## RESTful API 設計原則

### 什麼是 REST？

REST（Representational State Transfer）是一種軟件架構風格，用於設計網絡應用程序的 API。

### 核心原則

1. **資源導向**: 一切皆資源，使用名詞而非動詞
2. **HTTP 方法**: 使用標準 HTTP 方法表示操作
3. **無狀態**: 每個請求都包含所有必要的信息
4. **統一接口**: 資源通過 URI 唯一標識

### HTTP 方法映射

- `GET /users` - 獲取用戶列表
- `GET /users/123` - 獲取特定用戶
- `POST /users` - 創建新用戶
- `PUT /users/123` - 更新用戶
- `DELETE /users/123` - 刪除用戶

### API 設計最佳實踐

#### 1. 使用正確的 HTTP 狀態碼

```
200 OK - 成功
201 Created - 創建成功
400 Bad Request - 請求錯誤
401 Unauthorized - 未授權
404 Not Found - 資源不存在
500 Internal Server Error - 服務器錯誤
```

#### 2. 版本控制

在 URL 中包含 API 版本：

```
/api/v1/users
/api/v2/users
```

#### 3. 分頁

對於列表資源，實現分頁：

```
GET /api/v1/users?page=1&limit=20
```

響應包含分頁信息：

```json
{
  "data": [...],
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "total_items": 200
  }
}
```

#### 4. 過濾和排序

```
GET /api/v1/users?status=active&sort=created_at:desc
```

#### 5. 錯誤處理

返回結構化的錯誤信息：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email address",
    "details": {
      "field": "email",
      "value": "invalid-email"
    }
  }
}
```

## 認證與授權

### JWT (JSON Web Tokens)

JWT 是現代 Web 應用中常用的認證機制：

1. 用戶登錄後獲得 token
2. 後續請求在 Header 中攜帶 token
3. 服務器驗證 token 的有效性

```
Authorization: Bearer <token>
```

### 安全最佳實踐

1. **使用 HTTPS**: 所有 API 請求都應通過 HTTPS
2. **驗證輸入**: 永遠不要信任客戶端輸入
3. **限流**: 防止 API 濫用
4. **CORS 配置**: 正確配置跨域資源共享

## 數據庫設計

### 關係型數據庫 vs NoSQL

**關係型數據庫（PostgreSQL, MySQL）適用於：**
- 需要 ACID 事務
- 複雜的關聯查詢
- 數據結構明確

**NoSQL（MongoDB, Redis）適用於：**
- 靈活的數據結構
- 高性能讀寫
- 水平擴展

### 數據庫最佳實踐

1. **索引優化**: 為常查詢的字段添加索引
2. **避免 N+1 查詢**: 使用 JOIN 或預加載
3. **連接池**: 重用數據庫連接
4. **遷移管理**: 使用版本化的數據庫遷移

## 性能優化

### 緩存策略

1. **應用層緩存**: Redis, Memcached
2. **CDN**: 靜態資源分發
3. **HTTP 緩存**: 使用 Cache-Control headers

### 負載均衡

分散請求到多個服務器：

```
                  Load Balancer
                /       |        \
            Server1  Server2  Server3
```

### 異步處理

對於耗時操作，使用消息隊列：

1. 接收請求，返回任務 ID
2. 將任務放入隊列
3. Worker 異步處理
4. 客戶端輪詢或使用 WebSocket 獲取結果

## 監控與日誌

### 結構化日誌

使用 JSON 格式的結構化日誌：

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "level": "error",
  "service": "user-api",
  "message": "Database connection failed",
  "error": "connection timeout",
  "user_id": "12345"
}
```

### 監控指標

重要的監控指標包括：

1. **響應時間**: P50, P95, P99
2. **錯誤率**: 4xx, 5xx 錯誤比例
3. **吞吐量**: 每秒請求數（RPS）
4. **資源使用**: CPU, 內存, 數據庫連接

## 微服務架構

### 服務拆分原則

1. **單一職責**: 每個服務只負責一個業務域
2. **獨立部署**: 服務可以獨立發布
3. **數據隔離**: 每個服務有自己的數據庫

### 服務間通信

**同步通信：**
- REST API
- gRPC

**異步通信：**
- 消息隊列（RabbitMQ, Kafka）
- 事件驅動架構

### 服務發現

使用服務註冊中心（如 Consul, Eureka）實現動態服務發現。

## 部署與 CI/CD

### 容器化

使用 Docker 容器化應用：

```dockerfile
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o main .
CMD ["./main"]
```

### 持續集成/持續部署

1. 代碼提交觸發 CI 流程
2. 自動運行測試
3. 構建 Docker 鏡像
4. 部署到測試/生產環境

### 零宕機部署

使用滾動更新或藍綠部署策略，確保服務不中斷。
