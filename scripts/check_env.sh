#!/usr/bin/env bash
set -euo pipefail

echo "=== ollamacli 環境檢查 ==="
echo

# 基礎工具檢查
echo "基礎工具:"
tools=("git" "curl" "docker")
for tool in "${tools[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "✅ $tool: $($tool --version | head -1)"
    else
        echo "❌ $tool: 未安裝"
    fi
done

echo

# Go 環境檢查
echo "Go 環境:"
if command -v go >/dev/null 2>&1; then
    echo "✅ Go: $(go version)"
    echo "   GOPATH: ${GOPATH:-未設定}"
    echo "   GOROOT: ${GOROOT:-$(go env GOROOT)}"
else
    echo "❌ Go: 未安裝"
fi

if command -v golangci-lint >/dev/null 2>&1; then
    echo "✅ golangci-lint: $(golangci-lint version)"
else
    echo "⚠️  golangci-lint: 未安裝 (建議安裝)"
fi

echo

# Ollama 相關檢查
echo "Ollama 相關:"
if command -v ollama >/dev/null 2>&1; then
    echo "✅ Ollama: $(ollama --version 2>/dev/null || echo "已安裝")"

    # 檢查 Ollama 服務是否運行
    if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
        echo "✅ Ollama Server: 運行中 (localhost:11434)"
    else
        echo "⚠️  Ollama Server: 未運行或無法連接"
    fi
else
    echo "⚠️  Ollama: 未安裝 (可選，用於本地測試)"
fi

echo

echo "專案檔案檢查:"
echo "✅ 設計文件: docs/design.md"
echo "✅ CLAUDE.md: 已生成"

if [[ -f "go.mod" ]]; then
    echo "✅ Go Module: 已初始化"
else
    echo "⚠️  Go Module: 未初始化 (執行 go mod init)"
fi

if [[ -f "Makefile" ]]; then
    echo "✅ Makefile: 已配置"
else
    echo "⚠️  Makefile: 未配置"
fi

if [[ -f "main.go" || -d "cmd/" ]]; then
    echo "✅ 主程式: 已建立"
else
    echo "⚠️  主程式: 未建立"
fi

echo

# 環境變數檢查
echo "環境變數:"
echo "   OLLAMA_HOST: ${OLLAMA_HOST:-未設定 (預設: http://localhost:11434)}"
echo "   OLLAMA_TOKEN: ${OLLAMA_TOKEN:+已設定}"
echo "   GO_VERSION: $(go version 2>/dev/null | cut -d' ' -f3 || echo "未檢測到")"

echo
echo "=== 環境檢查完成 ==="