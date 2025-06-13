package main

import (
	"fmt"

	"github.com/etaseq/lenslocked/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	gs := models.GalleryService{}
	fmt.Println(gs.Images(5))
}
