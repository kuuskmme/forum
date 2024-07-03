package internal

// Thread represents a forum thread
type Thread struct {
	ID         int      `db:"id"`
	UUID       string   `db:"uuid"`
	Topic      string   `db:"topic"`
	Body       string   `db:"body"`
	UserID     int      `db:"user_id"`
	CreatedAt  string   `db:"created_at"`
	CategoryID int      `db:"category_id"`
	Rating     int      // Ratio of likes to dislikes populated when loading in threads
	Author     User     // This is a field only used when loading in users for threads
	Category   Category // This is a field populated via category_id field as go templates do not support fetching
	Views      int      // This is a field populated via thread_id field as a means to convey views accrued in thread.gohtml
}

// Post represents a post in a thread
type Post struct {
	ID          int    `db:"id"`
	UUID        string `db:"uuid"`
	Body        string `db:"body"`
	Approved    int    `db:"approved"`
	UserID      int    `db:"user_id"`
	ThreadID    int    `db:"thread_id"`
	CreatedAt   string `db:"created_at"`
	PostAuthor  User   // This is a field used only when loading in users for threadview
	ThreadTopic string // Associated threads' title. Not the best implementation but it works.
	Rating      int    // Ratio of likes to dislikes (likes-dislike) used while fetching posts
}

type LikedPost struct { // this more or less is a subset of Post but we dont have inheritance and our PageData assembler needs a different type..sooooyeah.
	ID          int    `db:"id"`
	UUID        string `db:"uuid"`
	Body        string `db:"body"`
	Approved    int    `db:"approved"`
	UserID      int    `db:"user_id"`
	ThreadID    int    `db:"thread_id"`
	CreatedAt   string `db:"created_at"`
	PostAuthor  User   // This is a field used only when loading in users for threadview
	ThreadTopic string // Associated threads' title. Not the best implementation but it works.
	Rating      int    // Ratio of likes to dislikes (likes-dislike) used while fetching posts
}

// ThreadReaction represents a reaction to a thread
type ThreadReaction struct {
	ID       int    `db:"id"`
	UUID     string `db:"uuid"`
	Key      int    `db:"key"`
	Seen     int    `db:"seen"`
	UserID   int    `db:"user_id"`
	ThreadID int    `db:"thread_id"`
}

// PostReaction represents a reaction to a post
type PostReaction struct {
	ID     int    `db:"id"`
	UUID   string `db:"uuid"`
	Key    int    `db:"key"`
	Seen   int    `db:"seen"`
	UserID int    `db:"user_id"`
	PostID int    `db:"post_id"`
}

// PostPicture represents a picture attached to a post
type PostPicture struct {
	ID     int    `db:"id"`
	UUID   string `db:"uuid"`
	Name   string `db:"name"`
	Path   string `db:"path"`
	UserID int    `db:"user_id"`
	PostID int    `db:"post_id"`
}

// Category represents a forum category
type Category struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// User represents a user in the system
type User struct {
	ID        int    `db:"id"`
	UUID      string `db:"uuid"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	CreatedAt string `db:"created_at"`
	Provider  string // Used only for 0auth routes
}

// Session represents a user session
type Session struct {
	ID        int    `db:"id"`
	UUID      string `db:"uuid"`
	Email     string `db:"email"`
	UserID    string `db:"user_id"`
	CreatedAt string `db:"created_at"`
}

// PostView represents a view of a post by a user
type PostView struct {
	UserID int `db:"user_id"`
	PostID int `db:"post_id"`
}

type ErrorMessages struct {
	SignUpErrors    []string
	LoginErrors     []string
	ProfileMessages []string
	SearchMessages  []string
	HttpError       string
	HttpCode        int
}

type PageData struct { // Implemented for legibility
	User           User
	PageUsers      []User
	UserCtx        UserCtx
	Errors         ErrorMessages
	Categories     []Category
	Category       Category
	Threads        []Thread
	Thread         Thread
	Posts          []Post
	UserLikedPosts []LikedPost
	Post           Post
}

// This constructs our page data by dereferencing pointers and identifying types
func InitPage(dataInstances ...interface{}) PageData {
	pageData := PageData{}
	// Init pagedata

	for _, instance := range dataInstances {
		switch v := instance.(type) {
		// CASES NEED TO BE DEFINED FOR ADDITIONAL PAGEDATA FIELDS IF/WHEN NECESSARY
		case *User:
			pageData.User = *v // dereference the pointer so our template gets the ctx data
		case User:
			pageData.User = v // dereference the pointer so our template gets the ctx data
		case UserCtx:
			pageData.UserCtx = v // dereference the pointer so our template gets the ctx data
		case []User:
			pageData.PageUsers = v
		case *[]User:
			pageData.PageUsers = *v
		case *ErrorMessages:
			pageData.Errors = *v
		case *[]Category:
			pageData.Categories = *v
		case *Category:
			pageData.Category = *v
		case *[]Thread:
			pageData.Threads = *v
		case *Thread:
			pageData.Thread = *v
		case *[]Post:
			pageData.Posts = *v
		case *[]LikedPost:
			pageData.UserLikedPosts = *v
		case *Post:
			pageData.Post = *v
		}
	}

	return pageData
}

func InitErrorPage(msg string, code int) HttpError {
	return HttpError{Code: code, Msg: msg}
}
