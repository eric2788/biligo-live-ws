package updater

type (
	ReleaseLatestResp struct {
		Url             string
		AssetsUrl       string
		UploadUrl       string
		HtmlUrl         string
		Id              int
		NodeId          string
		TagName         string
		TargetCommitish string
		Name            string
		Default         bool
		Prerelease      bool
		CreatedAt       string
		PublishedAt     string
	}
)
