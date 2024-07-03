module llforum

go 1.21.6

require (
	github.com/google/uuid v1.6.0 // Used for UUID functionality (UUID GENERATION FOR SESSION MANAGEMENT) - BONUS FUNCTIONALITY
	github.com/mattn/go-sqlite3 v1.14.22 // Used to interact with db.
	github.com/microcosm-cc/bluemonday v1.0.26 // used to sanitize inputs - BONUS FUNCTIONALITY
	golang.org/x/crypto v0.21.0 // Used to hash passwords before storage - BONUS FUNCTIONALITY
)

require ( // douceur, gorilla/css used by bluemonday as dependencies
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	golang.org/x/net v0.21.0 // indirect
)
