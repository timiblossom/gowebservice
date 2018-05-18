package route

import (
	"net/http"

	"app/controller"
	"app/route/middleware/acl"
	hr "app/route/middleware/httprouterwrapper"
	"app/route/middleware/logrequest"
	"app/route/middleware/pprofhandler"
	"app/shared/session"

	"github.com/gorilla/context"
	"github.com/josephspurrier/csrfbanana"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// Load returns the routes and middleware
func Load() http.Handler {
	return routes()
	//return middleware(routes())
}

// LoadHTTPS returns the HTTP routes and middleware
func LoadHTTPS() http.Handler {
	return middleware(routes())
}

// LoadHTTP returns the HTTPS routes and middleware
func LoadHTTP() http.Handler {
	return middleware(routes())

	// Uncomment this and comment out the line above to always redirect to HTTPS
	//return http.HandlerFunc(redirectToHTTPS)
}

// Optional method to make it easy to redirect from HTTP to HTTPS
func redirectToHTTPS(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "https://"+req.Host, http.StatusMovedPermanently)
}

// *****************************************************************************
// Routes
// *****************************************************************************

func routes() *httprouter.Router {
	r := httprouter.New()

	// // Set 404 handler
	r.NotFound = func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}(alice.
		New().
		ThenFunc(controller.Error404))

	// Serve static files, no directory browsing
	r.GET("/static/*filepath", hr.Handler(alice.
		New().
		ThenFunc(controller.Static)))

	// Home page
	r.GET("/", hr.Handler(alice.
		New().
		ThenFunc(controller.IndexGET)))

	// Login
	r.GET("/login", hr.Handler(alice.
		New(acl.DisallowAuth).
		Append(acl.AllowCORS).
		ThenFunc(controller.LoginGET)))
	r.POST("/login", hr.Handler(alice.
		New(acl.DisallowAuth).
		Append(acl.AllowCORS).
		ThenFunc(controller.LoginPOST)))
	r.GET("/logout", hr.Handler(alice.
		New().
		ThenFunc(controller.LogoutGET)))

	// Register
	r.GET("/register", hr.Handler(alice.
		New(acl.DisallowAuth).
		ThenFunc(controller.RegisterGET)))
	r.POST("/register", hr.Handler(alice.
		New(acl.DisallowAuth).
		ThenFunc(controller.RegisterPOST)))

	// About
	r.GET("/about", hr.Handler(alice.
		New().
		ThenFunc(controller.AboutGET)))


	// Enable Pprof
	r.GET("/debug/pprof/*pprof", hr.Handler(alice.
		New(acl.DisallowAnon).
		ThenFunc(pprofhandler.Handler)))

	r.POST("/api/posts", hr.Handler(alice.
		New(acl.DisallowAnon).
		ThenFunc(controller.Posts)))

	r.GET("/api/get", hr.Handler(alice.
		New(acl.DisallowAnon).
		ThenFunc(controller.Posts)))

	//***************************************************************************
	// Public Rest APIs
	//***************************************************************************
	// Public API: Best rates
	r.POST("/api/public/rate/list", hr.Handler(alice.
		New().
		ThenFunc(controller.LenderRateList)))

	r.POST("/api/public/fast_quote", hr.Handler(alice.
		New().
		ThenFunc(controller.FastQuotePost)))

	//***************************************************************************
	// Admin Rest APIs
	//***************************************************************************
	// Admin API: Register
	r.POST("/api/admin/register", hr.Handler(alice.
		New(acl.AllowCORS).
		ThenFunc(controller.UserRegisterPost)))

	// Admin API: Login
	r.POST("/api/admin/login", hr.Handler(alice.
		New(acl.AllowCORS).
		ThenFunc(controller.UserLoginPost)))


	//***************************************************************************
	// Customer Rest APIs
	//***************************************************************************
	// Customer API: Register
	r.POST("/api/customer/register", hr.Handler(alice.
		New(acl.AllowCORS).
		ThenFunc(controller.CustomerRegisterPost)))

	// Customer API: Login
	r.POST("/api/customer/login", hr.Handler(alice.
		New(acl.AllowCORS).
		ThenFunc(controller.CustomerLoginPost)))

	// Customer API: Logout
	r.POST("/api/customer/logout", hr.Handler(alice.
		New(acl.AllowCORS).
		ThenFunc(controller.CustomerLogoutPost)))

	// Customer API: Profile - returns user profile info
	r.POST("/api/customer/profile", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerProfileGetInfo)))

	// Customer API: Profile - put new user profile info
	r.PATCH("/api/customer/profile", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerProfilePatch)))

	// Customer API: Upload file
	r.POST("/api/customer/file/upload", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerUploadFilePost)))

	// Customer API: Get file list
	r.GET("/api/customer/file/list", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerFilesListGet)))

	// Customer API: Get file by id
	r.GET("/api/customer/file/get", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerServeFileGet)))

	// Customer API: Delete file
	r.DELETE("/api/customer/file", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.CustomerFileDelete)))

	//***************************************************************************
	// Messaging Rest APIs
	//***************************************************************************

	// Customer API: Get customer message list
	r.POST("/api/user/message/threads", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.MessageThreads)))

	// Customer API: Get customer message list
	r.POST("/api/user/message/list", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.UserMessageList)))

	r.POST("/api/user/message", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.UserMessagePost)))

	// Messaging API: mark message as readed
	r.PATCH("/api/user/message", hr.Handler(alice.
		New(acl.DisallowAnon).Append(acl.AllowCORS).
		ThenFunc(controller.MarkMessageAsReaded)))


	return r
}

// *****************************************************************************
// Middleware
// *****************************************************************************

func middleware(h http.Handler) http.Handler {
	// Prevents CSRF and Double Submits
	cs := csrfbanana.New(h, session.Store, session.Name)
	cs.FailureHandler(http.HandlerFunc(controller.InvalidToken))
	cs.ClearAfterUsage(true)
	cs.ExcludeRegexPaths([]string{"/static(.*)"})
	cs.ExcludeRegexPaths([]string{"/api/admin(.*)"})
	cs.ExcludeRegexPaths([]string{"/api/user(.*)"})
	cs.ExcludeRegexPaths([]string{"/api/customer(.*)"})
	cs.ExcludeRegexPaths([]string{"/api/public(.*)"})

	csrfbanana.TokenLength = 32
	csrfbanana.TokenName = "token"
	csrfbanana.SingleToken = false

	h = cs
	// Log every request
	h = logrequest.Handler(h)

	// Clear handler for Gorilla Context
	h = context.ClearHandler(h)

	return h
}
