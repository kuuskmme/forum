package internal

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// USER AUTH AND SESSIONS
// /register
func Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		generateHTML(w, r, InitPage(), "register")
		return
	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
		}
		user := User{}
		err := user.InternalLogin(w, r) // perform internal login on user
		if err != nil {
			return
		}
		r = user.ReloadUserIntoContext(r) // assuming user is already authenticated.
		Logger.Printf("Context after signin: %v", getUserCtx(r))
		user.RefreshCookies(w, r, "/home")
		// Implement custom route redirection here
		return

	default:
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Method not allowed!: %v", r.Method), http.StatusMethodNotAllowed))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// /logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Check that we're receiving a POST request
	if r.Method != "POST" {
		Logger.Printf("Method not allowed!")
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Method not allowed!: %v", r.Method), http.StatusMethodNotAllowed))
		return
	}

	// Delete the session cookie by setting its max age to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	// Optionally, perform other cleanup.

	// Redirect the user to the home page or login page
	http.Redirect(w, r, "/home", http.StatusContinue) // this is used to clear the context data. Yes. Double redirect.. sigh.
	generateHTML(w, r, InitPage(), "home")
}

// PAGE HANDLERS
// /home or /
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Method not allowed!: %v", r.Method), http.StatusMethodNotAllowed))
	}
	Logger.Println("Home handler called")
	generateHTML(w, r, InitPage(getUserCtx(r)), "home")
}

// /forum
func Forum(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	categories, err := getAllCategories() // since categories are by default always included under /forum href
	if err != nil {
		generateErrorHTML(w, r, InitErrorPage(("Failed to load categories!"), http.StatusInternalServerError))
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}
	switch len(pathSegments) {
	case 1: // /forum/
		generateHTML(w, r, InitPage(getUserCtx(r), categories), "forum")
		return
	case 2: // /forum/thread_id/
		threadPath, err := strconv.Atoi(pathSegments[1])
		if err != nil { // invalid pathspec in terms of datatype
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Failed to load thread from ID: %v", threadPath), http.StatusInternalServerError))
			generateHTML(w, r, InitPage(getUserCtx(r)), "home")
			return
		}
		getThreads(w, r, threadPath)
		return
	default: // In practice, this should never be encountered
		// Handle default case or error
		Logger.Printf("Page not found!")
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
		return
	}

}

// /thread/ID
func ThreadH(w http.ResponseWriter, r *http.Request) {
	// Separate handler since we need to handle views, reactions and pictures associated with thread posts inside here.. probably
	Logger.Printf("Path: %v", string(r.URL.Path))
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	categories, err := getAllCategories() // since categories are by default always included under /forum href
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}
	switch len(pathSegments) {
	case 1: // /thread
		generateHTML(w, r, InitPage(getUserCtx(r), categories), "forum")
		return
	// Empty threads without assigned ids should redirect to the forum homepage
	// We could also make this route to creating a new thread?
	case 2: // /thread/thread_id
		threadID, err := strconv.Atoi(pathSegments[1])
		if err != nil { // invalid pathspec
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
			return
		}
		posts, err := getPostsAndAuthorsByThreadIDAscendingWithRating(threadID)
		// Posts by ascending created_at time
		if err != nil {
			Logger.Printf("Failed to load posts from ID: %v", threadID)
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Failed to load posts from ID: %v", threadID), http.StatusInternalServerError))
			return
		}
		thread, err := getDisplayThreadByIDWithRating(threadID)
		// There is only one thread so no sorting here
		if err != nil {
			Logger.Printf("Failed to load thread from ID: %v", threadID)
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Failed to load thread from ID: %v", threadID), http.StatusInternalServerError))
			return
		}
		/*user, err := RetrieveUser(r)
		if err != nil {
			Logger.Printf("Failed to retrieve user from thread ID: %v\nError:", threadID, err)
			return
		}*/
		generateHTML(w, r, InitPage(getUserCtx(r), categories, posts, thread), "thread")
		return
	case 3: // /thread/thread_id/post_by_id // We should maybe do this with javascript
		// This way we could do highlighting and maybe integrate the reactions and pics more smoothly
		threadID, err := strconv.Atoi(pathSegments[1])
		if err != nil { // invalid pathspec
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
			return
		}
		posts, err := getPostsAndAuthorsByThreadIDAscendingWithRating(threadID)
		// Posts by ascending created_at time
		if err != nil {
			Logger.Printf("Failed to load posts from ID: %v", threadID)
			generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Failed to load posts from ID: %v", threadID), http.StatusInternalServerError))
			return
		}
		thread, err := getDisplayThreadByIDWithRating(threadID)
		// There is only one thread so no sorting here
		if err != nil {
			Logger.Printf("Failed to load thread from ID: %v", threadID)
			http.Error(w, "Failed to load thread for specified ID", http.StatusInternalServerError)
			return
		}
		generateHTML(w, r, InitPage(getUserCtx(r), categories, posts, thread), "thread")
		return
	default: // In practice, this should never be encountered
		// Handle default case or error
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
		return
	}
}

