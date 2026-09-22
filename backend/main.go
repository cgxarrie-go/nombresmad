package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type Item struct {
	ID               int    `json:"id"`
	DateCreated      string `json:"dateCreated"`
	Text             string `json:"text"`
	Size             string `json:"size"`
	Thicknesses      string `json:"thicknesses"`
	WoodType         string `json:"woodType"`
	DeliveryDate     string `json:"deliveryDate"`
	DeliveredTo      string `json:"deliveredTo"`
	Group            string `json:"group"`
	Price            int    `json:"price"`
	DeliveredWithBox bool   `json:"deliveredWithBox"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/nombresmad?sslmode=disable"
	}
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/items", withCORS(itemsHandler))
	http.HandleFunc("/api/items/", withCORS(itemHandler))
	http.HandleFunc("/api/import", withCORS(importHandler))
	http.HandleFunc("/api/migrate", withCORS(migrateHandler))

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.Query(`SELECT id, dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox FROM items ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		items := []Item{}
		for rows.Next() {
			var it Item
			var deliveredWithBox bool
			if err := rows.Scan(&it.ID, &it.DateCreated, &it.Text, &it.Size, &it.Thicknesses, &it.WoodType, &it.DeliveryDate, &it.DeliveredTo, &it.Group, &it.Price, &deliveredWithBox); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			it.DeliveredWithBox = deliveredWithBox
			items = append(items, it)
		}
		writeJSON(w, items)
	case "POST":
		var it Item
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow(`INSERT INTO items(dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox).Scan(&it.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, it)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// itemHandler handles requests for /api/items/{id}
func itemHandler(w http.ResponseWriter, r *http.Request) {
	// expect /api/items/{id}
	idStr := r.URL.Path[len("/api/items/"):]
	if idStr == "" {
		http.Error(w, "missing id", 400)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}
	switch r.Method {
	case "GET":
		var it Item
		row := db.QueryRow(`SELECT id, dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox FROM items WHERE id=$1`, id)
		var deliveredWithBox bool
		if err := row.Scan(&it.ID, &it.DateCreated, &it.Text, &it.Size, &it.Thicknesses, &it.WoodType, &it.DeliveryDate, &it.DeliveredTo, &it.Group, &it.Price, &deliveredWithBox); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		it.DeliveredWithBox = deliveredWithBox
		writeJSON(w, it)
	case "PUT":
		var it Item
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec(`UPDATE items SET dateCreated=$1, text=$2, size=$3, thicknesses=$4, woodType=$5, deliveryDate=$6, deliveredTo=$7, "group"=$8, price=$9, deliveredWithBox=$10 WHERE id=$11`, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]string{"status": "updated"})
	case "DELETE":
		_, err := db.Exec(`DELETE FROM items WHERE id=$1`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]string{"status": "deleted"})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// withCORS wraps handlers to add CORS headers and handle OPTIONS preflight
func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		h(w, r)
	}
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	// read NoMad.json from repo root
	data, err := ioutil.ReadFile("../NoMad.json")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// try to detect UTF-16 and convert
	var objs []Item
	if err := tryUnmarshalWithFallback(data, &objs); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`TRUNCATE items RESTART IDENTITY`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for _, it := range objs {
		_, err := tx.Exec(`INSERT INTO items(id, dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, it.ID, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]string{"status": "imported"})
}

// migrateHandler applies migrations from migrations.sql to the connected database
func migrateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	f, err := os.Open("migrations.sql")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if _, err := db.Exec(string(b)); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]string{"status": "migrated"})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// tryUnmarshalWithFallback attempts JSON unmarshal and falls back to UTF-16LE -> UTF-8 conversion
func tryUnmarshalWithFallback(data []byte, out interface{}) error {
	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}
	// try naive UTF-16LE to UTF-8 conversion
	buf := make([]byte, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		b1 := data[i]
		b2 := data[i+1]
		// skip BOM
		if i == 0 && b1 == 0xff && b2 == 0xfe {
			continue
		}
		if b1 == 0 && b2 == 0 {
			continue
		}
		buf = append(buf, b1)
	}
	return json.Unmarshal(buf, out)
}
