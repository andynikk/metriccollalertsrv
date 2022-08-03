package main

import (
	"net/http"
)

func BodyText(w http.ResponseWriter, r *http.Request) {

}

func main() {

	http.HandleFunc("/", BodyText)
	http.ListenAndServe(":8080", nil)

}
