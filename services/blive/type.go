package blive

type LiveInfo struct {
	RoomId          int64  `json:"room_id"`
	UID             int64  `json:"uid"`
	Title           string `json:"title"`
	Name            string `json:"name"`
	Cover           string `json:"cover"`
	UserFace        string `json:"user_face"`
	UserDescription string `json:"user_description"`
}

type ListeningInfo struct {
	*LiveInfo
	// 用於判斷主播類型
	OfficialRole int `json:"official_role"`
}
