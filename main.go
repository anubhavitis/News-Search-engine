package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var tpl = template.Must(template.ParseFiles("index.html"))
var apiKey = "5db4d29d8e334655862b2a42bcb17307"

// Results final struct for the response from the API
type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}
type Article struct {
	Source struct {
		ID   interface{} `json:"id"`
		Name string      `json:"name"`
	} `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// FormatPublishedDate formats date to normal form
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

// Search query
type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

func (s *Search) GoToNext() int {
	return s.NextPage + 1
}

func (s *Search) PreviousPage() int {
	return s.NextPage - 1
}
func indexHandler(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	search := &Search{}
	search.SearchKey = searchKey
	//converts page variable to integer and stores it in next
	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}
	if searchKey == "" {
		err = tpl.Execute(w, search)
		if err != nil {
			log.Println(err)
			return
		}
	}
	search.NextPage = next
	pageSize := 20

	//Sprintf is used to format stirng with given variables
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en",
		url.QueryEscape(search.SearchKey), pageSize, search.NextPage, apiKey)

	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	search.TotalPages = search.Results.TotalResults / pageSize
	if search.Results.TotalResults%pageSize != 0 {
		search.TotalPages++
	}

	err = tpl.Execute(w, search)
	if err != nil {
		log.Println(err)
	}

}

func main() {

	mux := http.NewServeMux()
	assets := http.FileServer(http.Dir("css"))
	mux.Handle("/css/", http.StripPrefix("/css/", assets))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, mux)
}
