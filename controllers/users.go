package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/etaseq/lenslocked/context"
	"github.com/etaseq/lenslocked/models"
)

// the Template type is an interface I define
// in controllers/template.go
type Users struct {
	Templates struct {
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
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
	var data struct {
		Email    string
		Password string
	}

	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Create(data.Email, data.Password)
	if err != nil {
		u.Templates.New.Execute(w, r, data, err)
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

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")

	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		// TODO: Handle other cases in the future. For instance, if a user does
		// exist with that email address.
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// The url.Values type is a map[string][]string. It is used to take
	// values I need to put to the url as query parameters.
	// Then I can use the Encode() method, which will return the query
	// parameters as a properly encoded string.
	vals := url.Values{
		"token": {pwReset.Token},
	}
	resetURL := "https://www.lenslocked.com/reset-pw?" + vals.Encode()

	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	// This is the token from the reset-pw url query
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		// TODO: Distinguish between types of errors.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Sign the user in now that the password has been reset.
	// Any errors from this point onwards should redirect the user to the
	// sign in page.
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
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
// if it finds a user, it is going to set it inside of the context and
// eventually proceed with the "next" http handler.
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

		// "next" here will be the usersC.CurrentUser
		next.ServeHTTP(w, r)
	})
}
