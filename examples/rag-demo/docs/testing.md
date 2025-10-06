# 測試最佳實踐

## 測試的重要性

軟件測試是確保代碼質量的關鍵環節。良好的測試能夠：

1. **發現 Bug**: 在早期發現問題
2. **文檔化行為**: 測試即文檔
3. **重構信心**: 安全地改進代碼
4. **減少技術債**: 防止代碼腐化

## 測試金字塔

```
        /\
       /  \  E2E Tests (少量)
      /____\
     /      \
    /        \ Integration Tests (適量)
   /__________\
  /            \
 /              \ Unit Tests (大量)
/________________\
```

### 單元測試（Unit Tests）

測試單個函數或方法的行為，快速且隔離。

**特點：**
- 運行快速
- 測試覆蓋率高
- 易於維護

**Go 語言示例：**

```go
package calculator

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}
```

### 集成測試（Integration Tests）

測試多個組件之間的交互，如數據庫、外部 API 等。

**特點：**
- 測試真實環境
- 發現組件間的問題
- 運行較慢

**示例：**

```go
func TestUserRepository(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    repo := NewUserRepository(db)
    user := &User{Name: "Alice", Email: "alice@example.com"}

    err := repo.Create(user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    found, err := repo.FindByEmail("alice@example.com")
    if err != nil {
        t.Fatalf("Failed to find user: %v", err)
    }

    if found.Name != user.Name {
        t.Errorf("Expected name %s, got %s", user.Name, found.Name)
    }
}
```

### 端到端測試（E2E Tests）

測試完整的用戶流程，從前端到後端。

**特點：**
- 最接近真實使用
- 運行最慢
- 維護成本高

## 測試驅動開發（TDD）

TDD 遵循紅-綠-重構循環：

1. **紅**: 編寫一個失敗的測試
2. **綠**: 編寫最少的代碼使測試通過
3. **重構**: 改進代碼質量

### TDD 的優勢

- 更好的代碼設計
- 高測試覆蓋率
- 減少過度設計
- 更快的反饋循環

### TDD 示例流程

```go
// 1. 紅：編寫失敗的測試
func TestUserValidation(t *testing.T) {
    user := User{Email: "invalid-email"}
    err := user.Validate()

    if err == nil {
        t.Error("Expected validation error for invalid email")
    }
}

// 2. 綠：編寫最小代碼使測試通過
func (u *User) Validate() error {
    if !strings.Contains(u.Email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

// 3. 重構：改進代碼
func (u *User) Validate() error {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
```

## Mock 和 Stub

### 為什麼需要 Mock？

在測試中隔離外部依賴（如數據庫、API）：

1. **測試速度**: 不需要真實的外部服務
2. **可控性**: 模擬各種場景（成功、失敗、超時）
3. **獨立性**: 測試不依賴外部狀態

### Go Mock 示例

使用接口實現 Mock：

```go
// 定義接口
type EmailSender interface {
    Send(to, subject, body string) error
}

// Mock 實現
type MockEmailSender struct {
    SentEmails []Email
}

func (m *MockEmailSender) Send(to, subject, body string) error {
    m.SentEmails = append(m.SentEmails, Email{to, subject, body})
    return nil
}

// 測試使用 Mock
func TestSendWelcomeEmail(t *testing.T) {
    mockSender := &MockEmailSender{}
    service := NewUserService(mockSender)

    user := &User{Email: "user@example.com", Name: "Alice"}
    err := service.SendWelcomeEmail(user)

    if err != nil {
        t.Fatalf("Failed to send email: %v", err)
    }

    if len(mockSender.SentEmails) != 1 {
        t.Errorf("Expected 1 email, got %d", len(mockSender.SentEmails))
    }
}
```

## 測試覆蓋率

### 測量覆蓋率

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 覆蓋率目標

- **80%+** : 良好的基準
- **90%+** : 優秀
- **100%** : 通常不必要

### 重要提醒

覆蓋率不是唯一指標：

- 高覆蓋率不等於高質量測試
- 關注測試的有效性而非數字
- 重點測試關鍵業務邏輯

## 測試最佳實踐

### 1. 測試命名

使用描述性名稱：

```go
// 好的命名
func TestUserValidation_InvalidEmail_ReturnsError(t *testing.T) { }

// 不好的命名
func TestUser1(t *testing.T) { }
```

### 2. 表驅動測試

處理多個測試案例：

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", 5, -3, 2},
        {"with zero", 0, 5, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 3. 測試隔離

每個測試應該獨立：

- 不依賴其他測試的執行順序
- 不共享可變狀態
- 清理測試數據

### 4. 測試邊界條件

確保測試：

- 空輸入
- nil 值
- 極大/極小值
- 錯誤情況

### 5. 快速測試

- 使用 mock 代替真實服務
- 避免 sleep
- 並行運行測試（`t.Parallel()`）

## 持續集成中的測試

### CI 流程

```yaml
# GitHub Actions 示例
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```

### 測試階段

1. **提交前**: 快速單元測試
2. **CI 流程**: 完整測試套件
3. **部署前**: 集成和 E2E 測試
4. **生產環境**: 冒煙測試

## 性能測試

### Go Benchmark

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

運行 benchmark：

```bash
go test -bench=. -benchmem
```

## 總結

好的測試實踐：

1. ✅ 遵循測試金字塔
2. ✅ 使用 TDD 方法
3. ✅ 編寫可讀的測試
4. ✅ 保持測試快速
5. ✅ 測試關鍵業務邏輯
6. ✅ 持續運行測試
7. ✅ 重視測試維護

記住：**測試是投資，不是成本！**
