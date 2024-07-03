package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func (user *User) InternalLogin(w http.ResponseWriter, r *http.Request) error {
	errors := ErrorMessages{} // create the errormessages and user structs pre-collection
	action := r.FormValue("action")
	switch action {
	case "signup":
		user.VerifyRegistration(w, r, &errors)
	case "signin":
		user.Authenticate(r, &errors)
	}
	if len(errors.SignUpErrors) != 0 || len(errors.LoginErrors) != 0 {
		// If errors are encountered during registration or login, refresh the page with errors.
		Logger.Printf("Errors during register/login: %v", errors)
		generateHTML(w, r, InitPage(&errors), "register")
		return fmt.Errorf("error during signup or registration")
	}
	return nil
}

func (user *User) Authenticate(r *http.Request, errors *ErrorMessages) {
	user.Email, user.Password = sanitizeInput(r.FormValue("loginemail")), sanitizeInput(r.FormValue("loginpassword"))
	// Query the database for the user by email
	// We only need the password and email to auth, DB query simplified.
	// We also need created_at and the name for context data eval later, so might as well grab em
	// User id is needed for our session logic as it is a FK
	err := Db.QueryRow("SELECT uuid, name, password, created_at FROM users WHERE email = $1", user.Email).Scan(&user.UUID, &user.Name, &user.Password, &user.CreatedAt)
	if err != nil {
		// Log the detailed error for internal debugging
		Logger.Printf("Failed to find user with email %s: %v", user.Email, err)

		// Return a generic error message without revealing the nature of the
		errors.LoginErrors = append(errors.LoginErrors, "Login invalid!")
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.FormValue("loginpassword")))
	if err != nil {
		// Log the detailed bcrypt comparison failure for internal debugging
		Logger.Printf("Password does not match for user %s", user.Email)

		// Return a generic error message to avoid revealing specifics about the failure
		errors.LoginErrors = append(errors.LoginErrors, "Login invalid!")
		return
	}
	// Successful authentication
}

func (user *User) VerifyRegistration(w http.ResponseWriter, r *http.Request, errors *ErrorMessages) {
	name, email, password := r.FormValue("name"), r.FormValue("email"), r.FormValue("password")
	user.Name, user.Email, user.Password = sanitizeInput(name), sanitizeInput(email), password // set the proper values for the user struct

	if !validatePassword(password) {
		errors.SignUpErrors = append(errors.SignUpErrors, "Password not secure!")
	}
	if !validEmail(email) {
		errors.SignUpErrors = append(errors.SignUpErrors, "Email is invalid!")
	}
	err := user.Check()
	if err != nil {
		errors.SignUpErrors = append(errors.SignUpErrors, "Email or name already exists!")
	}
	if len(errors.SignUpErrors) == 0 { // where the user is created and added to DB.
		err := user.Create()
		if err != nil { // If we had any issues, just return /w errors
			Logger.Printf("Failed to create user: %v", user)
			generateHTML(w, r, InitPage(&errors), "register")
			return
		}
	} else {
		return
	}
}

// Simplified user auth for external auth routes
func (user *User) AuthenticateExternal(r *http.Request, errors ErrorMessages) {
	// Query the database for the user by email
	// We only need the password and email to auth, DB query simplified.
	// We also need created_at and the name for context data eval later, so might as well grab em
	// User id is needed for our session logic as it is a FK
	err := Db.QueryRow("SELECT uuid, name, password, created_at FROM users WHERE email = $1", user.Email).Scan(&user.UUID, &user.Name, &user.Password, &user.CreatedAt)
	if err != nil {
		// Log the detailed error for internal debugging
		Logger.Printf("Failed to find user with email %s: %v", user.Email, err)

		// Return a generic error message without revealing the nature of the
		errors.LoginErrors = append(errors.LoginErrors, "Login invalid!")
		return
	}
	// Successful authentication
}

func validatePassword(s string) bool {
	var number, upper, lower bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case unicode.IsLetter(c) || c == ' ':
		default:
			return false
		}
	}
	return number == upper == lower
}

