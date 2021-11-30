package main

type Asset struct {
	URL                  string
	ID                   int
	Node_ID              string
	Name                 string
	Content_Type         string
	State                string
	Size                 int
	Download_Count       int
	Created_At           string
	Updated_At           string
	Browser_Download_URL string
}

type Release struct {
	URL              string
	Assets_URL       string
	HTML_URL         string
	ID               int
	Node_ID          string
	Tag_Name         string
	Target_Commitish string
	Name             string
	Draft            bool
	Prerelease       bool
	Created_At       string
	Published_At     string
	Assets           []Asset
}
