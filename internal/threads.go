package internal

import (
	"database/sql"
	"fmt"
	"net/http"
)

func getThreads(w http.ResponseWriter, r *http.Request, categoryID int) {
	category, threads, err := getCategoryAndThreadsWithAuthorByID(categoryID)
	if err != nil {
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Failed to load categories and threads from ID: %v", categoryID), http.StatusInternalServerError))
		Logger.Printf("Failed to retrieve categories and threads by ID, err: %v", err)
		return
	}
	generateHTML(w, r, InitPage(threads, category, getUserCtx(r)), "threads")
}

func getAllCategories() (*[]Category, error) {
	var categories []Category
	rows, err := Db.Query("SELECT id, name FROM categories")
	if err != nil {
		Logger.Printf("Failed to fetch categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() { // iterate through rows retrieved
		var cat Category // instantiate struct variable
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err // keep overwriting and adding to slice
		}
		categories = append(categories, cat)
	}

	return &categories, nil //
}

func getCategoryAndThreadsWithAuthorByID(categoryID int) (*Category, *[]Thread, error) {
	var category Category
	var threads []Thread

	// Fetch category details
	categoryQuery := "SELECT id, name FROM categories WHERE id = ?"
	err := Db.QueryRow(categoryQuery, categoryID).Scan(&category.ID, &category.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("no category found with ID %d", categoryID)
		}
		Logger.Printf("Failed to fetch category: %v", err)
		return nil, nil, err
	}

	// Adjusted query to correctly join and select user data
	query := `
    SELECT t.id, t.uuid, t.topic, t.body, t.created_at, u.id, u.name
    FROM threads t
    LEFT JOIN users u ON t.user_id = u.id
    WHERE t.category_id = ?
    ORDER BY t.created_at ASC`

	rows, err := Db.Query(query, categoryID)
	if err != nil {
		Logger.Printf("Failed to fetch threads for category: %v", err)
		return &category, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var userID int
		var userName string

		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.CreatedAt,
			&userID, &userName, // Scan user details
		); err != nil {
			Logger.Printf("Error scanning threads: %v", err)
			continue
		}

		// Assign user details to Thread.Author
		thread.Author = User{ID: userID, Name: userName}
		threads = append(threads, thread)
	}

	return &category, &threads, nil
}

