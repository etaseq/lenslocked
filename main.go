package main

import (
	"fmt"
	"net/http"

	"github.com/etaseq/lenslocked/controllers"
	"github.com/etaseq/lenslocked/templates"
	"github.com/etaseq/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	tpl, err := views.ParseFS(templates.FS, "home.html", "tailwind.html")
	if err != nil {
		panic(err)
	}
	r.Get("/", controllers.StaticHandler(tpl))

	tpl, err = views.ParseFS(templates.FS, "contact.html", "tailwind.html")
	if err != nil {
		panic(err)
	}
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl, err = views.ParseFS(templates.FS, "faq.html", "tailwind.html")
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.StaticHandler(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
