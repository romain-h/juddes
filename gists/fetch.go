package gists

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/romain-h/juddes/models"
)

const MAX_PER_PAGE = 30
const GITHUB_USER = "romain-h"
const GITHUB_API = "https://api.github.com"

func getMaxPage(res *http.Response) int {
	link := res.Header.Get("Link")
	pages := strings.Split(link, ",")
	re := regexp.MustCompile(`[\?|\&]page=(\d*)`)
	maxPageStr := re.FindStringSubmatch(pages[1])

	maxPage, _ := strconv.Atoi(maxPageStr[1])
	return maxPage
}

func requestGithub(path string) *http.Response {
	client := &http.Client{}

	url := fmt.Sprintf("%v%v", GITHUB_API, path)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func fullGist(gist *Gist) {
	path := fmt.Sprintf("/gists/%v", gist.ID)
	res := requestGithub(path)
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&gist); err != nil {
		log.Fatal(err)
	}
}

func addContent(gists *[]Gist) {
	var wg sync.WaitGroup
	wg.Add(len(*gists))

	for _, gist := range *gists {
		go func(g Gist) {
			fullGist(&g)
			wg.Done()
		}(gist)
	}
}

func fetchGistsPage(page int) ([]Gist, int) {
	maxPage := 0
	path := fmt.Sprintf(
		"/users/%v/gists?page=%v&per_page=%v",
		GITHUB_USER,
		page,
		MAX_PER_PAGE,
	)

	res := requestGithub(path)
	defer res.Body.Close()
	if page == 1 { // Only parse max page on first page
		maxPage = getMaxPage(res)
	}

	var gists []Gist
	if err := json.NewDecoder(res.Body).Decode(&gists); err != nil {
		log.Fatal(err)
	}
	return gists, maxPage
}

func fetch() []Gist {
	var allGists []Gist
	var wg sync.WaitGroup

	firstPage, maxPage := fetchGistsPage(1)
	addContent(&firstPage)
	wg.Add(maxPage)

	allGists = append(allGists, firstPage...)
	wg.Done()

	if maxPage > 2 {
		for i := 2; i <= maxPage; i++ {
			go func(page int) {
				gists, _ := fetchGistsPage(page)
				addContent(&gists)
				allGists = append(allGists, gists...)
				wg.Done()
			}(i)
		}
	}

	wg.Wait()
	fmt.Printf("Returning final res %v\n", len(allGists))
	return allGists
}

func load(gists []Gist) {
	db, err := models.LoadDB()
	if err != nil {
		log.Fatal(err)
	}

	currentTime := time.Now()
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	tx.Exec("TRUNCATE files;")
	for _, gist := range gists {
		_, err := tx.Exec(`
		INSERT INTO gists (id, url, description, created_at, updated_at, last_loaded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE
		SET url = excluded.url,
		description = excluded.description,
		updated_at = excluded.updated_at,
		last_loaded_at = excluded.last_loaded_at;
		`, gist.ID, gist.URL, gist.Description, gist.CreatedAt, gist.UpdatedAt, currentTime)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range gist.Files {
			_, err := tx.Exec(`
			INSERT INTO files (filename, type, language, content, gist_id)
			VALUES ($1, $2, $3, $4, $5);
			`, file.Filename, file.Type, file.Language, file.Content, gist.ID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Clean deleted gists
	tx.Exec("DELETE FROM gists WHERE last_loaded_at < $1", currentTime)
	tx.Commit()
}

func Search(q string) []string {
	var res []string
	db, err := models.LoadDB()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(`
	SELECT gid
	FROM (SELECT gists.id as gid,
		setweight(to_tsvector(gists.description), 'B') ||
		setweight(to_tsvector(files.filename), 'A') ||
		setweight(to_tsvector(files.content), 'C') as document
	FROM gists
	JOIN files ON files.gist_id = gists.id) p_search
	WHERE p_search.document @@ to_tsquery($1)
	ORDER BY ts_rank(p_search.document, to_tsquery($1)) DESC;
	`, q)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		res = append(res, id)
	}

	return res
}

func Sync() {
	gists := fetch()
	load(gists)
}
