package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/api/v1/", Test)
	http.ListenAndServe("localhost:51505", route)
}

func Test(w http.ResponseWriter,r *http.Request) {
	w.WriteHeader(http.StatusOK)
	urlStr := fmt.Sprintf("%s%s", "51505:", r.URL.RequestURI())
	w.Write([]byte(urlStr))
}