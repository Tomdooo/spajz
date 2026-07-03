package main

import (
	"fmt"
	"net/http"

	"github.com/Tomdooo/storos/internal/api"
	"github.com/Tomdooo/storos/internal/buckets"
)

func main() {
	err := buckets.Create("test")
	if err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/upload", api.UploadHandler)
	http.ListenAndServe(":8080", nil)
}
