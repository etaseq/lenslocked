package controllers

import (
	"net/http"
)

// Notice that the StaticHandler receives a controllers.Template
// interface that I have defined in controllers/template.go.
// But when I call it from the routes I pass a views.Template
// struct instead. Nevertheless the views.Template implements
// the controllers.Template interface since it has the
// required Execute method.
func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	}
}
