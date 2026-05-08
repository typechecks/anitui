package models

type Anime struct {
	Title       string
	URL         string
	Description string
	CoverURL    string
	Source      string
}

type SearchResult struct {
	Anime  []Anime
	Source string
	Error  error
}
