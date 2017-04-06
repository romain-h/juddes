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

	"github.com/romain-h/juddes/models"
)

const MAX_PER_PAGE = 30

func getMaxPage(res *http.Response) int {
	link := res.Header.Get("Link")
	pages := strings.Split(link, ",")
	re := regexp.MustCompile(`[\?|\&]page=(\d*)`)
	maxPageStr := re.FindStringSubmatch(pages[1])

	maxPage, _ := strconv.Atoi(maxPageStr[1])
	return maxPage
}

func fetchGistPage(page int) ([]Gist, int) {
	maxPage := 0
	client := &http.Client{}

	url := fmt.Sprintf(
		"https://api.github.com/users/romain-h/gists?page=%v&per_page=%v",
		page,
		MAX_PER_PAGE,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

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
	gists := make(chan []Gist)

	go func() {
		for res := range gists {
			allGists = append(allGists, res...)
		}
	}()

	firstPage, maxPage := fetchGistPage(1)
	wg.Add(maxPage)

	go func(res []Gist) {
		defer wg.Done()
		gists <- res
	}(firstPage)

	if maxPage > 2 {
		for i := 2; i <= maxPage; i++ {
			go func(page int) {
				defer wg.Done()
				res, _ := fetchGistPage(page)
				gists <- res
			}(i)
		}
	}

	wg.Wait()
	fmt.Printf("Returning final res %v", len(allGists))
	return allGists
}

func load(gists []Gist) {
	db, err := models.LoadDB()
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, gist := range gists {
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

func Sync() {
	gists := fetch()
	load(gists)
}
