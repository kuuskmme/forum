package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Create a new user and add to db

func (user *User) Create() error {
	// Generate a new UUID for the user.
	user.UUID = uuid.New().String()

	// Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		Logger.Printf("ERROR | Could not hash password: %v", err)
		return err
	}

	// Prepare the insert statement.
	stmt, err := Db.Prepare("INSERT INTO users (uuid, name, email, password, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		Logger.Printf("ERROR | Could not prepare user insert statement: %v", err)
		return err
	}
	defer stmt.Close()
	// Format according to constant that defines the global DateTime that we use.
	currentTime := time.Now().Format(DateTime)
	// Execute the insert statement.
	_, err = stmt.Exec(user.UUID, user.Name, user.Email, string(hashedPassword), currentTime)
	if err != nil {
		Logger.Printf("ERROR | Could not insert new user %v: %v", user, err)
		return err
	}
	user.CreatedAt = currentTime // Make sure the user has the createdAt also
	Logger.Printf("Proceeding to create title")
	err = user.CreateTitle()
	if err != nil {
		Logger.Printf("Could not create title!")
	}
	// Creation successful
	Logger.Printf("User %v created successfully!", user)
	return nil
}

func (user *User) Check() error {
	user.UUID = uuid.New().String()

	// Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	currentTime := time.Now().Format(DateTime)
	if err != nil {
		Logger.Printf("ERROR | Could not hash password: %v", err)
		return err
	}
	tx, err := Db.Begin()
	if err != nil {
		Logger.Printf("ERROR | Could not start transaction: %v", err)
		return err
	}

	// Prepare the insert statement within the transaction.
	stmtTx, err := tx.Prepare("INSERT INTO users (uuid, name, email, password, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		Logger.Printf("ERROR | Could not prepare user verification insert statement: %v", err)
		return err
	}
	_, err = stmtTx.Exec(user.UUID, user.Name, user.Email, string(hashedPassword), currentTime)
	if err != nil {
		Logger.Printf("ERROR | Could not insert new user %v: %v", user, err)
		tx.Rollback() // Roll back the tx anyway
		return err
	}
	tx.Rollback() // Roll back the tx anyway
	defer stmtTx.Close()
	return err

}
func (user *User) CreateTitle() error {
	// Now, prepare to insert into the titles table
	err := Db.QueryRow("SELECT id from users WHERE uuid = $1", user.UUID).Scan(&user.ID)
	if err != nil {
		Logger.Printf("ERROR | Could not SELECT user via UUID?")
		return err
	}
	if err != nil {
		Logger.Printf("ERROR | Could not prepare title insert statement: %v", err)
		return err
	}
	stmtTitle, err := Db.Prepare("INSERT INTO titles (user_id) VALUES (?)")
	if err != nil {
		Logger.Printf("ERROR | Could not prepare title insert statement: %v", err)
		return err
	}
	defer stmtTitle.Close()

	// Execute with the user's ID
	_, err = stmtTitle.Exec(user.ID)
	if err != nil {
		Logger.Printf("ERROR | Could not insert title for new user: %v", err)
		return err
	}
	return nil
}

func (user *User) RefreshCookies(w http.ResponseWriter, r *http.Request, targetPath string) {
	// Assuming CreateSession() returns a session object with a UUID
	session, err := user.CreateSession()
	if err != nil {
		Logger.Printf("Failed to create session for user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

	}

	// Set the session cookie
	setSessionCookie(w, session.UUID)

	// Redirect to the target path which is home if arriving from Register handler//
	generateHTML(w, r, InitPage(r.Context().Value(UserCtxKey)), "home", fmt.Sprintf("%v", targetPath))
}

func (user *User) GetByUUID(UUID string) (err error) {
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE uuid = $1", UUID).
		Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		Logger.Printf("Failed to retrive User %v via UUID %v.\nError: %v", user, UUID, err)
	}
	return
}

// Retrieves all threads created by the user
func (user *User) GetThreads() ([]Thread, error) {
	var threads []Thread
	// SQL query matches all category ids with corresponding category when returning the data.
	query := `
        SELECT t.id, t.uuid, t.topic, t.body, t.user_id, t.created_at, t.category_id, 
               c.id, c.name 
        FROM threads t
        INNER JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1 
        ORDER BY t.created_at DESC
    `
	rows, err := Db.Query(query, user.ID)
	if err != nil {
		Logger.Printf("Failed to find threads for user: %v", err)
		return nil, err
	}
	defer rows.Close() // Ensure that rows are closed after function execution

	for rows.Next() {
		var thread Thread
		var category Category
		// Adjust the Scan to include the category data
		if err = rows.Scan(&thread.ID, &thread.UUID, &thread.Topic, &thread.Body, &thread.UserID, &thread.CreatedAt, &thread.CategoryID, &category.ID, &category.Name); err != nil {
			Logger.Printf("Failed to scan thread and category from db. Error: %v", err)
			return nil, err // Return immediately on error
		}
		thread.Category = category        // Associate the fetched category with the thread
		threads = append(threads, thread) // Append each successfully scanned thread
	}
	if err = rows.Err(); err != nil {
		Logger.Printf("Error occurred during rows iteration: %v", err)
		return nil, err
	}

	Logger.Printf("Successfully finished loading threads for user %v", user.Name)
	return threads, nil
}

