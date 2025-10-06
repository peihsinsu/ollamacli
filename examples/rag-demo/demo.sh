#!/bin/bash

# RAG Demo Script for ollamacli
# This script demonstrates the RAG (Retrieval Augmented Generation) functionality

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}ollamacli RAG Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if ollamacli is available
if ! command -v ollamacli &> /dev/null; then
    echo -e "${YELLOW}Warning: ollamacli not found in PATH${NC}"
    echo -e "${YELLOW}Trying to use local build...${NC}"
    OLLAMACLI="../../build/ollamacli"
    if [ ! -f "$OLLAMACLI" ]; then
        echo -e "${YELLOW}Please build ollamacli first: make build${NC}"
        exit 1
    fi
else
    OLLAMACLI="ollamacli"
fi

# Step 1: Index documents
echo -e "${GREEN}Step 1: Indexing documents...${NC}"
echo "This will process all markdown files in the docs/ directory"
echo ""

$OLLAMACLI embed-files \
  --dir ./docs \
  --patterns "*.md" \
  --model mxbai-embed-large

echo ""
echo -e "${GREEN}Indexing complete!${NC}"
echo ""
echo -e "${BLUE}========================================${NC}"

# Step 2: Ask questions
echo -e "${GREEN}Step 2: Asking questions based on the knowledge base${NC}"
echo ""

questions=(
    "Go 語言有什麼特點？"
    "什麼是 REST API？"
    "測試金字塔是什麼？"
    "如何實現微服務架構？"
    "TDD 的工作流程是什麼？"
)

for question in "${questions[@]}"; do
    echo -e "${YELLOW}Question: ${question}${NC}"
    echo ""

    $OLLAMACLI rag-chat --prompt "$question"

    echo ""
    echo -e "${BLUE}----------------------------------------${NC}"
    echo ""

    # Optional: wait for user to continue
    # read -p "Press Enter to continue to next question..."
    # echo ""
done

echo -e "${GREEN}Demo complete!${NC}"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Next steps:${NC}"
echo "1. Try your own questions: ollamacli rag-chat --prompt \"your question\""
echo "2. Adjust retrieval: ollamacli rag-chat --prompt \"...\" --top-k 5"
echo "3. Use different model: ollamacli rag-chat llama3.2 --prompt \"...\""
echo -e "${BLUE}========================================${NC}"
