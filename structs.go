package main

type Asset struct {
	URL                  string
	ID                   int
	NodeID               string
	Name                 string
	ContentType          string
	State                string
	Size                 int
	DownloadCount        int
	CreatedAt            string
	UpdatedAt            string
	Browser_Download_URL string
}

type Release struct {
	URL             string
	AssetsURL       string
	HTMLURL         string
	ID              int
	NodeID          string
	TagName         string
	TargetCommitish string
	Name            string
	Draft           bool
	Prerelease      bool
	Created_At      string
	Published_At    string
	Assets          []Asset
}

type TagCommit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

type Tag struct {
	Name       string    `json:"name"`
	ZipballURL string    `json:"zipball_url"`
	TarballURL string    `json:"tarball_url"`
	Commit     TagCommit `json:"commit"`
	NodeID     string    `json:"node_id"`
}

type FFMPEGMacInfoDownload struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
	Sig  string `json:"sig"`
}

type FFMPEGMacInfoDownloads struct {
	SevenZip FFMPEGMacInfoDownload `json:"7z"`
	ZIP      FFMPEGMacInfoDownload `json:"zip"`
}

type FFMPEGMacInfoInternalLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type FFMPEGMacInfoExternalLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

type FFMPEGMacInfo struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Version   string                 `json:"version"`
	Size      int                    `json:"size"`
	Download  FFMPEGMacInfoDownloads `json:"download"`
	Libraries struct {
		Internal []FFMPEGMacInfoInternalLibrary `json:"internal"`
		External []FFMPEGMacInfoExternalLibrary `json:"external"`
	}
	RSSFeed string `json:"rss_feed"`
}
