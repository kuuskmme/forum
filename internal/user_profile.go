package internal

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (user *User) populateProfile() ([]Post, []Thread, []LikedPost, error) {
	posts, err := user.GetPostsWithTopic()
	if err != nil {
		Logger.Printf("Failed to retrieve user %v posts, Error: %v", user.Name, err)
	}
	threads, err := user.GetThreads()
	if err != nil {
		Logger.Printf("Failed to retrieve user %v threads, Error: %v", user.Name, err)
	}
	likedPosts, err := user.GetLikedPostsWithTopic()
	if err != nil {
		Logger.Printf("Failed to retrieve users liked %v posts, Error: %v", user.Name, err)
	}
	Logger.Printf("User %v threads: %v posts: %v", user.Name, threads, posts)
	return posts, threads, likedPosts, nil
}

func (user *User) handleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	errors := ErrorMessages{}
	user.validateUserProfileUpdate(r, &errors)

	if len(errors.ProfileMessages) > 0 {
		Logger.Printf("User profile update validation failed.")
		pageData := InitPage(r.Context().Value(UserCtxKey), user, errors)
		generateHTML(w, r, pageData, "settings")
		return
	}

	if err := user.UpdateUserProfile(); err != nil {
		Logger.Printf("Failed to update user profile: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
}

// Change password and email for user
func (user *User) UpdateUserProfile() (err error) {
	stmt, err := Db.Prepare("UPDATE users SET email = ?, password = ? WHERE UUID = ?")
	if err != nil {
		Logger.Printf("ERROR | Could not prepare users table update for email and password: %v", err)
		return err
	}
	_, err = stmt.Exec(user.Email, user.Password, user.UUID)
	if err != nil {
		Logger.Printf("ERROR | Could not overwrite fields for user %v: %v", user, err)
		return err
	}
	Logger.Printf("User profile updated successfully!")
	return nil
}

func (user *User) validateUserProfileUpdate(r *http.Request, errors *ErrorMessages) {
	email := sanitizeInput(r.FormValue("email"))
	password := r.FormValue("password")
	var emailChange, passwordChange bool
	if r.FormValue("emailChangeRequested") == "true" {
		err := user.Check()
		if !validEmail(email) {
			errors.ProfileMessages = append(errors.ProfileMessages, "New email is invalid!")
		}
		if err != nil {
			errors.ProfileMessages = append(errors.ProfileMessages, "Email or name already exists!")
		}
		if len(errors.ProfileMessages) == 0 {
			emailChange = true
		}
		// Assuming email is valid
	}
	if r.FormValue("passwordChangeRequested") == "true" {
		if !validatePassword(password) {
			errors.ProfileMessages = append(errors.ProfileMessages, "Password must contain numbers, and unicode uppercase/lowercase letters!")
		} else {
			passwordChange = true
		}
	}
	if len(errors.ProfileMessages) == 0 {
		if emailChange {
			user.Email = email
		}
		if passwordChange {
			if hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
				Logger.Printf("ERROR | Could not hash password: %v", err)
				// 5-7382 = hashing failed, internal err
				errors.ProfileMessages = append(errors.ProfileMessages, "Error 5-7382")
			} else {
				user.Password = string(hashedPassword)
			}
		}
	}
	Logger.Printf("State of profilemessages: %v", errors.ProfileMessages)
}
