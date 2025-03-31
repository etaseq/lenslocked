package controllers

import (
	"fmt"
	"net/http"

	"github.com/etaseq/lenslocked/context"
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
	ctx := r.Context()
	user := context.User(ctx)

	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	err = u.SessionService.Delete(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusFound)

}

// I am adding the middleware for the users here because the application
// is small at this point. This is the same reason I added the cookies
// inside the controllers package.
// In case I wanted to keep all separated I could rename the controllers
// package to "handlers" and only include the http handlers in it.
// Then I would create a separate package called middleware and another
// one called cookies.

// The SetUser middleware will look up the session token via the user's
// cookie, it will then query for a valid session using that token and
// if it finds a user it is going to set it inside of the context and
// eventually proceed with the next http handler [which is every handler
// in the app since I wrap the whole router (r) with it. So it is like
// doing r.Post("/signin", umw.SetUser(usersC.ProcessSignIn)) for each
// handler].
// If it will have an issue looking up the cookie or anything else it
// is going to proceed with the next http handler since this middleware
// is not designed to restrict access, it is just meant to set the user
// in the context.

type UserMiddleWare struct {
	SessionService *models.SessionService
}

func (umw UserMiddleWare) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// if a user has been found, get the Context set the value and
		// UPDATE THE REQUEST with the new context.
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleWare) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is present and if not redirect to the sign in page
		ctx := r.Context()
		user := context.User(ctx)
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
