package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/etaseq/lenslocked/context"
	"github.com/etaseq/lenslocked/models"
	"github.com/gorilla/csrf"
)

type public interface {
	Public() string
}

type Template struct {
	htmlTpl *template.Template
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

// ParseFS loads & parses templates from an embedded filesystem (fs.FS).
// This file system is defined in the templates/fs.go.
// The reason to use fs.FS and embed is that the final Go binary contains
// everything, which means better performance and security.
func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	tpl := template.New(patterns[0])
	// This is how to create a template function.
	// In this case I have created just a placeholder function so that
	// when the template is parsed doesn't return an error.
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("currentUser not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)

	// Always define your template functions first
	// and then parse the template afterward. If you try to parse
	// the template first and this template happens to have the
	// field {{ csrfField }} it will throw an error because
	// it cannot find the template function.
	tpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}

	return Template{
		htmlTpl: tpl,
	}, nil
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{},
	errs ...error) {
	// When you call tpl.Execute(), it can modify the internal state of the
	// template object. Cloning ensures each request gets a fresh copy to work with.
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}

	errMsgs := errMessages(errs...)
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				// The csrf.TemplateField(r) looks for the CSRF token that was set in
				// the cookie (_gorilla_csrf), then it renders that token into
				// the HTML 'hidden' form.
				// This is what is returned inside the {{csrfField}}
				// <input type="hidden" name="_csrf" value="VWNEO674goZGNWpw20t49v0n1984fcCE">
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			"errors": func() []string {
				return errMsgs
			},
		},
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Make sure that the page is fully rendered before send it to the writer.
	// Keep in mind that if you have a massive page this might cause performance
	// issues. So you should weight whether or not in case of an error you
	// would prefer a half rendered page or a slower rendering.
	// If you have pages that have a couple MBs of HTML rethink to use a buffer.
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

func errMessages(errs ...error) []string {
	var msgs []string

	for _, err := range errs {
		var pubErr public
		// Check if this error is of a type that satisfies the public interface
		// by implementing the Public() method.
		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			fmt.Println(err)
			msgs = append(msgs, "Something went wrong.")
		}
	}

	return msgs
}
