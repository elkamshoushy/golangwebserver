package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elkamshushi/golangwebserver/models"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		term := r.URL.Query().Get("term")
		jsonTerm, err := json.Marshal(term)
		if err != nil {
			http.Error(w, "Error formating term to json", http.StatusInternalServerError)
			log.Println("Error formating term to json:", err)
			return
		}
		var rows *sql.Rows
		if term != "" {
			// Getting posts that have that term in their title, content, category or tags
			query := "SELECT * FROM posts WHERE title LIKE ? OR content LIKE ? OR category LIKE ? OR JSON_CONTAINS(tags, ?)"
			rows, err = db.Query(query, "%"+term+"%", "%"+term+"%", "%"+term+"%", string(jsonTerm))

		} else {
			// Getting all posts
			query := "SELECT * FROM posts"
			rows, err = db.Query(query)
		}

		if err != nil {
			http.Error(w, "Error quering the database", http.StatusInternalServerError)
			log.Println("Error quering the database:", err)
			return
		}

		var posts []models.Post
		for rows.Next() {
			var post models.Post
			var jsonTags []byte
			err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.Category, &jsonTags, &post.CreatedAt, &post.UpdatedAt)
			if err != nil {
				http.Error(w, "Error reading database results", http.StatusInternalServerError)
				log.Println("Error reading database results", err)
				return
			}

			err = json.Unmarshal(jsonTags, &post.Tags)
			if err != nil {
				http.Error(w, "Error Unmarshaling jsonTags to []string", http.StatusInternalServerError)
				log.Println("Error Unmarshaling jsonTags to []string", err)
				return
			}

			posts = append(posts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(posts)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Println("Error encoding response", err)
			return
		}

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error while reading request's body", http.StatusInternalServerError)
			log.Println("Error while reading request's body", err)
			return
		}
		defer r.Body.Close()
		var post models.Post
		err = json.Unmarshal(body, &post)
		if err != nil {
			http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
			log.Println("Error parsing JSON", err)
			return
		}
		jsonTags, err := json.Marshal(post.Tags)
		if err != nil {
			http.Error(w, "Error Marshaling []string tags to json", http.StatusInternalServerError)
			log.Println("Error Marshaling []string tags to json", err)
		}

		query := "INSERT INTO posts (title, content, category, tags) values (?, ?, ?, ?)"
		result, err := db.Exec(query, &post.Title, &post.Content, &post.Category, &jsonTags)
		if err != nil {
			http.Error(w, "Error inserting into Database", http.StatusInternalServerError)
			log.Println("Error inserting into Database", err)
			return
		}
		lastInsertedId, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Error getting last inserted id", http.StatusInternalServerError)
			log.Println("Error gettinglast inserted id", err)
			return
		}
		ptr := &post
		ptr.Id = int(lastInsertedId)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Println("Error encoding response", err)
			return
		}

	default:
		w.Header().Set("Allow", "GET, POST")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method Not Allowed"}`))
	}
}

func singlePostHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/posts/")
	// TODO: it doesnt catch this error idk why
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Invalid or missing ID"`))
		return
	}
	switch r.Method {
	case http.MethodGet:
		row := db.QueryRow("SELECT * FROM posts WHERE id = (?)", id)
		var post models.Post
		var tags string
		err := row.Scan(&post.Id, &post.Title, &post.Content, &post.Category, &tags, &post.CreatedAt, &post.UpdatedAt)

		if err != nil {
			// Checks if no post with that id
			if err == sql.ErrNoRows {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "No post with that ID"}`))
				return
			} else {
				http.Error(w, "Error reading database results", http.StatusInternalServerError)
				log.Println("Error reading database results", err)
				return
			}
		}

		err = json.Unmarshal([]byte(tags), &post.Tags)
		if err != nil {
			http.Error(w, "Error parsing tags", http.StatusInternalServerError)
			log.Println("Error parsing tags", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Println("Error encoding response", http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		deleteQuery := "DELETE FROM posts WHERE id = ?"
		result, err := db.Exec(deleteQuery, id)
		if err != nil {
			http.Error(w, "Error while deleting from database", http.StatusInternalServerError)
			log.Println("Error while deleting from database", err)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Error while checking affected rows", http.StatusInternalServerError)
			log.Println("Error while checking affected rows", err)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "The post is deleted successfully"}`))
	case http.MethodPut:
		// TODO update it from db
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method Not Allowed}`))
	}
}

func main() {
	godotenv.Load()
	dsn := os.Getenv("DB_DSN")

	// Connecting to the db
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening database:", err)
		return
	}

	defer db.Close()

	// Pinging db
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	http.HandleFunc("/posts", postsHandler)           // 'GET' gets all posts or posts with specific term, 'POST' create a post,
	http.HandleFunc("/posts/{id}", singlePostHandler) // 'GET' gets a post, 'DELETE' delete a post, 'UPDATE' update a post,

	// Running the server
	address := "localhost:8080"
	fmt.Println("Server is running at:", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
