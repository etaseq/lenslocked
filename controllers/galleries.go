package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/etaseq/lenslocked/context"
	"github.com/etaseq/lenslocked/models"
	"github.com/go-chi/chi/v5"
)

type Galleries struct {
	Templates struct {
		Show  Template
		New   Template
		Edit  Template
		Index Template
	}
	GalleryService *models.GalleryService
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}

	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID int
		Title  string
	}
	// I assume that the user is signed in because in the routes I
	// use the RequireUser middleware. This is the reason I use
	// context without even checking if the user is nil.
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	var data struct {
		ID     int
		Title  string
		Images []string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	// Fake Images for now
	for i := 0; i < 20; i++ {
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		catImageURL := fmt.Sprintf("https://picsum.photos/%d/%d", w, h)
		data.Images = append(data.Images, catImageURL)
	}
	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorization. Verify that the user has access to edit the gallery.
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	var data struct {
		ID    int
		Title string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	// All previous code is the same with Edit. This part just updates
	// the gallery.
	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

// Render all the galleries of a user.
func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	// While I could pass the full Gallery model directly to the template,
	// I'm creating a separate, simpler type specifically for the view.
	// This makes it clearer what data the template actually needs and
	// also it allows to modify data from the database if needed.
	type Gallery struct {
		ID    int
		Title string
	}

	var data struct {
		Galleries []Gallery
	}

	user := context.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID:    gallery.ID,
			Title: gallery.Title,
		})
	}
	g.Templates.Index.Execute(w, r, data)

}

// Helper function to avoid minor repetition across Edit, Update, and Show
// handlers. While the duplication isn't excessive to justify the decision,
// it improves readability and maintains cleaner handler logic
func (g Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (
	*models.Gallery, error) {
	// Get the {id} of the gallery I want to work with.
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, err
	}

	// Query for the gallery to make sure it exists.
	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, err
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}

	return gallery, nil
}
