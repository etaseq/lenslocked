package main

import (
	"fmt"
	"net/http"

	"github.com/etaseq/lenslocked/controllers"
	"github.com/etaseq/lenslocked/models"
	"github.com/etaseq/lenslocked/templates"
	"github.com/etaseq/lenslocked/views"
	"github.com/go-chi/chi/v5"
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
	userService := models.UserService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService: &userService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.html", "tailwind.html",
	))
	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
