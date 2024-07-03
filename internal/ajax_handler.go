package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid" // Used for bonus UUID functionality
)

// User ID
// /get-user-id?
func GetUserID(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != "GET" {
		errMsg := ErrorMessages{HttpError: "Method not allowed", HttpCode: http.StatusMethodNotAllowed}
		generateErrorHTML(w, r, InitPage(&errMsg))
		Logger.Printf("Invalid method: %v", r.Method)
		return
	}

	// Get the UUID from query parameters
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		Logger.Printf("UUID invalid! Status: %v", http.StatusBadRequest)
		return
	}
	user := User{}
	// Yes, it would be "more correct" to use the session UUID in cookies to retrieve the session
	// And then we would lookup the user based on the user_id contained in the session and blablabla
	// Lookup the user ID based on the UUID
	err := user.GetByUUID(uuid)
	if err != nil {
		Logger.Printf("Error looking up userID for UUID %s: %v", uuid, err)
		generateErrorHTML(w, r, InitErrorPage("UserID lookup failed", http.StatusInternalServerError))
		return
	}

	// Prepare the response
	response := map[string]int{"userID": user.ID}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Printf("Error encoding response: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Response encoding failed.", http.StatusInternalServerError))
		return
	}
}

// Rating management
// /update-rating
func UpdateRating(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		Logger.Printf("Invalid method: %v", r.Method)
		return
	}

	var payload RatingPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	intDataID, err := strconv.Atoi(payload.DataID)
	if err != nil {
		Logger.Printf("Error converting ID to int: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Call a unified function to update the rating
	newRating, err := updateRating(intDataID, payload.DataType, payload.Action, payload.UserID)
	if err != nil {
		Logger.Printf("Error updating rating: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send back the new rating
	json.NewEncoder(w).Encode(map[string]int{"newRating": newRating})
}

// main sub-method for UpdateRating
func updateRating(dataID int, dataType, action string, userID int) (int, error) {
	tx, err := Db.Begin()
	if err != nil {
		return 0, err
	}

	// Determine table names based on dataType and action
	likeTable, dislikeTable, insertTable, removeTable, err := determineRatingType(dataType, action)
	if err != nil {
		Logger.Printf("Failed to determine update params %v", err)
	}
	// Remove opposite reaction if it exists
	_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s_id = ? AND user_id = ?", removeTable, dataType), dataID, userID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Attempt to insert the new reaction
	// First, check if a reaction already exists in the target table
	var existingID int
	err = tx.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE %s_id = ? AND user_id = ?", insertTable, dataType), dataID, userID).Scan(&existingID)
	if err == nil {
		// If the row exists, delete it
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", insertTable), existingID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	} else if err != sql.ErrNoRows {
		tx.Rollback()
		return 0, err
	} else {
		// If the row does not exist, insert the new reaction
		_, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (%s_id, user_id) VALUES (?, ?)", insertTable, dataType), dataID, userID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Count likes and dislikes
	var likes, dislikes int
	err = tx.QueryRow(fmt.Sprintf(`
        SELECT
            (SELECT COUNT(*) FROM %s WHERE %s_id = $1) AS likes,
            (SELECT COUNT(*) FROM %s WHERE %s_id = $1) AS dislikes`,
		likeTable, dataType, dislikeTable, dataType), dataID).Scan(&likes, &dislikes)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return likes - dislikes, nil
}

// sub-method for UpdateRating
func determineRatingType(dataType, action string) (string, string, string, string, error) {
	var likeTable, dislikeTable string
	switch dataType {
	case "thread":
		likeTable, dislikeTable = "thread_likes", "thread_dislikes"
	case "post":
		likeTable, dislikeTable = "post_likes", "post_dislikes"
	default:
		return "", "", "", "", fmt.Errorf("invalid dataType")
	}

	var insertTable, removeTable string
	// Determine which table to target based on the action
	switch action {
	case "like":
		insertTable, removeTable = likeTable, dislikeTable
	case "dislike":
		insertTable, removeTable = dislikeTable, likeTable
	default:
		return "", "", "", "", fmt.Errorf("invalid action")
	}
	Logger.Println("Rating successfully determined!")
	return likeTable, dislikeTable, insertTable, removeTable, nil
}

// SEARCH
// /search/posts
func SearchThreadsByPost(w http.ResponseWriter, r *http.Request) {
	Logger.Printf("SearchThreads called!")
	if r.Method != "GET" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		Logger.Printf("Invalid method: %v", r.Method)
	}
	var response SearchResponse
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) != 3 {
		Logger.Printf("Invalid method or path! len: %v, method: %v", len(pathSegments), r.Method)
		response.Errors = append(response.Errors, "Invalid method or path.")
	} else {
		threadsFiltered, err := getAllThreadsFilteredByPost(pathSegments[2])
		Logger.Printf("Threads filtered by user: %v", threadsFiltered)
		if err != nil {
			response.Errors = append(response.Errors, "No threads matching search criteria.")
			// If you still want to return all threads in case of an
		} else {
			response.Threads = threadsFiltered
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// /search/users
func SearchThreadsByUser(w http.ResponseWriter, r *http.Request) {
	Logger.Printf("SearchThreads called!")
	if r.Method != "GET" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		Logger.Printf("Invalid method: %v", r.Method)
	}
	var response SearchResponse
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) != 3 {
		Logger.Printf("Invalid method or path! len: %v, method: %v", len(pathSegments), r.Method)
		response.Errors = append(response.Errors, "Invalid method or path.")
	} else {
		threadsFiltered, err := getAllThreadsFilteredbyUser(pathSegments[2])
		Logger.Printf("Threads filtered by user: %v", threadsFiltered)
		if err != nil {
			response.Errors = append(response.Errors, "No threads matching search criteria.")
			// If you still want to return all threads in case of an
		} else {
			response.Threads = threadsFiltered
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// /search/threads
func SearchThreadsByThread(w http.ResponseWriter, r *http.Request) {
	Logger.Printf("SearchThreads called!")
	if r.Method != "GET" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		Logger.Printf("Invalid method: %v", r.Method)
	}
	var response SearchResponse
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) != 3 {
		Logger.Printf("Invalid method or path! len: %v, method: %v", len(pathSegments), r.Method)
		response.Errors = append(response.Errors, "Invalid method or path.")
	} else {
		threadsFiltered, err := getAllThreadsFilteredbyThread(pathSegments[2])
		Logger.Printf("Threads filtered by thread: %v", threadsFiltered)
		if err != nil {
			response.Errors = append(response.Errors, "No threads matching search criteria.")
			// If you still want to return all threads in case of an
		} else {
			response.Threads = threadsFiltered
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// Category AJAX
// action/new-category
func AddCategory(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Exists bool   `json:"exists"`
		Error  string `json:"error,omitempty"`
	}
	if r.Method == "POST" {
		var categoryReq CategoryRequest
		err := json.NewDecoder(r.Body).Decode(&categoryReq) // Decode the request body into the struct
		if err != nil {
			Logger.Printf("Bad request!")
			// Handle error: Bad request
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		Logger.Printf("Received category name: %s", categoryReq.CategoryName)

		err = addCategoryByName(categoryReq.CategoryName)
		Logger.Printf("Error: %v", err)
		if err != nil {
			response.Error = err.Error()
			response.Exists = true
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			Logger.Printf("Error encoding response: %v", err)
			generateErrorHTML(w, r, InitErrorPage("Encoding error", http.StatusInternalServerError))
			return
		}
	}
	if r.Method == "GET" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		Logger.Printf("Invalid method: %v", r.Method)
		response.Error = "Invalid method or path."
		return
	}
	// generateHTML(w, r, InitPage(r.Context().Value(UserCtxKey)), "new-category")
}

// action/remove-category
func RemoveCategory(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Exists bool   `json:"exists"`
		Error  string `json:"error,omitempty"`
	}
	if r.Method == "POST" {
		var categoryReq CategoryRequest
		err := json.NewDecoder(r.Body).Decode(&categoryReq) // Decode the request body into the struct
		if err != nil {
			Logger.Printf("Bad request!")
			// Handle error: Bad request
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		Logger.Printf("Received category name: %s", categoryReq.CategoryName)

		err = removeCategoryByName(categoryReq.CategoryName)
		Logger.Printf("Error: %v", err)
		if err != nil {
			response.Error = err.Error()
			response.Exists = true
		}

		// Prepare your response, for example, check if the category exists

		// Example check (replace with your actual category existence check)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			Logger.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
	if r.Method == "GET" {
		errMsg := ErrorMessages{HttpError: "Method not allowed", HttpCode: http.StatusMethodNotAllowed}
		generateErrorHTML(w, r, InitPage(&errMsg))
		Logger.Printf("Invalid method: %v", r.Method)
		response.Error = "Invalid method or path."
		return
	}
	// generateHTML(w, r, InitPage(r.Context().Value(UserCtxKey)), "new-category")
}

// /get-categories
func GetCategories(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Categories []Category `json:"categories"`
	}
	if r.Method == "GET" { // Correct method
		// Fetch categories
		categories, err := getAllCategories()
		if err != nil {
			Logger.Printf("Failed to fetch categories!")
			http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		}
		if categories != nil {
			response.Categories = *categories
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			Logger.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	} else {
		errMsg := ErrorMessages{HttpError: "Method not allowed", HttpCode: http.StatusMethodNotAllowed}
		generateErrorHTML(w, r, InitPage(&errMsg))
		return
	}
}

// Thread AJAX

// /thread/new-thread
func NewThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		return
	} else {
		var req NewThreadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			generateErrorHTML(w, r, InitErrorPage("Invalid request body", http.StatusBadRequest))
			return
		}
		user := User{}
		err := user.GetByUUID(req.UserUUID)
		if err != nil {
			Logger.Printf("Failed to fetch user ID")
			generateErrorHTML(w, r, InitErrorPage("Failed to fetch user ID", http.StatusInternalServerError))
			return
		}
		// Insert the new thread into the database
		_, err = Db.Exec(`
    		INSERT INTO threads (uuid, topic, body, created_at, user_id, category_id) 
    		VALUES (?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), req.Topic, req.Body, time.Now().Format(DateTime), user.ID, req.Category)
		if err != nil {
			generateErrorHTML(w, r, InitErrorPage("Failed to create new thread", http.StatusInternalServerError))
			return
		}
	}
	// Respond to the client indicating success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Thread created successfully"})
}

// /thread/delete-thread
func DeleteThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req DeleteThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		generateErrorHTML(w, r, InitErrorPage("Invalid request body", http.StatusBadRequest))
		return
	}

	// Delete the thread from the database
	numberifiedID, err := strconv.Atoi(req.ThreadID)
	if err != nil {
		Logger.Printf("Failed to convert string to int: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Invalid thread ID request", http.StatusExpectationFailed))
		return
	}
	_, err = Db.Exec(`DELETE FROM threads WHERE id = ?`, numberifiedID)
	if err != nil {
		Logger.Printf("Failed to delete thread: Err: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Failed to delete thread", http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Thread deleted successfully"})
}

// /thread/new-post
func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		return
	} else {
		var req NewPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			generateErrorHTML(w, r, InitErrorPage("Invalid request body", http.StatusBadRequest))
			return
		}
		user := User{}
		err := user.GetByUUID(req.UserUUID)
		if err != nil {
			Logger.Printf("Failed to fetch user ID")
			generateErrorHTML(w, r, InitErrorPage("Failed to fetch user ID", http.StatusInternalServerError))
			return
		}
		// Insert the new thread into the database
		Logger.Printf("What we got: %v", req)
		_, err = Db.Exec(`
    		INSERT INTO posts (uuid, body, created_at, user_id, thread_id) 
    		VALUES (?, ?, ?, ?, ?)`,
			uuid.New().String(), req.Body, time.Now().Format(DateTime), user.ID, req.ThreadID)
		if err != nil {
			Logger.Printf("Failed to create new post: Err: %v", err)
			generateErrorHTML(w, r, InitErrorPage("Failed to create new post", http.StatusInternalServerError))
			return
		}
	}
	// Respond to the client indicating success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
}

// /thread-delete-post
func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req DeletePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Logger.Printf("Invalid request body, %v\nState of req: %v", err, req)
		generateErrorHTML(w, r, InitErrorPage("Invalid request body", http.StatusBadRequest))
		return
	}

	// Delete the thread from the database
	numberifiedID, err := strconv.Atoi(req.PostID)
	if err != nil {
		Logger.Printf("Failed to convert string to int: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Invalid post ID request", http.StatusExpectationFailed))
		return
	}
	Logger.Printf("Pre-delete")
	_, err = Db.Exec("DELETE FROM posts WHERE id = ?", numberifiedID)
	if err != nil {
		Logger.Printf("Failed to delete post: Err: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Failed to delete post", http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Thread deleted successfully"})
}

// Thread AJAX View Count Management
// thread/update-views
func AddViewCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		generateErrorHTML(w, r, InitErrorPage("Method not allowed", http.StatusMethodNotAllowed))
		return
	}
	var req AddViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Logger.Printf("Invalid request body, %v\nState of req: %v", err, req)
		generateErrorHTML(w, r, InitErrorPage("Invalid request body", http.StatusBadRequest))
		return
	}
	user := User{}
	err := user.GetByUUID(req.UserUUID)
	if err != nil {
		Logger.Printf("Failed to fetch user ID")
		generateErrorHTML(w, r, InitErrorPage("Failed to fetch user ID", http.StatusInternalServerError))
		return
	}
	_, err = Db.Exec(`
		INSERT INTO thread_views (user_id, thread_id) VALUES (?, ?)`,
		user.ID, req.ThreadID)
	if err != nil {
		// This is to be expected as the user_id-thread_id combos are unique by default.
		Logger.Printf("Failed to add thread view: Err: %v", err)
		// DONT RETURN HERE SINCE THIS MAY FAIL AND IS EXPECTED TO FAIL IF USER REVISITS THREADS HE HAS ALREADY VIEWED!
	}
	var count int
	err = Db.QueryRow(`
    	SELECT COUNT(*) FROM thread_views WHERE thread_id = ?`,
		req.ThreadID).Scan(&count)
	if err != nil {
		Logger.Printf("Failed to count thread views: Err: %v", err)
		generateErrorHTML(w, r, InitErrorPage("Failed to count thread views", http.StatusInternalServerError))
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"newViews": count})
}
