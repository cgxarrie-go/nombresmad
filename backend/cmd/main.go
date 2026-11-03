package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"nombresmad/backend/internal/api"
	"nombresmad/backend/internal/repo"
	"nombresmad/backend/internal/service"
)

func main() {
	config, err := readAuthConfig()
	if err != nil {
		log.Fatal(err)
	}
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
	auth := service.NewAuth(repo.NewAdminRepo(db), config.password, config.secret, config.googleClientID, nil)
	if err = auth.EnsureDefaultAdmin(os.Getenv("ADMIN_USER")); err != nil {
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

type authConfig struct {
	googleClientID string
	password       string
	secret         []byte
}

func readAuthConfig() (authConfig, error) {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if clientID == "" {
		return authConfig{}, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	secret := strings.TrimSpace(os.Getenv("ADMIN_SECRET"))
	if secret == "" {
		return authConfig{}, fmt.Errorf("ADMIN_SECRET is required and must remain stable")
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		log.Println("ADMIN_PASSWORD is empty; default admin password and resets are disabled")
	}
	return authConfig{
		googleClientID: clientID,
		password:       password,
		secret:         []byte(secret),
	}, nil
}
