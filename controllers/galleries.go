package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
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

// Render all the galleries of a user.
func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	// While I could pass the full Gallery model directly to the template,
	// I'm creating a separate, simpler type specifically for the view.
	// This makes it clearer what data the template actually needs and
	// it allows to modify data from the database if needed.
	// It is also much safer; decoupling the database model from the
	// view-layer model helps avoid leaking internal state by potentially
	// exposing important fields to the view.
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

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorize
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	// While I could pass the full Image model directly to the template,
	// I'm creating a separate, simpler type specifically for the view.
	// All the code below is the same as in the Show handler.
	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	var data struct {
		ID     int
		Title  string
		Images []Image
	}
	data.ID = gallery.ID
	data.Title = gallery.Title

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorize
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorize
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	err = g.GalleryService.Delete(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// While I could pass the full Image model directly to the template,
	// I'm creating a separate, simpler type specifically for the view.
	// This makes it clearer what data the template actually needs and
	// also it allows to modify data from the database if needed.
	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	var data struct {
		ID     int
		Title  string
		Images []Image
	}
	data.ID = gallery.ID
	data.Title = gallery.Title

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}

	image, err := g.GalleryService.Image(galleryID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, image.Path)
}

func (g Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorize
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	// 5 << 20 is the equivalent of 5 * 1024 * 1024
	// which means that it allows to upload files 5mb max
	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// "images" is the name of the field in the html form.
	// So this is going to return all the file headers that were
	// uploaded under the name "images" in the html form.
	// NOTE: MultipartForm is the parsed multipart form, including file uploads.
	// This field is only available after ParseMultipartForm is called.
	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			var fileErr models.FileError
			if errors.As(err, &fileErr) {
				msg := fmt.Sprintf("%v has an invalid content type or extensions. "+
					"Only png, jpg and gif files can be uploaded.", fileHeader.Filename)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Authorize
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.
			StatusForbidden)
		return
	}

	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

// Sanitize the "filename" to prevent directory traversal attacks.
func (g Galleries) filename(w http.ResponseWriter, r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}

// Helper function to avoid minor repetition across Edit, Update, Show and
// Delete handlers. While the duplication isn't excessive to justify the
// decision, it improves readability and maintains cleaner handler logic.
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
