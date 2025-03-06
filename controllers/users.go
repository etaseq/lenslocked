package controllers

import (
	"fmt"
	"net/http"
)

// the Template type is an interface I define
// in controllers/template.go
type Users struct {
	Templates struct {
		New Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	// This is essentially checking the "email" field in the request
	// (from the query string or form data). In the case of a GET
	// request, the email can only be sent via the query string
	// (like, /signup?email=something@example.com)
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	// FormValue calls r.ParseForm first so I do not need
	// to call it separately
	fmt.Fprint(w, "Email: ", r.FormValue("email"))
	fmt.Fprint(w, "Password: ", r.FormValue("password"))
}