// /user/view-profile/
func Profile(w http.ResponseWriter, r *http.Request) {
	//pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// intuitive way for user to change password, email
	Logger.Printf("Path: %v", string(r.URL.Path))
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	userData := User{}
	if len(pathSegments) == 3 { // make sure the path segment is valid
		// Retrieve the user via link source
		if err := userData.GetByUUID(pathSegments[2]); err == nil {
			// Query and populate userData with posts and threads
			posts, threads, likedPosts, err := userData.populateProfile()
			if err != nil {
				Logger.Printf("Failed to populate user profile")
			}
			ctx := getUserCtx(r)
			if r.Method == "GET" {
				generateHTML(w, r, InitPage(ctx, userData, &posts, &threads, &likedPosts), "profile")
				return
			}
			// User not found, serious issue encountered. Should never be encountered as /user/profile is protected path
		} else {
			Logger.Printf("Invalid User UUID")
			http.Redirect(w, r, "/oops", 500)
		}
	}
	// User not found with matching UUID
	Logger.Printf("Path segments not of length 3")
	generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
}

// /user/edit-profile/
func Settings(w http.ResponseWriter, r *http.Request) {
	Logger.Printf("Path: %v", r.URL.Path)
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) != 3 {
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Page not found: %v", r.URL.Path), http.StatusNotFound))
		return
	}

	userUUID := pathSegments[2]
	userData := User{}
	ctx := getUserCtx(r)
	if ctx == nil {
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Unauthorized access! Log in first. %v", r.URL.Path), http.StatusForbidden))
		return
	}
	err := userData.GetByUUID(userUUID)
	if err != nil {
		Logger.Printf("Failed to retrieve user: %v", err)
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("User not found: %v", r.URL.Path), http.StatusNotFound))
		return
	}
	posts, threads, likedPosts, err := userData.populateProfile()
	if err != nil {
		Logger.Printf("Failed to populate user profile: %v", err)
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Error encountered when fetching profile data: %v", r.URL.Path), http.StatusInternalServerError))
		return
	}
	switch r.Method {
	case "GET": // Just load the damn page CJ
		generateHTML(w, r, InitPage(getUserCtx(r), &userData, &posts, &threads, &likedPosts), "settings")
	case "POST":
		if err := r.ParseForm(); err != nil {
			Logger.Printf("Failed to parse form: %v", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		userData.handleProfileUpdate(w, r)
		Logger.Printf("User profile updated successfully.")
		generateHTML(w, r, InitPage(getUserCtx(r), &userData, &posts, &threads, &likedPosts), "settings")
	default:
		generateErrorHTML(w, r, InitErrorPage(fmt.Sprintf("Method not allowed!: %v", r.Method), http.StatusMethodNotAllowed))
		return
	}
}

// /search
func Search(w http.ResponseWriter, r *http.Request) {
	var threads []Thread
	err := getAllThreadsSorted(&threads)
	if err != nil {
		Logger.Printf("Failed to retrieve any threads!")
	}
	generateHTML(w, r, InitPage(&threads, getUserCtx(r)), "search")
}

// Mermaid ERD
// view-erd
func ERD(w http.ResponseWriter, r *http.Request) {
	generateERD(w, r, struct{}{})
}

// ERROR HANDLERS
// /oops
func Oops(w http.ResponseWriter, r *http.Request) { // Handler for HTTP 500, HTTP 403 etc
	Logger.Println("400/500 error handler called")
	tmpl, err := template.ParseFiles(TemplatePath + "oops.gohtml") // Adjust the path accordingly
	if err != nil {
		Logger.Printf("Error parsing oops template: %v", err)
		http.Error(w, "Error loading error page", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, InitErrorPage("", 0))
	if err != nil {
		Logger.Printf("Error executing oops template: %v", err)
		// Directly write to response as a last resort
		http.Error(w, "An error occurred", http.StatusInternalServerError)
	}
}
