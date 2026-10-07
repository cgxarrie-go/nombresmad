package api

import (
	_ "embed"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
)

//go:embed openapi.json
var openAPISpec string

type openAPIDoc struct{}

func (openAPIDoc) ReadDoc() string { return openAPISpec }

func init() {
	swag.Register(swag.Name, openAPIDoc{})
}

func registerSwagger(mux *http.ServeMux) {
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
