package main

import (
	"net/http"
	"text/template"
)

var tpl = template.Must(template.ParseFiles("index.html"))

func user(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, nil)
}
func main() {
	mux := http.NewServeMux()
	assets := http.FileServer(http.Dir("css"))
	mux.Handle("/css/", http.StripPrefix("/css/", assets))
	mux.HandleFunc("/", user)
	http.ListenAndServe(":8080", mux)
}
