package internal

type SearchResponse struct {
	Threads []Thread `json:"threads,omitempty"`
	Posts   []Post   `json:"posts,omitempty"`
	Users   []User   `json:"users,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type NewThreadRequest struct {
	Category string `json:"category"`
	Topic    string `json:"topic"`
	Body     string `json:"body"`
	UserUUID string `json:"user_uuid"`
}

type DeleteThreadRequest struct {
	ThreadID string `json:"threadid"`
}

type NewPostRequest struct {
	ThreadID string `json:"threadid"`
	Body     string `json:"body"`
	UserUUID string `json:"useruuid"`
}

type DeletePostRequest struct {
	PostID string `json:"postid"`
}

type AddResponse struct {
	Exists bool   `json:"exists,omitempty"`
	Error  string `json:"error,omitempty"`
}

type CategoryRequest struct {
	CategoryName string `json:"categoryName"`
}

type RatingPayload struct {
	DataID   string `json:"id"`     // ID for thread or post
	DataType string `json:"type"`   // "thread" or "post"
	Action   string `json:"action"` // "like" or "dislike"
	UserID   int    `json:"userID"` // User's ID who liked or disliked
}

type AddViewRequest struct {
	ThreadID string `json:"threadid"`
	UserUUID string `json:"useruuid"`
}
