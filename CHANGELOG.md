.zip 內有 window 和 linux 的點擊運行程序
如欲無需 go 環境，可以下載 zip 直接打開程序運行

更新記錄：

- 短號監聽會自動接駁到真正的房間號
- `/listening/:room_id` 資訊新增 `official_role` 判定主播類型
- 透過 `NO_LISTENING_LOG` 的環境參數禁用 `/listening/*` 的記錄防止洗屏
- 十分鐘冷卻改用新的 set 而不是 listening 防止請求持續頻繁
- 修復檢查更新的 url 不正確的問題