func validEmail(email string) bool {
	// This regular expression is a simple version and might not cover all edge cases.
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Auth methods related to 0auth

func (user *User) GetGoogleProfile(token string) (err error) {
	parameters := url.Values{}
	parameters.Add("access_token", token)

	resp, err := http.Get(googleInfo + "?" + parameters.Encode())
	if err != nil {
		Logger.Printf("can't make the GET request for google: %s \n", err)
		return
	}
	defer resp.Body.Close()

	data, err := toJSON(resp)
	if err != nil {
		Logger.Printf("can't marshal google to JSON: %s \n", err)
		return
	}
	user.Name, user.Email, user.Provider = data["name"].(string), data["email"].(string), "google"
	user.Password = "tokenpassword"
	// We should add a separate field to the users table which denounces the origin of the user account (internal vs google vs github etc)
	// But thats extra work though...

	return nil
}

func (user *User) GetGitProfile(token string) (err error) {

	data, err := getGitUsername(gitUser, token)
	if err != nil {
		Logger.Printf("can't get the username from github: %s \n", err)
		return
	}

	email, err := getGitEmail(gitEmail, token)
	if err != nil {
		Logger.Printf("can't get the email from github: %s \n", err)
		return
	}

	user.Name, user.Email, user.Provider = data["login"].(string), email, "github"
	user.Password = "tokenpassword"

	return nil
}

func (user *User) GetDiscordData(token string) error {
	req, err := http.NewRequest("GET", discordUserAPI, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch user data, status code: %d", resp.StatusCode)
	}

	var userData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return err
	}
	user.Name, user.Email = userData.Username, userData.Email
	return nil
}

func GetToken(code, provider string) (token string, err error) {
	parameters := url.Values{}

	if provider == "google" {
		parameters.Add("client_id", googleClientID)
		parameters.Add("client_secret", googleClientSecret)
		parameters.Add("redirect_uri", googleRedirectURI)
		parameters.Add("grant_type", "authorization_code")
		parameters.Add("code", code)

		resp, err := http.Post(googleToken, format, strings.NewReader(parameters.Encode()))
		if err != nil {
			Logger.Printf("Failed to post to provider, %v", err)
			return "", err
		}
		defer resp.Body.Close()

		mappedBody, err := toJSON(resp)
		if err != nil {
			Logger.Printf("Failed to marshal to JSON, %v", err)
			return "", err
		}

		token = mappedBody["access_token"].(string)
		return token, nil
	}

	if provider == "github" {
		parameters.Add("client_id", githubClientID)
		parameters.Add("client_secret", githubClientSecret)
		parameters.Add("redirect_uri", githubRedirectURI)
		parameters.Add("code", code)

		resp, err := http.Post(githubToken, format, strings.NewReader(parameters.Encode()))
		if err != nil {
			Logger.Printf("Failed to post to provider, %v", err)
			return "", err
		}
		defer resp.Body.Close()

		byted, err := io.ReadAll(resp.Body)
		if err != nil {
			Logger.Printf("Failed to read into byteform, %v", err)
			return "", err
		}
		parts := strings.Split(string(byted), "&")
		for _, part := range parts {
			param := strings.Split(part, "=")
			if len(param) == 2 && param[0] == "access_token" {
				token = param[1]
				break
			}
		}

		if token == "" {
			Logger.Printf("Auth token not provided in response, %v", err)
			return "", errors.New("no access token in response")
		}

		return token, nil
	}

	if provider == "discord" {
		parameters.Add("client_id", discordClientID)
		parameters.Add("client_secret", discordClientSecret)
		parameters.Add("code", code)
		parameters.Add("grant_type", "authorization_code")
		parameters.Add("redirect_uri", discordRedirectURI)

		resp, err := http.PostForm(discordTokenURL, parameters)
		if err != nil {
			Logger.Printf("Failed to post to Discord provider: %v", err)
			return "", err
		}
		defer resp.Body.Close()

		var response struct {
			AccessToken string `json:"access_token"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			Logger.Printf("Failed to decode JSON response: %v", err)
			return "", err
		}

		if response.AccessToken == "" {
			Logger.Printf("Auth token not provided in response")
			return "", errors.New("no access token in response")
		}

		return response.AccessToken, nil
	}
	return "", errors.New("no provider detected")
}

func gitRequest(url, token string, target interface{}) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logger.Printf("Error performing GET request when querying GitHub")
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, target); err != nil {
		return err
	}

	return nil
}

// get email
func getGitUsername(url, token string) (map[string]interface{}, error) {
	var data map[string]interface{}

	if err := gitRequest(url, token, &data); err != nil {
		return nil, errors.New("gitrequest failed for user")
	}

	return data, nil
}

// get username
func getGitEmail(url, token string) (string, error) {
	var data []map[string]interface{}

	if err := gitRequest(url, token, &data); err != nil {
		return "", errors.New("gitrequest not working for email")
	}

	var email string
	for _, params := range data {
		if params["primary"].(bool) {
			email = params["email"].(string)
			break
		}
	}

	return email, nil
}
