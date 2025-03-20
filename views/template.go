package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

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

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
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
