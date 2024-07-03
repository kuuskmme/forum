# Literary Lions Forum Project

## Overview
The Literary Lions Forum project is a fullstack web-based forum that allows users to engage in discussions, manage categories, and interact with posts and comments. 

The forum incorporates advanced features like [AJAX](https://en.wikipedia.org/wiki/Ajax_(programming))  requests for dynamic content updates, [OAuth](https://en.wikipedia.org/wiki/OAuth) login capabilities, and session management.

## Features

- **User Profiles and Settings** [BONUS FUNCTIONALITY]
  - Profile pages accessible from a JavaScript-powered sidebar.
  - Profile pages contain data containg the users' posts, threads and liked posts.
  - Settings allow users to change their username and password.
  
- **Navigation and Categories**
  - Navigation bar includes access to various categories via FORUM button, access to the register/login route via LOGIN, access to the homepage via HOME and access to the search via SEARCH
  
- **Threads and Comments**
  - Registered users can add and delete categories
  - Categories contain threads where registered users can add/remove their own threads
  - Users can respond to thread via posts/comments and can also delete them on the same page.
  
- **Likes/Dislikes and Views**
  - Users can also like or dislike threads and posts.
  - An eye icon shows the count of unique views per thread, updated via JavaScript on DOM load. [BONUS FUNCTIONALITY]
  
- **Session Management**
  - Utilizes custom session headers and cookies to maintain context for registered users.
  - Utilizes contextkeys to cut down on database requests when reloading the user dash.
  
- **Authentication**
  - Supports OAuth with GitHub, Google, and LinkedIn.
  - Local sign-up available with secure password requirements.
  - Sign up also invalidates emails with incorrect format along with detecting duplicate names/emails.
  
- **Search Functionality**
  - Allows users to search for threads by categories, user details, post details, or thread specifics.
  
- **Directory Structure**
  - `server/cmd`: Contains the `main.go` entry point.
  - `data/forum.db`: Houses the project's SQLite3 database.
  - `/internal`: Contains all business logic.
  - `/templates`: Stores `.gohtml` template files.
  - `/static`: Includes `/js`, `/js/ajax`, `/img`, and `/css` directories.
  - `/static/js`: JavaScript files including global functionality (global.js) and AJAX-specific scripts located under `/ajax`.
  - `/static/img`: Images used within the forum.
  - `/static/css`: CSS files named mostly correspondingly to their Go HTML templates.

## Technologies and Libraries

- **Backend**: Go (Golang)
- **Frontend**: JavaScript, HTML5, CSS3
- **Database**: SQLite3
- **Authentication**: OAuth 2.0 [BONUS FUNCTIONALITY] along with inhouse custom auth
- **Session Management**: Custom session handling in Go

- **Go Templates**: For server-side rendering.
- **Google UUID**: Utilizes [Google's UUID package](https://github.com/google/uuid) to generate UUIDs for user sessions. [BONUS FUNCTIONALITY]
- **Bluemonday**: Utilizes [Bluemonday package](https://github.com/microcosm-cc/bluemonday) for sanitizing user inputs to prevent XSS attacks. [BONUS FUNCTIONALITY]
- **Go-SQLite3**: Utilizes [Mattns' Go-SQLITE3](https://github.com/mattn/go-sqlite3) for driving SQLite3 in Golang.
- **SQL JOINs**: Advanced SQL queries, leveraging JOIN statements to optimize database interactions and maintain data in 3NF (Third Normal Form). [BONUS FUNCTIONALITY]
- **ERD with MermaidJS**: [Advanced ERD diagram created with MermaidJS](http://localhost:8080/view-erd), available directly inside the web app itself (make sure project is running before clicking hyperlink)
- **Docker**: Includes Dockerfile and build scripts for containerization and easy deployment.

## Database Design

The project includes an Entity-Relationship Diagram (ERD) to represent the database schema which is normalized to the third normal form to ensure minimal redundancy and dependency.

## Error Handling

Implements robust error handling across both Go template-driven pages and AJAX routes, ensuring a graceful user experience during exceptions.

## Running the Project

### MAKE SURE YOU ADD PRIVILEGES TO THE BASH SCRIPTS BEFORE EXECUTION!
#### !After navigating to the root directory of the project!

    sudo chmod +x startup.sh chmod +x dockerbuild.sh

### MAKE SURE YOU ARE NOT DOING ANYTHING ELSE DOCKER-RELATED AS THE SCRIPT BY DEFAULT PRUNES STOPPED CONTAINERS, UNUSED NETWORKS AND DANGLING IMAGES!
- Ensure port `8080` is clear! 

You can run the forum on other ports via changing the `const` in `utility.go`,
    but the 0auth is specifically configured to run only via port 8080!
- **Docker Setup**:
  - Build the container: 

    ```bash dockerbuild.sh```.
    
- **Running without Docker**:
  
    ```bash startup.sh```.
    
    The bash scripts set the environment variables along with the Docker script automatically unpacking the project itself.
`

## Future

Project has database tables to implement images associated with posts (same can be done for threads). This could be implemented as an improvement.

AJAX logic can be systematized further and likely refactored, especially if this project were to have more functionality added to it, as the current ajax_handler.go file is rather large.

User recovery has not been fully set up. This should involve integration with an existing framework to send emails and generate links or custom implementation!

## Credits

User Dash base styling was grabbed from Codepen and edited 

https://codepen.io/FlorinPop17/pen/vPKWjd

Nav Bar base template styling grabbed from Codepen and edited

https://codepen.io/jhancock532/pen/GRZrLwY

Thanks to Florin Pop and James Hancock for their amazing work and indirect CSS lesson!