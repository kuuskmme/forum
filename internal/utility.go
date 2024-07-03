package internal

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// CONFIG
const (
	TemplatePath = "./templates/"                 // change this if your templates are located elsewhere
	DateTime     = "15:04:05 Monday, 02 Jan 2006" // change this if you want to store date-time in some other format
	Port         = "8080"                         // change this if you would like to host the server on some other port.
)

// Used to package template components together to generate the requested page
func generateHTML(w http.ResponseWriter, r *http.Request, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf(TemplatePath+"%s.gohtml", file)) // iterate over variadic filename input to create list of template names to parse
	}
	files = append(files, fmt.Sprintf(TemplatePath+"layout.gohtml")) // add these by default
	files = append(files, fmt.Sprintf(TemplatePath+"base_components.gohtml"))
	templates := template.Must(template.ParseFiles(files...)) // parse files variadically and construct the template
	err := templates.ExecuteTemplate(w, "layout", data)       // include our layout by default
	err500(w, nil, err)                                       // check for internal error
	if err != nil {
		Logger.Fatalln("Failed to execute template files", err)
	}
}

func generateErrorHTML(w http.ResponseWriter, r *http.Request, data interface{}) {
	template := template.Must(template.ParseFiles(TemplatePath + "oops.gohtml"))
	err := template.Execute(w, data) // include our error layout by default
	err500(w, nil, err)              // check for internal error
	if err != nil {
		Logger.Fatalln("Failed to execute template files", err)
	}
}

func generateERD(w http.ResponseWriter, r *http.Request, data interface{}) {
	template := template.Must(template.ParseFiles(TemplatePath + "ERD/mermaid.gohtml"))
	err := template.Execute(w, data) // include our mermaid template by default
	if err != nil {
		Logger.Fatalln("Failed to execute template files", err)
	}
}

func generateAJAXErrorHTML(w http.ResponseWriter, r *http.Request, data HttpError) {
	template := template.Must(template.ParseFiles(TemplatePath + "oops.gohtml"))
	err := template.Execute(w, data) // include our error layout by default
	err500(w, nil, err)              // check for internal error
	if err != nil {
		Logger.Fatalln("Failed to execute template files", err)
	}
}

func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		Logger.Printf("Failed to retrieve session cookie: %v", err)
		return nil, err
	}
	Logger.Printf("\033[1m\033[92mSession cookie retrieved: %v\033[0m", sessionCookie)
	return sessionCookie, nil
}

func setSessionCookie(w http.ResponseWriter, sessionUUID string) {
	cookie := http.Cookie{
		Name:     "session", // this is a custom value
		Value:    sessionUUID,
		Path:     "/",  // cookie accessible thru entire domain
		HttpOnly: true, // Helps mitigate the risk of client side script accessing the protected cookie
		MaxAge:   4000, // Add custom logic for this later via selector or checkboxes..
		Secure:   true,
		// Add Secure : true if we plan on adding HTTPS integration later!
	}
	http.SetCookie(w, &cookie)
}

// Unmarshal body data from JSON
func toJSON(r *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var mappedBody map[string]interface{}
	err = json.Unmarshal(body, &mappedBody)
	if err != nil {
		return nil, err
	}

	return mappedBody, nil
}