func (user User) GetPostsWithTopic() ([]Post, error) {
	var posts []Post
	query := `
        SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at, t.topic AS thread_topic
        FROM posts p
        JOIN threads t ON p.thread_id = t.id
        WHERE p.user_id = ?
        ORDER BY p.created_at DESC`
	rows, err := Db.Query(query, user.ID)
	if err != nil {
		Logger.Printf("Failed to execute query to get user's posts with thread topics: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var threadTopic string
		if err = rows.Scan(&post.ID, &post.UUID, &post.Body, &post.UserID, &post.ThreadID, &post.CreatedAt, &threadTopic); err != nil {
			Logger.Printf("Failed to scan post and thread topic from db. Error: %v", err)
			return nil, err
		}
		post.ThreadTopic = threadTopic
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		Logger.Printf("Error occurred during rows iteration for user's posts with thread topics: %v", err)
		return nil, err
	}

	Logger.Printf("Successfully finished loading posts with thread topics for user %v", user.Name)
	return posts, nil
}

func (user User) GetLikedPostsWithTopic() ([]LikedPost, error) {
	var posts []LikedPost
	query := `
        SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at, t.topic AS thread_topic
        FROM post_likes pl
        JOIN posts p ON pl.post_id = p.id
        JOIN threads t ON p.thread_id = t.id
        WHERE pl.user_id = ?
        ORDER BY p.created_at DESC`
	rows, err := Db.Query(query, user.ID) // Use user.ID to filter likes by this user
	if err != nil {
		Logger.Printf("Failed to execute query to get user's liked posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post LikedPost
		var threadTopic string // Variable to hold the thread topic
		if err = rows.Scan(&post.ID, &post.UUID, &post.Body, &post.UserID, &post.ThreadID, &post.CreatedAt, &threadTopic); err != nil {
			Logger.Printf("Failed to scan liked post and thread topic from db. Error: %v", err)
			return nil, err
		}
		post.ThreadTopic = threadTopic // Assign the fetched thread topic to the LikedPost struct
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		Logger.Printf("Error occurred during rows iteration for liked posts: %v", err)
		return nil, err
	}

	Logger.Printf("Successfully finished loading liked posts with thread topics for user %v", user.Name)
	return posts, nil
}

func (user *User) CreateSession() (session Session, err error) {
	sess, err := user.Session() // query session
	if err == nil {             // reset UUID if session exists
		sess.DeleteByUUID()
	}
	statement := "insert into sessions (uuid, email, user_id, created_at) values ($1, $2, $3, $4) returning uuid, email, user_id, created_at"
	stmt, err := Db.Prepare(statement) // prepare the UUID
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(uuid.New().String(), user.Email, user.UUID, time.Now()).Scan(&session.UUID, &session.Email, &session.UserID, &session.CreatedAt)
	return
}

// Get the session for an existing user
func (user *User) Session() (session Session, err error) {
	session = Session{}
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at FROM sessions WHERE user_id = $1", user.UUID).
		Scan(&session.ID, &session.UUID, &session.Email, &session.UserID, &session.CreatedAt)
	Logger.Printf("State of Session data after user fetch: %v", session.UUID)
	return
}

// Reserve of useful but not used functions. R.I.P

// Gets posts by user in the database and returns them
func (user *User) GetPosts() ([]Post, error) {
	var posts []Post
	rows, err := Db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE user_id = $1 ORDER BY created_at DESC", user.ID) // Make sure to use user.ID if it's the actual foreign key
	if err != nil {
		Logger.Printf("Failed to execute query to get user's posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.ID, &post.UUID, &post.Body, &post.UserID, &post.ThreadID, &post.CreatedAt); err != nil {
			Logger.Printf("Failed to scan post from db. Error: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		Logger.Printf("Error occurred during rows iteration: %v", err)
		return nil, err
	}

	Logger.Printf("Successfully finished loading posts for user %v", user.Name)
	return posts, nil
}

func (user *User) GetLikedPosts() ([]LikedPost, error) {
	var posts []LikedPost
	query := `
			SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at
			FROM post_likes pl
			JOIN posts p ON pl.post_id = p.id
			WHERE pl.user_id = ?
			ORDER BY p.created_at DESC`
	rows, err := Db.Query(query, user.ID) // Use user.ID to filter likes by this user
	if err != nil {
		Logger.Printf("Failed to execute query to get user's liked posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post LikedPost
		if err = rows.Scan(&post.ID, &post.UUID, &post.Body, &post.UserID, &post.ThreadID, &post.CreatedAt); err != nil {
			Logger.Printf("Failed to scan liked post from db. Error: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		Logger.Printf("Error occurred during rows iteration for liked posts: %v", err)
		return nil, err
	}

	Logger.Printf("Successfully finished loading liked posts for user %v", user.Name)
	return posts, nil
}
