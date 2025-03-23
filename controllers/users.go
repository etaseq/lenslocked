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
		New    Template
		SignIn Template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
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
	u.Templates.New.Execute(w, r, data)
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

	// I want to create a session and put it in a cookie AFTER I
	// know that the user has been created.
	// Notice that instead of ((*user).ID) I do (user.ID) although
	// the user is a pointer returned from models.Create.
	// In Go when I have pointer to a struct like User, I can access
	// the fields of that struct using either the pointer or the
	// struct itself, thanks to automatic dereferencing.
	// It is perfectly fine to use this as well ((*user).ID) although
	// the idiomatic approach is (user.ID).
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		// TODO: Long term, I should show a warning about not being able to sign in
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Replaced this with the helper setCookie below
	//cookie := http.Cookie{
	//	Name:     "session",
	//	Value:    session.Token,
	//	Path:     "/",
	//	HttpOnly: true,
	//}
	//http.SetCookie(w, &cookie)
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")

	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

// The goal of this function is to take a web request and print
// out the current user's information
func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	// Instead of this you can use the readCookie helper which is really
	// not necessary. Do not do this in your project!
	//tokenCookie, err := r.Cookie("CookieSession")
	token, err := readCookie(r, CookieSession)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Change this as well since I am using readCookie helper
	//user, err := u.SessionService.User(tokenCookie.Value)
	user, err := u.SessionService.User(token)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}
