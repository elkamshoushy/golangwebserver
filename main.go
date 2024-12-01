package main

import (
	"fmt"
	"net/http"
)

type Post struct {
	Id     int      `json:"id"`
	Author string   `json:"author"`
	Title  string   `json:"title"`
	Tags   []string `json:"tags"`
}

// func allPostsHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "all posts i guess")
// }

func postsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	id := query.Get("id")

	if id != "" {
		fmt.Fprintf(w, "post who's id is %s", id)
	} else {
		fmt.Fprintf(w, "all Posts")
	}

}
func main() {
	http.HandleFunc("/api/v1/posts", postsHandler) // PUT updates post, DELETE post, GET post

	fmt.Println("Server is running")
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		fmt.Println(err.Error())
	}
}
