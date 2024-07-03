package internal

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	googleToken        = "https://oauth2.googleapis.com/token"
	googleClientID     = "878005344262-6mmgs4uiuufc6rg3hjlumv90b80vv4eq.apps.googleusercontent.com"
	googleClientSecret = "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"
	googleRedirectURI  = "http://localhost:8080/auth/google/callback"
	googleInfo         = "https://www.googleapis.com/oauth2/v1/userinfo"
	googleAuth         = "https://accounts.google.com/o/oauth2/auth"

	githubToken        = "https://github.com/login/oauth/access_token"
	githubClientID     = "35e721a0a29e3ef9abd7"
	githubClientSecret = "4fd6b9e323648a1e85da82d04729839b8dd6fdd5"
	githubRedirectURI  = "http://localhost:8080/auth/github/callback"

	discordTokenURL     = "https://discord.com/api/oauth2/token"
	discordAuthURL      = "https://discord.com/api/oauth2/authorize"
	discordClientID     = "1229038247150223411"
	discordClientSecret = "exsq0FQ9O3BwUPDzATAnmqXeSTO0klDY"
	discordRedirectURI  = "http://localhost:8080/auth/discord/callback"
	discordUserAPI      = "https://discord.com/api/users/@me"

	discord0auth = "https://discord.com/oauth2/authorize?client_id=1229038247150223411&response_type=code&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fdiscord%2Fcallback&scope=identify+email"

	gitUser  = "https://api.github.com/user"
	gitEmail = "https://api.github.com/user/emails"
	git0auth = "https://github.com/login/oauth/authorize"
	format   = "application/x-www-form-urlencoded"
	// TO-DO: ADD linkedin
)

func AuthCallback(w http.ResponseWriter, r *http.Request) {
	var provider string
	// make sure who the provider is
	if strings.Contains(r.URL.Path, "google") {
		provider = "google"
	} else if strings.Contains(r.URL.Path, "github") {
		provider = "github"
	} else if strings.Contains(r.URL.Path, "discord") {
		provider = "discord"
	} else {
		Logger.Printf("Provider not found in path %v", r.URL.Path)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		Logger.Println("Code not found in the callback URL query params")
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token and fetch user profile
	accessToken, err := GetToken(code, provider)
	if err != nil {
		Logger.Printf("Error fetching access token: %v", err)
		http.Error(w, "Failed to fetch access token", http.StatusInternalServerError)
		return
	}

	// get user profile from provider
	var user User
	switch provider {
	case "google":
		err := user.GetGoogleProfile(accessToken)
		if err != nil {
			Logger.Printf("Unable to retrive user profile from %v, \n%v", provider, err)
			return
		}
	case "github":
		err := user.GetGitProfile(accessToken)
		if err != nil {
			Logger.Printf("Unable to retrive user profile from %v, \n%v", provider, err)
			return
		}
	case "discord":
		err := user.GetDiscordData(accessToken)
		if err != nil {
			Logger.Printf("Unable to retrive user profile from %v, \n%v", provider, err)
			return
		}
	default:
		Logger.Printf("Provider unspecified...State: %v", provider)
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}
	errors := ErrorMessages{}
	err = user.Check()
	if err != nil {
		Logger.Printf("User %v with this email already exists...", user.Name)
		Logger.Printf("Init auth for user %v", user.Name)
		user.AuthenticateExternal(r, errors)
		//user.AuthenticateExternal(r, errors)
	} else {
		Logger.Printf("User %v with this email does not exist...", user.Name)
		Logger.Printf("Init registration for user %v", user.Name)
		// insert the user into the DB
		user.Create()
	}

	// create session
	r = user.ReloadUserIntoContext(r) // assuming user is already authenticated.
	Logger.Printf("User loaded into context")
	Logger.Printf("Context after signin: %v", r.Context().Value(UserCtxKey))
	Logger.Printf("Context displayed after signin")
	user.RefreshCookies(w, r, "/home") // refresh cookies and redir home
	Logger.Printf("Cookies refreshed.")
}

func GoogleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		Logger.Printf("Invalid request method: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	params := url.Values{}
	params.Add("client_id", googleClientID)
	params.Add("redirect_uri", googleRedirectURI)
	params.Add("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")
	params.Add("response_type", "code")
	params.Add("access_type", "offline")

	http.Redirect(w, r, googleAuth+"?"+params.Encode(), http.StatusTemporaryRedirect)
}

func GithubAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		Logger.Printf("Requests not defined as GET are not allowed!")
		return
	}

	params := url.Values{}
	params.Add("client_id", githubClientID)
	params.Add("redirect_uri", githubRedirectURI)
	params.Add("scope", "user user:email")

	http.Redirect(w, r, git0auth+"?"+params.Encode(), http.StatusFound)
}

func DiscordAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Redirect to Discord's OAuth2.0 authorization page
	http.Redirect(w, r, discord0auth, http.StatusTemporaryRedirect)
}
