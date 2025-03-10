package controllers

import (
	"fmt"
	"net/http"

	"github.com/etaseq/lenslocked/models"
)

// the Template type is an interface I define
// in controllers/template.go
type Users struct {
	Templates struct {
		New Template
	}
	UserService *models.UserService
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
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)

	if err != nil {
		fmt.Println()
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User created: %+v", user)
}
