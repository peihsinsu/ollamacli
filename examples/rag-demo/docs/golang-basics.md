# Go 語言基礎知識

## 什麼是 Go？

Go（也稱為 Golang）是 Google 開發的開源編程語言，於 2009 年首次發布。Go 的設計目標是提供一種簡單、高效、可靠的編程語言，特別適合構建大規模的並發系統和網絡服務。

## 主要特點

### 1. 簡潔的語法

Go 的語法簡潔明了，沒有過多的語法糖。這使得代碼易於閱讀和維護。

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 2. 內建並發支持

Go 通過 goroutines 和 channels 提供了強大的並發支持：

- **Goroutines**: 輕量級的線程，可以輕鬆創建數千個並發任務
- **Channels**: 用於 goroutines 之間的通信和同步

```go
go func() {
    // 並發執行的代碼
}()
```

### 3. 快速編譯

Go 編譯器非常快，能夠在幾秒鐘內編譯大型項目。

### 4. 靜態類型與類型推斷

Go 是靜態類型語言，但提供了類型推斷功能：

```go
var x int = 10    // 明確類型
y := 20           // 類型推斷
```

### 5. 垃圾回收

Go 有自動垃圾回收機制，開發者不需要手動管理內存。

## 核心概念

### 包（Packages）

Go 程序由包組成。每個源文件都屬於一個包：

```go
package main  // 可執行程序的入口包
```

### 函數（Functions）

函數定義語法簡潔：

```go
func add(a int, b int) int {
    return a + b
}

// 參數類型相同時可以簡寫
func multiply(a, b int) int {
    return a * b
}
```

### 結構體（Structs）

Go 使用結構體而不是類來組織數據：

```go
type Person struct {
    Name string
    Age  int
}

person := Person{Name: "Alice", Age: 30}
```

### 接口（Interfaces）

接口定義行為，實現是隱式的：

```go
type Writer interface {
    Write(data []byte) (int, error)
}
```

### 錯誤處理

Go 使用顯式的錯誤處理而不是異常：

```go
result, err := someFunction()
if err != nil {
    // 處理錯誤
    return err
}
```

## 工具鏈

### go build

編譯 Go 程序：

```bash
go build main.go
```

### go run

編譯並運行：

```bash
go run main.go
```

### go test

運行測試：

```bash
go test ./...
```

### go mod

管理依賴：

```bash
go mod init myproject
go mod tidy
```

## 適用場景

Go 特別適合以下場景：

1. **Web 服務和 API**: 內建的 net/http 包功能強大
2. **微服務架構**: 輕量級、快速部署
3. **CLI 工具**: 單一二進制文件，易於分發
4. **並發處理**: goroutines 使並發編程變得簡單
5. **DevOps 工具**: 如 Docker、Kubernetes 都是用 Go 編寫

## 最佳實踐

1. **遵循命名規範**: 使用駝峰命名法
2. **錯誤優先處理**: 總是首先檢查錯誤
3. **保持代碼簡潔**: 避免過度設計
4. **使用 gofmt**: 統一代碼格式
5. **編寫測試**: 使用 testing 包編寫單元測試
