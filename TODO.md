# TODO

# Complete
[*] 簡化REPL的啟動，當使用ollamacli chat時，直接進入REPL模式，並預載llama3.1:8b模型 - 2025-09-30
[*] REPL增加/save --previous --output $path功能，可以將上一個回應的內容單獨儲存出來 - 2025-09-30
[*] 在REPL模式中實作 /model use $mode_name 功能 - 2025-09-30
[*] 加入文件匯入功能，規劃建立不同知識庫的模式 - 2025-09-30
[*] 加入build custom model功能，可以產出sample modelfile讓我編輯 - 2025-09-30
[*] 修改interactive mode，我希望他像是Claude Code的互動模式，進入一個獨立的迴圈，我問問題然後AI回答，回答完等待下一個問題 - 2025-09-30
[*] 修改Interactive mode使用REPL模式 - 2025-09-30
[*] 在Interactive mode裡面復現 /help, /model list, /model pull ..., /status 等功能 - 2025-09-30
[*] 幫我在互動模式加上一些顏色，並且可以按上下鍵列出之前的指令 - 2025-09-30
[*] 互動模式下，如果不是下slash command，請值接當做問提輸入，並給我回覆 - 2025-09-30
[*] 離開互動模式除了Ctrl+C，請加上exit也可以離開 - 2025-09-30
