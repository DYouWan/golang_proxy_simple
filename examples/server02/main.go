package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/api/v1", Test1)
	route.HandleFunc("/api/v1/t1", Test2)
	http.ListenAndServe("localhost:51503", route)
}

func Test1(w http.ResponseWriter,r *http.Request) {
	w.WriteHeader(http.StatusOK)
	urlStr := fmt.Sprintf("%s%s", "51503:", r.URL.RequestURI())
	w.Write([]byte(urlStr))
}
func Test2(w http.ResponseWriter,r *http.Request) {
	w.WriteHeader(http.StatusOK)
	urlStr := fmt.Sprintf("%s%s", "51503:", r.URL.RequestURI())
	w.Write([]byte(urlStr))
}