func getPostsAndAuthorsByThreadIDAscendingWithRating(threadID int) (*[]Post, error) {
	var posts []Post

	query := `
        SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at, 
               u.id, u.uuid, u.name, u.email, u.created_at,
               (SELECT COUNT(*) FROM post_likes WHERE post_id = p.id) AS likes,
               (SELECT COUNT(*) FROM post_dislikes WHERE post_id = p.id) AS dislikes
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
		var likes, dislikes int
		post.PostAuthor = User{} // Initialize the PostAuthor to avoid nil pointer dereference
		if err := rows.Scan(
			&post.ID, &post.UUID, &post.Body, &post.UserID, &post.ThreadID, &post.CreatedAt,
			&post.PostAuthor.ID, &post.PostAuthor.UUID, &post.PostAuthor.Name, &post.PostAuthor.Email, &post.PostAuthor.CreatedAt,
			&likes, &dislikes,
		); err != nil {
			return nil, err
		}
		post.Rating = likes - dislikes
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}

func getDisplayThreadByIDWithRating(threadID int) (*Thread, error) {
	var thread Thread
	thread.Author = User{}       // Initialize Author
	thread.Category = Category{} // Initialize Category

	// Adjusted SQL query to include JOIN with both users and categories tables
	// and to add subqueries for counting likes, dislikes, and views
	query := `SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id, 
                  u.id, u.uuid, u.name, u.email, u.created_at, 
                  c.id, c.name,
                  (SELECT COUNT(*) FROM thread_likes WHERE thread_id = t.id) AS likes,
                  (SELECT COUNT(*) FROM thread_dislikes WHERE thread_id = t.id) AS dislikes,
                  (SELECT COUNT(*) FROM thread_views WHERE thread_id = t.id) AS views
              FROM threads t
              JOIN users u ON t.user_id = u.id 
              JOIN categories c ON t.category_id = c.id
              WHERE t.id = ?`

	var likes, dislikes, views int
	err := Db.QueryRow(query, threadID).Scan(
		&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
		&thread.Author.ID, &thread.Author.UUID, &thread.Author.Name, &thread.Author.Email, &thread.Author.CreatedAt,
		&thread.Category.ID, &thread.Category.Name,
		&likes, &dislikes, &views,
	)
	if err != nil {
		Logger.Printf("Failed to fetch thread with author, category, ratings, and views: %v", err)
		return nil, err
	}

	// Calculate and set the thread rating and views
	thread.Rating = likes - dislikes
	thread.Views = views // Set the thread view count

	return &thread, nil
}

// To populate default state of Search handler
func getAllThreadsSorted(threads *[]Thread) error {

	// Adjust this query to fetch all threads along with their category names,
	// without filtering by a specific category ID.
	query := `
    SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, c.id AS category_id, c.name AS category_name
    FROM threads t
    INNER JOIN categories c ON t.category_id = c.id
    ORDER BY t.created_at DESC` // Using DESC as per the requirement

	rows, err := Db.Query(query)
	if err != nil {
		Logger.Printf("Failed to fetch threads: %v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var category Category
		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt,
			&category.ID, &category.Name,
		); err != nil {
			Logger.Printf("Error scanning threads: %v", err)
			return err
		}
		thread.Category = category // Assuming Thread struct has a Category field of type Category
		*threads = append(*threads, thread)
	}
	if threads == nil || len(*threads) == 0 {
		return fmt.Errorf("no threads matching search criteria")
	}
	return nil
}

// Filter logic for search
func getAllThreadsFilteredbyThread(searchQuery string) ([]Thread, error) {
	var threads []Thread

	// Use parameters in the LIKE clause to prevent SQL injection
	// The % symbols are wildcards for the LIKE operator, allowing for partial matches
	param := "%" + searchQuery + "%"

	// Your SQL query with JOINs to fetch associated category and user,
	// and a WHERE clause to filter threads based on partial match with searchQuery
	query := `
        SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id,
               c.id, c.name,
               u.id, u.uuid, u.name, u.email
        FROM threads t
        INNER JOIN categories c ON t.category_id = c.id
        INNER JOIN users u ON t.user_id = u.id
        WHERE t.topic LIKE ? OR t.body LIKE ?
        ORDER BY t.created_at DESC`

	rows, err := Db.Query(query, param, param) // Reuse param for each LIKE clause
	if err != nil {
		Logger.Printf("Failed to execute filtered threads query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var category Category
		var user User
		// Ensure the order and number of fields in Scan match those selected in the query
		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
			&category.ID, &category.Name,
			&user.ID, &user.UUID, &user.Name, &user.Email,
		); err != nil {
			Logger.Printf("Error scanning filtered thread: %v", err)
			return nil, err
		}
		thread.Category = category
		thread.Author = user // Assuming Thread struct has an Author field of type User
		threads = append(threads, thread)
	}

	return threads, nil
}

func getAllThreadsFilteredbyUser(searchQuery string) ([]Thread, error) {
	var threads []Thread

	// Use parameters in the LIKE clause to prevent SQL injection
	// The % symbols are wildcards for the LIKE operator, allowing for partial matches
	param := "%" + searchQuery + "%"

	// Your SQL query with JOINs to fetch associated category and user,
	// and a WHERE clause to filter threads based on partial match with searchQuery
	query := `
        SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id,
               c.id, c.name,
               u.id, u.uuid, u.name, u.email
        FROM threads t
        INNER JOIN categories c ON t.category_id = c.id
        INNER JOIN users u ON t.user_id = u.id
        WHERE u.name LIKE ? OR u.email LIKE ?
        ORDER BY t.created_at DESC`

	rows, err := Db.Query(query, param, param) // Reuse param for each LIKE clause
	if err != nil {
		Logger.Printf("Failed to execute filtered threads query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var category Category
		var user User
		// Ensure the order and number of fields in Scan match those selected in the query
		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
			&category.ID, &category.Name,
			&user.ID, &user.UUID, &user.Name, &user.Email,
		); err != nil {
			Logger.Printf("Error scanning filtered thread: %v", err)
			return nil, err
		}
		thread.Category = category
		thread.Author = user // Assuming Thread struct has an Author field of type User
		threads = append(threads, thread)
	}

	return threads, nil
}

func getAllThreadsFilteredByPost(searchQuery string) ([]Thread, error) {
	var threads []Thread

	// The % symbols are wildcards for the LIKE operator, allowing for partial matches
	param := "%" + searchQuery + "%"

	// Adjusted SQL query to join threads with posts and filter based on post body's partial match
	query := `
        SELECT DISTINCT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id,
                        c.id, c.name,
                        u.id, u.uuid, u.name, u.email
        FROM threads t
        INNER JOIN categories c ON t.category_id = c.id
        INNER JOIN users u ON t.user_id = u.id
        INNER JOIN posts p ON t.id = p.thread_id
        WHERE p.body LIKE ?
        ORDER BY t.created_at DESC`

	rows, err := Db.Query(query, param)
	if err != nil {
		Logger.Printf("Failed to execute query for threads filtered by post body: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var category Category
		var user User

		if err := rows.Scan(
			&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID,
			&category.ID, &category.Name,
			&user.ID, &user.UUID, &user.Name, &user.Email,
		); err != nil {
			Logger.Printf("Error scanning thread with post body filter: %v", err)
			return nil, err
		}

		thread.Category = category
		thread.Author = user
		threads = append(threads, thread)
	}

	return threads, nil
}

func addCategoryByName(categoryName string) error {
	// Prepare query for checking the existence of the category.
	var categoryID int
	err := Db.QueryRow("SELECT id FROM categories WHERE name = ? LIMIT 1", categoryName).Scan(&categoryID)

	if err == nil {
		// If the category exists, return an appropriate error.
		return fmt.Errorf("category already exists")
	} else if err != sql.ErrNoRows {
		// Log and return any unexpected error during the check.
		Logger.Printf("ERROR | Failed to query category existence: %v", err)
		return err
	}

	// If we reach here, it means the category does not exist and we should insert it.
	_, err = Db.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
	if err != nil {
		// Log and return any error occurred during insertion.
		Logger.Printf("ERROR | Could not insert new category %v: %v", categoryName, err)
		return err
	}

	return nil
}

func removeCategoryByName(categoryName string) error {
	var categoryID int
	// Query to check if the category exists by trying to fetch its ID based on the given name.
	err := Db.QueryRow("SELECT id FROM categories WHERE name = ? LIMIT 1", categoryName).Scan(&categoryID)

	if err == sql.ErrNoRows {
		// If the category does not exist, return an error indicating so.
		return fmt.Errorf("category does not exist")
	} else if err != nil {
		// Log and return any other unexpected error during the check.
		Logger.Printf("ERROR | Failed to query category existence: %v", err)
		return err
	}

	// If we reach here, it means the category exists, and we can proceed to delete it.
	_, err = Db.Exec("DELETE FROM categories WHERE name = ?", categoryName)
	if err != nil {
		// Log and return any error that occurred during the delete operation.
		Logger.Printf("ERROR | Could not delete category %v: %v", categoryName, err)
		return err
	}

	// If the deletion is successful, return nil indicating no errors.
	return nil
}
