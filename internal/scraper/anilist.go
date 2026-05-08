package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const anilistAPI = "https://graphql.anilist.co"

var anilistHTTP = &http.Client{Timeout: 10 * time.Second}

type anilistEpisode struct {
	Title string `json:"title"`
}

func fetchEpisodeTitles(animeName string) ([]string, error) {
	query := `query($s:String){Media(search:$s,type:ANIME){streamingEpisodes{title}}}`
	body := map[string]interface{}{
		"query":     query,
		"variables": map[string]string{"s": animeName},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := anilistPost(jsonBody)
	if err != nil {
		return nil, fmt.Errorf("anilist fetch: %w", err)
	}

	var result struct {
		Data struct {
			Media struct {
				StreamingEpisodes []anilistEpisode `json:"streamingEpisodes"`
			} `json:"Media"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("anilist parse: %w", err)
	}

	var titles []string
	for _, ep := range result.Data.Media.StreamingEpisodes {
		titles = append(titles, ep.Title)
	}

	return titles, nil
}

func anilistPost(body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", anilistAPI, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AniTUI/0.1")

	resp, err := anilistHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
