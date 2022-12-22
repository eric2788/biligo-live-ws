.zip 內有 window 和 linux 的點擊運行程序
如欲無需 go 環境，可以下載 zip 直接打開程序運行

更新記錄：

- 修复 goroutine leaks 导致的内存泄漏
- heartbeat 自動斷線逾时改为三分钟
- 优化内存使用量
- 数据库新增闲置时关闭以进一步减少内存使用量