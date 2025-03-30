package main

import (
	"fmt"
	"net/http"

	"github.com/etaseq/lenslocked/controllers"
	"github.com/etaseq/lenslocked/migrations"
	"github.com/etaseq/lenslocked/models"
	"github.com/etaseq/lenslocked/templates"
	"github.com/etaseq/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {
	r := chi.NewRouter()
	// Parse all templates at start up. If you were parsing them every
	// time a request comes in, it would be much slower.
	tpl := views.Must(views.ParseFS(templates.FS, "home.html", "tailwind.html"))
	r.Get("/", controllers.StaticHandler(tpl))

	tpl = views.Must(views.ParseFS(templates.FS, "contact.html", "tailwind.html"))
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl = views.Must(views.ParseFS(templates.FS, "faq.html", "tailwind.html"))
	r.Get("/faq", controllers.StaticHandler(tpl))

	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Run the migrations when the application starts up
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.html", "tailwind.html",
	))

	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.html", "tailwind.html",
	))
	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/users/me", usersC.CurrentUser)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Create an instance of User middleware
	umw := controllers.UserMiddleWare{
		SessionService: &sessionService,
	}

	// When a user first accesses your site, this middleware generates a CSRF
	// token and stores it in the _gorilla_csrf cookie.
	// Every time a user submits a form (like signing up or logging in), the
	// CSRF token needs to be included in the form submission.
	// This can be done using the csrf.TemplateField(r) function, which
	// automatically retrieves the CSRF token from the cookie and generates
	// a hidden input field with it.
	csrfKey := "VWNEO674goZGNWpw20t49v0n1984fcCE"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: Fix this before deploying
		csrf.Secure(false),
	)

	// The csrf middleware will run first, then the umw
	// and finally the router will kick in
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", csrfMw(umw.SetUser(r)))
}
