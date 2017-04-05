package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/romain-h/juddes/models"
)

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

func fetchGist() []Gist {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://api.github.com/users/romain-h/gists", nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	var gists []Gist
	if err := json.NewDecoder(res.Body).Decode(&gists); err != nil {
		log.Fatal(err)
	}
	return gists
}

func loadGists(gists []Gist) {
	db, err := models.LoadDB()
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, gist := range gists {
		fmt.Println("%s\n", gist)
		_, err := tx.Exec(`
		INSERT INTO gists (id, url, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET url = excluded.url,
		description = excluded.description;
		`, gist.ID, gist.URL, gist.Description)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}

func main() {
	gists := fetchGist()
	// fmt.Println(gists)
	loadGists(gists)
}
