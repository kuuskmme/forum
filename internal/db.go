package internal

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func init() {
	DbTables := []string{
		`CREATE TABLE IF NOT EXISTS "users" (
			"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"	TEXT NOT NULL UNIQUE,
			"name"	TEXT NOT NULL UNIQUE,
			"email"	TEXT NOT NULL UNIQUE,
			"password"	TEXT NOT NULL,
			"created_at"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS "threads" (
			"id"            INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"          TEXT NOT NULL UNIQUE,
			"topic"         TEXT NOT NULL,
			"body"          TEXT NOT NULL,
			"user_id"       INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"created_at"    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"category_id"   INTEGER NOT NULL REFERENCES "categories"("id") ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS "posts" (
			"id"            INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"          TEXT NOT NULL UNIQUE,
			"body"          TEXT NOT NULL,
			"user_id"       INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"thread_id"     INTEGER NOT NULL REFERENCES "threads"("id") ON DELETE CASCADE,
			"created_at"    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS "thread_reactions" (
			"id"        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"      TEXT NOT NULL UNIQUE,
			"key"       INTEGER NOT NULL,
			"seen"      INTEGER NOT NULL DEFAULT 0,
			"user_id"   INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"thread_id" INTEGER NOT NULL REFERENCES "threads"("id") ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS "post_pictures" (
			"id"        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"      TEXT NOT NULL UNIQUE,
			"name"      TEXT NOT NULL,
			"path"      TEXT NOT NULL,
			"user_id"   INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"post_id"   INTEGER NOT NULL REFERENCES "posts"("id") ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS "categories" (
			"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"name"	TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS "sessions" (
			"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"uuid"	TEXT NOT NULL UNIQUE,
			"email"	TEXT NOT NULL UNIQUE,
			"user_id"	TEXT NOT NULL REFERENCES "users"("uuid") ON DELETE CASCADE,
			"created_at"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
		`CREATE TABLE IF NOT EXISTS "titles" (
			"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"user_id"	TEXT NOT NULL UNIQUE REFERENCES "users"("id") ON DELETE CASCADE,
			"key"	INTEGER NOT NULL DEFAULT 1,
			"title"	TEXT NOT NULL DEFAULT 'Pleb'
		)`,
		`CREATE TABLE IF NOT EXISTS "thread_views" (
			"user_id"   INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"thread_id"   INTEGER NOT NULL REFERENCES "threads"("id") ON DELETE CASCADE,
			PRIMARY KEY ("user_id", "thread_id")
		);`,
		`CREATE TABLE IF NOT EXISTS "thread_likes" (
			"id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"thread_id"   INTEGER NOT NULL REFERENCES "threads"("id") ON DELETE CASCADE,
			"user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"liked"       INTEGER NOT NULL DEFAULT 1,
			UNIQUE ("thread_id", "user_id")
		);`,
		`CREATE TABLE IF NOT EXISTS "thread_dislikes" (
			"id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"thread_id"   INTEGER NOT NULL REFERENCES "threads"("id") ON DELETE CASCADE,
			"user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"disliked"    INTEGER NOT NULL DEFAULT 1,
			UNIQUE ("thread_id", "user_id")
		);`,
		`CREATE TABLE IF NOT EXISTS "post_likes" (
			"id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"post_id"     INTEGER NOT NULL REFERENCES "posts"("id") ON DELETE CASCADE,
			"user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"liked"       INTEGER NOT NULL DEFAULT 1,
			UNIQUE ("post_id", "user_id")
		);`,
		`CREATE TABLE IF NOT EXISTS "post_dislikes" (
			"id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"post_id"     INTEGER NOT NULL REFERENCES "posts"("id") ON DELETE CASCADE,
			"user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
			"disliked"    INTEGER NOT NULL DEFAULT 1,
			UNIQUE ("post_id", "user_id")
		);`,
	}

	//Create database

	dbPath := os.Getenv("FORUM_DB_PATH") // using env vars for simplicity
	if dbPath == "" {
		log.Fatal("FORUM_DB_PATH environment variable is not set. Retry the bash script?")
	}

	var err error
	Db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		Logger.Println("ERROR | Unable to create forum.db")
		panic(err)
	}

	// Create each database table :
	for _, table := range DbTables {
		err := createDbTable(table)
		if err != nil {
			panic(err)
		}
	}
	_, err = Db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		Logger.Printf("Error after processing")
		panic(err)
	}
	log.Println("\033[1m\033[92mDATABASE | Database created and initialized, or found successfully.\033[0m")

}

func createDbTable(table string) error {
	statement, err := Db.Prepare(table)
	if err != nil {
		log.Println("ERROR | Unable to create tables in the database.")
		return err
	}

	defer statement.Close()

	statement.Exec()
	return nil
}
