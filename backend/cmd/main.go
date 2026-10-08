package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"nombresmad/backend/internal/api"
	"nombresmad/backend/internal/repo"
	"nombresmad/backend/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/nombresmad?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	items := repo.NewItemRepo(db)
	pictures := repo.NewPictureRepo(os.Getenv("PICTURES_DIR"))
	catalog := service.NewCatalog(items, pictures)
	if err = catalog.Bootstrap(); err != nil {
		log.Fatal(err)
	}
	if err = pictures.EnsureDir(); err != nil {
		log.Fatal(err)
	}
	auth, err := newAuth()
	if err != nil {
		log.Fatal(err)
	}

	server := api.New(
		service.NewItems(items, pictures),
		service.NewPictures(items, pictures),
		catalog,
		auth,
	)
	fmt.Println("Server listening on :8080")
	fmt.Println("Swagger UI at http://localhost:8080/swagger/")
	log.Fatal(http.ListenAndServe(":8080", server.Handler()))
}

func newAuth() (*service.Auth, error) {
	user := os.Getenv("ADMIN_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	var secret []byte
	if value := os.Getenv("ADMIN_SECRET"); value != "" {
		secret = []byte(value)
	} else {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
	}
	if password == "" {
		log.Println("ADMIN_PASSWORD is empty; admin login is disabled")
	}
	return service.NewAuth(user, password, secret), nil
}
