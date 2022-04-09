package updater

type ReleaseLatestResp struct {
	Url             string `json:"url"`
	AssetsUrl       string `json:"assets_url"`
	UploadUrl       string `json:"upload_url"`
	HtmlUrl         string `json:"html_url"`
	Id              int    `json:"id"`
	NodeId          string `json:"node_id"`
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Default         bool   `json:"default"`
	Prerelease      bool   `json:"prerelease"`
	CreatedAt       string `json:"created_at"`
	PublishedAt     string `json:"published_at"`
}
