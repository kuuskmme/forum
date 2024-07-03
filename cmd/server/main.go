package main

import (
	"fmt"
	"html/template"
	"llforum/internal" // Adjust this import path based on your module's actual path and structure
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Tmpl *template.Template
}

// Validate session UUID, reload context when necessary
func checkContext(sessionCookie *http.Cookie, r *http.Request, err error) (*http.Request, error) {
	if err == nil && internal.ValidateSessionFromDB(sessionCookie.Value) { // if session is valid
		internal.Logger.Printf("\033[1m\033[92mSession valid. User login detected.\033[0m")
		if user, ok := r.Context().Value(internal.UserCtxKey).(*internal.Session); ok {
			// Successfully retrieved the user from the context
			internal.Logger.Printf("User retrieved from context within: %+v", user)
			return r, nil // User already in context, skip this function
		} else { // only if the user does not exist in the context do we reload into the context
			// Failed to retrieve the user from the context
			internal.Logger.Printf("User not found in context, preparing context...")
			userSession := internal.Session{UUID: sessionCookie.Value} // Define session object and set UUID
			userSession.GetUserIDBySessionUUID()                       // this should fetch the necessary user id
			user, err := userSession.LoadUserIntoContext()             // load user with only context-req data
			if err != nil {
				internal.Logger.Printf("Failed to fetch UserContext via UserContextOnly(): %v", err)
				return r, fmt.Errorf("usercontext fetch failed") // queue internal error
			}
			r = user.ReloadUserIntoContext(r) // reload user data into request context and log
			return r, nil                     // User loaded in context.
		}
	}
	return r, fmt.Errorf("failed to load http.Request with context")
}
func loggingAndSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				internal.Logger.Printf("Recovered from panic: %v", err)
				// Log the error and redirect to custom error page
				http.Redirect(w, r, "/oops", http.StatusInternalServerError)
			}
		}()
		internal.Logger.Printf("\033[1m\033[33mReceived %s request for %s from %s\033[0m", r.Method, r.URL.Path, r.RemoteAddr)

		// Static assets are public by design
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		// Include all user-specific routes in this gr8 map
		protectedPaths := map[string]bool{
			// Add other non-AJAX handler paths that require the user to be logged in
		}
		// Attempt to get the session cookie
		sessionCookie, err := internal.GetSessionCookie(r)
		// Verify context validity assuming we have a sessionCookie
		r, err = checkContext(sessionCookie, r, err)

		// If it's a protected path and the user is not logged in, redirect to login page. I am not sure this is necessary as permissions logic is not required in the project brief.
		if protectedPaths[r.URL.Path] && err != nil {
			internal.Logger.Printf("Protected route access attempt without auth: Redirecting user..")
			http.Redirect(w, r, "/register", http.StatusForbidden)
			// A good place to do context-less redirect!
			return
		}
		lrw := internal.NewLoggingResponseWriter(w)

		// Immediately handle error status codes.
		if statusCode := lrw.StatusCode(); statusCode >= 400 && statusCode < 600 {
			// Log the occurrence of an error status code
			internal.Logger.Printf("Error status code: %d", statusCode)

			// Load and render your oops.gohtml template.
			tmpl, err := template.ParseFiles(internal.TemplatePath + "oops.gohtml")
			if err != nil {
				internal.Logger.Printf("Error loading template: %v", err)
				return // Returning since you can't recover from this error in this context.
			}

			// Clear response writer to prepare for sending error page.
			lrw.Header().Set("Content-Type", "text/html; charset=utf-8")

			// Execute the template directly to the LoggingResponseWriter.
			err = tmpl.Execute(lrw, nil) // Pass any necessary data for the template.
			if err != nil {
				internal.Logger.Printf("Error executing template: %v", err)
				return
			}

			lrw.WriteHeader(statusCode) // Re-apply the original status code.
			lrw.Flush()                 // Flush the custom error page content.
			return
		}
		next.ServeHTTP(w, r) // Execute the handler chain
		// Log the status code for successful requests
		internal.Logger.Printf("Response status code: %d", lrw.StatusCode())

	})
}

func main() {
	internal.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	mux := http.NewServeMux()

	// Utilizing the handlers from the internal package
	//mux.HandleFunc("/", internal.Home)             // base redirect
	mux.HandleFunc("/register", internal.Register) // Register/Login
	mux.HandleFunc("/", internal.Home)             // placeholder homepage
	// Forum View
	mux.HandleFunc("/get-categories", internal.GetCategories)
	mux.HandleFunc("/new-category", internal.AddCategory)
	mux.HandleFunc("/remove-category", internal.RemoveCategory)
	mux.HandleFunc("/forum", internal.Forum)
	mux.HandleFunc("/forum/", internal.Forum) // Catch subpaths
	// Search View
	mux.HandleFunc("/search", internal.Search)
	mux.HandleFunc("/search/posts/", internal.SearchThreadsByPost)
	mux.HandleFunc("/search/threads/", internal.SearchThreadsByThread)
	mux.HandleFunc("/search/users/", internal.SearchThreadsByUser)
	// Thread View
	mux.HandleFunc("/thread", internal.ThreadH) //
	mux.HandleFunc("/thread/new-thread", internal.NewThread)
	mux.HandleFunc("/thread/delete-thread", internal.DeleteThread)
	mux.HandleFunc("/thread/new-post", internal.NewPost)
	mux.HandleFunc("/thread/delete-post", internal.DeletePost)
	mux.HandleFunc("/thread/update-views", internal.AddViewCount)
	mux.HandleFunc("/thread/", internal.ThreadH) // Catch subpaths
	// User routes
	mux.HandleFunc("/logout", internal.Logout) // Logout
	mux.HandleFunc("/user/edit-profile/", internal.Settings)
	mux.HandleFunc("/user/view-profile/", internal.Profile)
	// Error routes
	mux.HandleFunc("/oops", internal.Oops) // Error 500, 5XX
	// 0auth routes
	mux.HandleFunc("/auth/google", internal.GoogleAuth)             // Initial auth GET
	mux.HandleFunc("/auth/google/callback", internal.AuthCallback)  // Callback handler
	mux.HandleFunc("/auth/github", internal.GithubAuth)             // Initial auth GET
	mux.HandleFunc("/auth/github/callback", internal.AuthCallback)  // Callback handler
	mux.HandleFunc("/auth/discord", internal.DiscordAuth)           // Initial auth GET
	mux.HandleFunc("/auth/discord/callback", internal.AuthCallback) // Callback handler
	// AJAX routes
	mux.HandleFunc("/get-user-id", internal.GetUserID)
	mux.HandleFunc("/update-rating", internal.UpdateRating)

	// View ERD
	mux.HandleFunc("/view-erd", internal.ERD)
	//mux.HandleFunc("/", internal.Home)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static")) // Adjusted the path to "./static" assuming your project root is the working directory
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Wrap the mux with the logging and session middleware
	enhancedMux := loggingAndSessionMiddleware(mux)

	// Start the server with the enhanced mux
	internal.Logger.Printf("\033[1m\033[92mListening on port %s\n\033[0m", internal.Port)
	if err := http.ListenAndServe(":"+internal.Port, enhancedMux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
