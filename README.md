# go-sqlite-sessions
Wraps [gorilla/sessions](https://github.com/gorilla/sessions) and [maxbarbieri/sqlitestore](https://github.com/maxbarbieri/sqlitestore) in a single package with a simpler interface.


Example:
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	gosqlitesessions "github.com/maxbarbieri/go-sqlite-sessions"
)

func main() {
	log.Println("maxbarbieri/go-sqlite-sessions example")

	//initialize sessions manager
	gosqlitesessions.InitializeSessionsManagerWithOptions(gosqlitesessions.Options{
		Path:                   "/",
		HttpOnly:               true,
		Secure:                 false,
		SqliteDatabaseFilename: "./sessions.sqlite",
		SessionCookieName:      "session-id",
		MaxAge:                 7200, //2h in seconds
		SecretKey:              []byte("super-secret-key"),
		CleanupInterval:        time.Minute * 5, //cleanup each 5 minute
		// Other possible options are
		// SameSite of type http.SameSite
		// Domain of type string
	})

	//set the callback function to be called before an expired session gets deleted
	gosqlitesessions.SetExpiredSessionPreDeleteCallback(func(expiredSession *sessions.Session) {
		log.Println("Session with ID", expiredSession.ID, "is expired and is going to be deleted from the database")
	})

	//define routes
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/protected-area", verifyLogin(protectedAreaHandler))
	http.HandleFunc("/logout", logoutHandler)

	log.Println("Listening at localhost:8080")

	//start http server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homePageHandler(resWriter http.ResponseWriter, request *http.Request) {
	fmt.Fprint(resWriter, "<html><body>Endpoints:<br>")
	fmt.Fprint(resWriter, "<a href=\"/login\">/login</a><br>")
	fmt.Fprint(resWriter, "<a href=\"/protected-area\">/protected-area</a><br>")
	fmt.Fprint(resWriter, "<a href=\"/logout\">/logout</a></body></html>")
}

func loginHandler(resWriter http.ResponseWriter, request *http.Request) {
	//get session info
	session, err := gosqlitesessions.GetSession(resWriter, request)
	if err != nil {
		sendHttpInternalServerError(resWriter, "An error occurred while getting session information: "+err.Error())
		return
	}

	// Authentication goes here ...

	session.Values["loggedIn"] = true

	//save the session
	if err := session.Save(request, resWriter); err != nil {
		sendHttpInternalServerError(resWriter, "An error occurred while saving session information: "+err.Error())
		return
	}

	log.Println("User logged in, sessionID:", session.ID)
}

//login check middleware
func verifyLogin(nextHandler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(resWriter http.ResponseWriter, request *http.Request) {
		//get session info
		session, err := gosqlitesessions.GetSession(resWriter, request)
		if err != nil {
			sendHttpInternalServerError(resWriter, "An error occurred while getting session information: "+err.Error())
			return
		}

		//if not logged in
		if auth, ok := session.Values["loggedIn"].(bool); !ok || !auth {
			sendUnauthorizedHttpError(resWriter) //send Unauthorized error
			return
		}

		//if logged in

		//save session to update expire date (both server-side session and cookie)
		if err := session.Save(request, resWriter); err != nil {
			sendHttpInternalServerError(resWriter, "An error occurred while saving session information: "+err.Error())
			return
		}

		//pass the request to the next handler
		nextHandler(resWriter, request)
	}
}

func logoutHandler(resWriter http.ResponseWriter, request *http.Request) {
	//get settion info
	session, _ := gosqlitesessions.GetSession(resWriter, request)

	log.Println("Logging out user with sessionID", session.ID)

	//delete the session and the cookie
	gosqlitesessions.DeleteSession(resWriter, request, session)
}

func protectedAreaHandler(resWriter http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(resWriter, "LOGIN PROTECTED AREA - YOU ARE LOGGED IN")
}

func sendUnauthorizedHttpError(resWriter http.ResponseWriter) {
	http.Error(resWriter, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func sendHttpInternalServerError(resWriter http.ResponseWriter, errorMessage string) {
	http.Error(resWriter, http.StatusText(http.StatusInternalServerError)+" "+errorMessage, http.StatusInternalServerError)
}
```
