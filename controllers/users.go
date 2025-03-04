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
	u.Templates.New.Execute(w, nil)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	// FormValue calls r.ParseForm first so I do not need
	// to call it separately
	fmt.Fprint(w, "Email: ", r.FormValue("email"))
	fmt.Fprint(w, "Password: ", r.FormValue("password"))
}
