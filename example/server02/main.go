package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/api/v1/t1", Test)
	http.ListenAndServe("localhost:51503", route)
}

func Test(w http.ResponseWriter,r *http.Request) {
	urlStr := fmt.Sprintf("%s%s", "51503:", r.URL.RequestURI())
	w.Write([]byte(urlStr))
}