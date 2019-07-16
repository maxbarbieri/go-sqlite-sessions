package gosqlitesessions

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/maxbarbieri/sqlitestore"
)

type Options struct {
	SameSite               http.SameSite
	Domain                 string
	Path                   string
	HttpOnly               bool
	Secure                 bool
	MaxAge                 int
	SessionCookieName      string
	SqliteDatabaseFilename string
	SecretKey              []byte
	CleanupInterval        time.Duration
}

var (
	options       Options
	sessionsStore *sqlitestore.SqliteStore
)

func InitializeSessionsManager() {
	//default values for options
	options.Path = "/"
	options.HttpOnly = true
	options.Secure = false
	options.MaxAge = 3600
	options.SessionCookieName = "session-id"
	options.SqliteDatabaseFilename = "./sessions.sqlite"
	options.SecretKey = []byte("super-secret-key")
	options.CleanupInterval = time.Minute * 5

	//create a sqliteStore for sessions
	createSqliteStore()
}

func InitializeSessionsManagerWithOptions(opt Options) {
	options = opt

	//create a sqliteStore for sessions
	createSqliteStore()
}

func createSqliteStore() {
	sessionsOptions := sessions.Options{
		Path:     options.Path,
		MaxAge:   options.MaxAge,
		Domain:   options.Domain,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}

	var err error
	sessionsStore, err = sqlitestore.NewSqliteStore(options.SqliteDatabaseFilename, "sessions", sessionsOptions, options.SecretKey)
	if err != nil {
		log.Println("An error occurred while creating the store for sessions:", err.Error())
	}

	//start periodic cleanup of expired sessions from the database
	sessionsStore.StartCleanup(options.SessionCookieName, options.CleanupInterval)
}

func GetSession(resWriter http.ResponseWriter, request *http.Request) (*sessions.Session, error) {
	//get session from store
	session, err := sessionsStore.Get(request, options.SessionCookieName)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func DeleteSession(resWriter http.ResponseWriter, request *http.Request, session *sessions.Session) {
	sessionsStore.Delete(request, resWriter, session)
}

func DeleteFromDatabaseSessionWithID(sessionID string) error {
	return sessionsStore.DeleteFromDatabaseSessionWithID(sessionID)
}

func SetExpiredSessionPreDeleteCallback(callback func(*sessions.Session)) {
	sessionsStore.SetExpiredSessionPreDeleteCallback(callback)
}
