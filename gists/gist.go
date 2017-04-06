package gists

import "time"

type File struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"langugae"`
	Raw_url  string `json:"raw_url"`
}

type Gist struct {
	URL         string          `json:"url"`
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Files       map[string]File `json:"files"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
