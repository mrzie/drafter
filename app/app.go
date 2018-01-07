package app

import (
	. "drafter/setting"
	"net/http"
	"time"

	"drafter/service"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var session_store = sessions.NewCookieStore([]byte(Settings.Salt.SessionSalt))

var R = mux.NewRouter()

var (
	blogS     = service.BlogService
	tagS      = service.TagService
	noteS     = service.NoteService
	authS     = service.AuthService
	notebookS = service.NotebookService
	readerS   = service.ReaderService
	configS   = service.ConfigService
)

func init() {
	R.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	API_v1(R.PathPrefix("/v1").Subrouter())
	R.PathPrefix("/admin").HandlerFunc(adminViews)
	R.HandleFunc("/blog/{id}", blogView)
	R.HandleFunc("/tag/{tag}", mainView)
	R.HandleFunc("/", mainView)
	R.HandleFunc("/error", errorView)
}

func API_v1(router *mux.Router) {
	// r := new(Router)
	// r.r = router
	r := NewRouter(router)

	r.Handle("/test", "GET,POST", func(ctx *context) error {
		var rsp = struct {
			Ok  bool   `json:"ok"`
			Msg string `json:"msg"`
			Now int64  `json:"now"`
		}{true, "hello", time.Now().Unix() * 1000}
		ctx.SendJson(rsp)
		return nil
	})

	r.Handle("/test2/*", "GET", func(ctx *context) error {
		ctx.Send([]byte("blocked"))
		return nil
	})

	r.Handle("/blogs", "GET", ListBlogHandler)
	r.Handle("/blog/{id}", "GET", GetBlogHandler)
	r.Handle("/tags", "GET", GetTagsHandler)
	r.Handle("/login", "POST", LoginHandler)

	admin := r.HandleMidware("/admin", "", VerifyController)
	admin.Handle("/editPassword", "POST", EditPasswordController)

	admin.Handle("/notebook", "POST", NewNotebookController)
	admin.Handle("/notebook", "DELETE", DeleteNotebookController)
	admin.Handle("/notebook", "PATCH", RenameNotebookController)
	admin.Handle("/notebooks", "GET", ListNotebookController)

	admin.Handle("/notes", "GET", ListNoteController)
	admin.Handle("/wastenote", "GET", ListWasteNoteController)
	admin.Handle("/note", "GET", GetNotesController)
	admin.Handle("/note", "POST", NewNoteController)
	admin.Handle("/note/{id}", "PATCH", EditNoteController)
	admin.Handle("/note/{id}", "DELETE", RemoveNoteController)

	admin.Handle("/blog", "GET", GetBlogController)
	admin.Handle("/blog", "POST", ComposeBlogController)
	admin.Handle("/blog/{id}", "PUT", EditBlogController)
	admin.Handle("/blogs", "GET", ListBlogController)
	admin.Handle("/blog/{id}", "DELETE", RemoveBlogController)

	admin.Handle("/user-preference", "GET", UserPreferenceController)
	admin.Handle("/user-preference", "PUT", SetUserPreferenceController)

	admin.Handle("/describe-tag", "PUT", DescribeTagController)
	admin.Handle("/tags", "GET", ListTagsController)
	admin.Handle("/tag/{name}", "DELETE", DeleteTagController)
	// r.HandlePrefix("/admin", "GET,POST,PUT,PATCH,OPTIONS,HEAD", VerifyController)
}
