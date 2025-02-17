package services

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/elkamshushi/golangwebserver/models"
)

type PostService struct {
	DB *sql.DB
}

func NewPostService(db *sql.DB) *PostService {
	return &PostService{DB: db}
}

func (s *PostService) GetAllPosts() ([]models.Post, error) {
	query := "SELECT * FROM posts"
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var jsonTags []byte
		err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.Category, &jsonTags, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonTags, &post.Tags)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}
	return posts, nil

}

func (s *PostService) GetAllPostsWithTerm(term string) ([]models.Post, error) {
	query := "SELECT * FROM posts WHERE title LIKE ? OR content LIKE ? OR category LIKE ? OR JSON_CONTAINS(tags, ?)"
	rows, err := s.DB.Query(query, "%"+term+"%", "%"+term+"%", "%"+term+"%", term)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var jsonTags []byte
		err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.Category, &jsonTags, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonTags, &post.Tags)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}
	return posts, nil
}

func (s *PostService) PostAPost(post *models.Post) (*models.Post, error) {
	jsonTags, err := json.Marshal(post.Tags)
	if err != nil {
		return nil, err
	}

	query := "INSERT INTO POSTS (title, content, category, tags) values (?, ?, ?, ?)"
	result, err := s.DB.Exec(query, &post.Title, &post.Content, &post.Category, &jsonTags)
	if err != nil {
		return nil, err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	post.Id = int(lastInsertedId)
	return post, nil

}

func (s *PostService) GetPostById(id string) (*models.Post, error) {
	var post models.Post
	var jsonTags []byte
	query := "SELECT * FROM posts WHERE id = (?)"
	row := s.DB.QueryRow(query, id)
	err := row.Scan(&post.Id, &post.Title, &post.Content, &post.Category, &jsonTags)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonTags, &post.Tags)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostService) DeletePostById(id string) error {
	query := "DELETE FROM posts WHERE id = (?)"
	result, err := s.DB.Exec(query, id)
	if err != nil {
		return err
	}

	AffectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if AffectedRows == 0 {
		return errors.New("post not found")
	}

	return nil
}

// TODO: implement UpdatePost
