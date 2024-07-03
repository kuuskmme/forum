package internal

import (
	"database/sql"
	"fmt"
	"strings"
)

// This file contains unused utility functions.
func getCategoryAndThreadsByID(categoryID int) (*Category, []Thread, error) {
	var category Category
	var threads []Thread

	// First, fetch the category details regardless of whether it has threads.
	categoryQuery := "SELECT id, name FROM categories WHERE id = ?"
	err := Db.QueryRow(categoryQuery, categoryID).Scan(&category.ID, &category.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("no category found with ID %d", categoryID)
		}
		Logger.Printf("Failed to fetch category: %v", err)
		return nil, nil, err
	}

	// Then, attempt to fetch threads for this category using LEFT JOIN to include the category even if no threads exist.
	query := `
    SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at
    FROM categories c
    LEFT JOIN threads t ON c.id = t.category_id AND c.id = ?
    ORDER BY t.created_at ASC`

	rows, err := Db.Query(query, categoryID)
	if err != nil {
		Logger.Printf("Failed to fetch threads for category: %v", err)
		return &category, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		// Note: No need to scan category.ID and category.Name again since we've already fetched it.
		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt,
		); err != nil {
			Logger.Printf("Error scanning threads: %v", err)
			continue // Handle null values gracefully, allowing for empty threads.
		}
		Logger.Printf("State of thread: %v", thread.Topic)
		threads = append(threads, thread)
	}

	return &category, threads, nil
}

func getThreadByID(threadId int) (*Thread, error) {
	var thread Thread
	query := `SELECT id, uuid, topic, body, user_id, created_at, category_id FROM threads WHERE id = ?`
	err := Db.QueryRow(query, threadId).Scan(&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func getThreadWithAuthorByID(threadId int) (*Thread, error) {
	var thread Thread
	thread.Author = User{} // Initialize the Author field to avoid nil pointer dereference

	// Adjusted SQL query to perform a JOIN with the users table
	query := `SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id, 
              u.id, u.uuid, u.name, u.email, u.created_at 
              FROM threads t
              JOIN users u ON t.user_id = u.id 
              WHERE t.id = ?`

	err := Db.QueryRow(query, threadId).Scan(
		&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
		&thread.Author.ID, &thread.Author.UUID, &thread.Author.Name, &thread.Author.Email, &thread.Author.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func getPostByID(threadID int) (*Post, error) {
	var post Post
	query := `SELECT id, uuid, body, approved, user_id, thread_id, created_at FROM posts WHERE id = $1`
	err := Db.QueryRow(query, threadID).Scan(&post.ID, &post.UUID, &post.Body, &post.Approved, &post.UserID, &post.ThreadID, &post.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// Get Posts by created_at ascending (Oldest first)
func getPostsByIDAscending(threadID int) (*[]Post, error) {
	var posts []Post
	// Add ORDER BY created_at ASC to ensure the posts are returned in ascending order
	query := `SELECT id, uuid, body, approved, user_id, thread_id, created_at FROM posts WHERE thread_id = $1 ORDER BY created_at ASC`
	rows, err := Db.Query(query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UUID, &post.Body, &post.Approved, &post.UserID, &post.ThreadID, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		Logger.Printf("Error when querying rows. Data may be malformed")
		return nil, err
	}

	return &posts, nil
}

func (thread *Thread) getAuthor() error {
	user := User{}
	err := Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id = $1", thread.UserID).
		Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		Logger.Printf("Failed to return author for thread %v. Error: %v", thread, err)
		return fmt.Errorf("failed to return author for thread %v. Error: %v", thread, err)
	}
	thread.Author = user
	return nil
}

// Retrieve all Users associated with Posts in the current Thread and returns a map of userID to User
func getPostsAuthors(posts *[]Post) error {
	userMap := make(map[int]User) // Map userID - User

	// Deduplicate userIDs and prepare args for the query
	var userIDs []int
	for _, post := range *posts {
		if _, exists := userMap[post.UserID]; !exists {
			userIDs = append(userIDs, post.UserID)
		}
	}

	// Return early if no userIDs to prevent invalid SQL query execution
	if len(userIDs) == 0 {
		return nil
	}

	// Convert userIDs to []interface{} for the query
	args := make([]interface{}, len(userIDs))
	for i, userID := range userIDs {
		args[i] = userID
	}

	// Prepare the IN clause with correct placeholders
	inClause := strings.Repeat(",?", len(args)-1)
	query := fmt.Sprintf("SELECT id, uuid, name, email, created_at FROM users WHERE id IN (?%s)", inClause)

	rows, err := Db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Populate userMap with actual fetched users
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return err
		}
		userMap[user.ID] = user // Store fetched user in map
	}
	for i, post := range *posts {
		// Update the slice element directly using its index
		(*posts)[i].PostAuthor = userMap[post.UserID]
	}
	Logger.Printf("%v", (*posts)[0].PostAuthor.Name)
	return nil
}

func getPostsAndAuthorsByThreadIDAscending(threadID int) (*[]Post, error) {
	var posts []Post

	// Query to fetch posts and their author's details using JOIN
	query := `
        SELECT p.id, p.uuid, p.body, p.approved, p.user_id, p.thread_id, p.created_at, 
               u.id, u.uuid, u.name, u.email, u.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.thread_id = ? ORDER BY p.created_at ASC
    `

	rows, err := Db.Query(query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		post.PostAuthor = User{} // Initialize the PostAuthor to avoid nil pointer dereference
		if err := rows.Scan(
			&post.ID, &post.UUID, &post.Body, &post.Approved, &post.UserID, &post.ThreadID, &post.CreatedAt,
			&post.PostAuthor.ID, &post.PostAuthor.UUID, &post.PostAuthor.Name, &post.PostAuthor.Email, &post.PostAuthor.CreatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}

func getDisplayThreadByID(threadID int) (*Thread, error) {
	var thread Thread
	thread.Author = User{}       // Initialize Author
	thread.Category = Category{} // Initialize Category

	// Adjusted SQL query to include JOIN with both users and categories tables
	query := `SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id, 
              u.id, u.uuid, u.name, u.email, u.created_at, 
              c.id, c.name 
              FROM threads t
              JOIN users u ON t.user_id = u.id 
              JOIN categories c ON t.category_id = c.id
              WHERE t.id = ?`

	err := Db.QueryRow(query, threadID).Scan(
		&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
		&thread.Author.ID, &thread.Author.UUID, &thread.Author.Name, &thread.Author.Email, &thread.Author.CreatedAt,
		&thread.Category.ID, &thread.Category.Name,
	)
	if err != nil {
		Logger.Printf("Failed to fetch thread with author and category: %v", err)
		return nil, err
	}

	return &thread, nil
}

func getThreadsByID(categoryID int) (*[]Thread, error) {
	var threads []Thread
	query := `SELECT id, uuid, topic, body, user_id, created_at, category_id FROM threads WHERE category_id = ?`
	rows, err := Db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID); err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	return &threads, nil
}
