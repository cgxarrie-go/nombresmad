package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	_ "github.com/lib/pq"

	"nombresmad/backend/internal/api"
	"nombresmad/backend/internal/repo"
	"nombresmad/backend/internal/service"
)

func main() {
	db, err := sql.Open("postgres", databaseURL())
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
	port := getenv("PORT", "8080")
	fmt.Printf("Server listening on :%s\n", port)
	fmt.Println("Swagger UI at http://localhost:8080/swagger/")
	log.Fatal(http.ListenAndServe(":"+port, server.Handler()))
}

func databaseURL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getenv("PGUSER", "postgres"), getenv("PGPASSWORD", "postgres")),
		Host:     net.JoinHostPort(getenv("PGHOST", "localhost"), getenv("PGPORT", "5432")),
		Path:     getenv("PGDATABASE", "nombresmad"),
		RawQuery: url.Values{"sslmode": {getenv("PGSSLMODE", "disable")}}.Encode(),
	}
	return u.String()
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
