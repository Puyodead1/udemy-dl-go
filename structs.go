package main

type Asset struct {
	URL                  string `json:"url"`
	ID                   int    `json:"id"`
	NodeID               string `json:"node_id"`
	Name                 string `json:"name"`
	ContentType          string `json:"content_type"`
	State                string `json:"state"`
	Size                 int    `json:"size"`
	DownloadCount        int    `json:"download_count"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
	Browser_Download_URL string `json:"browser_download_url"`
}

type Release struct {
	URL             string  `json:"url"`
	AssetsURL       string  `json:"assets_url"`
	HTMLURL         string  `json:"html_url"`
	ID              int     `json:"id"`
	NodeID          string  `json:"node_id"`
	TagName         string  `json:"tag_name"`
	TargetCommitish string  `json:"target_commitish"`
	Name            string  `json:"name"`
	Draft           bool    `json:"draft"`
	Prerelease      bool    `json:"prerelease"`
	Created_At      string  `json:"created_at"`
	Published_At    string  `json:"published_at"`
	Assets          []Asset `json:"assets"`
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
