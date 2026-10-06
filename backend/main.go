package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"

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
	Picture          string `json:"picture"`
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
	if err = applyMigrations(); err != nil {
		log.Fatal(err)
	}
	if err = seedIfEmpty(); err != nil {
		log.Fatal(err)
	}
	if err = syncIDSequence(); err != nil {
		log.Fatal(err)
	}
	if err = os.MkdirAll(picturesDir(), 0o755); err != nil {
		log.Fatal(err)
	}
	initAuth()

	http.HandleFunc("/api/login", withCORS(loginHandler))
	http.HandleFunc("/api/logout", withCORS(logoutHandler))
	http.HandleFunc("/api/session", withCORS(sessionHandler))
	http.HandleFunc("/api/items", withCORS(itemsHandler))
	http.HandleFunc("/api/items/", withCORS(itemHandler))
	http.HandleFunc("/api/options", withCORS(optionsHandler))
	http.HandleFunc("/api/import", withCORS(importHandler))
	http.HandleFunc("/api/migrate", withCORS(migrateHandler))
	registerSwagger()

	fmt.Println("Server listening on :8080")
	fmt.Println("Swagger UI at http://localhost:8080/swagger/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listItems(w, r)
	case "POST":
		if !requireAuth(w, r) {
			return
		}
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

const pageSize = 25

const itemSelect = `id, COALESCE(dateCreated, ''), COALESCE(text, ''), COALESCE(size, ''), COALESCE(thicknesses, ''), COALESCE(woodType, ''), COALESCE(deliveryDate, ''), COALESCE(deliveredTo, ''), COALESCE("group", ''), COALESCE(price, 0), COALESCE(deliveredWithBox, false), COALESCE(picture, '')`

type itemPage struct {
	Items      []Item `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
}

type listQuery struct {
	Page     int
	PageSize int
	WhereSQL string
	Args     []any
	OrderSQL string
}

var sortColumns = map[string]string{
	"id":               "id",
	"dateCreated":      dateSortExpr("dateCreated"),
	"text":             "text",
	"size":             numericSortExpr("size"),
	"thicknesses":      numericSortExpr("thicknesses"),
	"woodType":         "woodType",
	"deliveryDate":     dateSortExpr("deliveryDate"),
	"deliveredTo":      "deliveredTo",
	"group":            `"group"`,
	"price":            "price",
	"deliveredWithBox": "deliveredWithBox",
}

func dateSortExpr(col string) string {
	return fmt.Sprintf(`CASE WHEN %s ~ '^[0-9]{2}/[0-9]{2}/[0-9]{4}$' THEN to_date(%s, 'DD/MM/YYYY') END`, col, col)
}

func numericSortExpr(col string) string {
	return fmt.Sprintf(`CASE WHEN %s ~ '^[0-9]+([.][0-9]+)?$' THEN %s::numeric END`, col, col)
}

func parseListQuery(values url.Values) (listQuery, error) {
	lq := listQuery{Page: 1, PageSize: pageSize, WhereSQL: "TRUE"}
	if raw := values.Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return lq, fmt.Errorf("invalid page")
		}
		lq.Page = n
	}
	if raw := values.Get("pageSize"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 60 {
			return lq, fmt.Errorf("invalid pageSize")
		}
		lq.PageSize = n
	}
	sortKey := values.Get("sort")
	if sortKey == "" {
		sortKey = "id"
	}
	expr, ok := sortColumns[sortKey]
	if !ok {
		return lq, fmt.Errorf("invalid sort")
	}
	order := strings.ToLower(values.Get("order"))
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		return lq, fmt.Errorf("invalid order")
	}
	dir := "ASC"
	if order == "desc" {
		dir = "DESC"
	}
	if sortKey == "id" {
		lq.OrderSQL = "id " + dir
	} else {
		lq.OrderSQL = expr + " " + dir + ", id ASC"
	}

	var where []string
	if s := strings.TrimSpace(values.Get("q")); s != "" {
		lq.Args = append(lq.Args, likePattern(s))
		p := len(lq.Args)
		where = append(where, fmt.Sprintf(`(
			COALESCE(text, '') ILIKE $%d ESCAPE '\' OR
			COALESCE("group", '') ILIKE $%d ESCAPE '\' OR
			COALESCE(woodType, '') ILIKE $%d ESCAPE '\' OR
			COALESCE(deliveredTo, '') ILIKE $%d ESCAPE '\' OR
			COALESCE(size, '') ILIKE $%d ESCAPE '\' OR
			COALESCE(thicknesses, '') ILIKE $%d ESCAPE '\'
		)`, p, p, p, p, p, p))
	}
	if s := strings.TrimSpace(values.Get("group")); s != "" {
		lq.Args = append(lq.Args, likePattern(s))
		where = append(where, fmt.Sprintf(`COALESCE("group", '') ILIKE $%d ESCAPE '\'`, len(lq.Args)))
	}
	if s := strings.TrimSpace(values.Get("name")); s != "" {
		lq.Args = append(lq.Args, likePattern(s))
		where = append(where, fmt.Sprintf(`COALESCE(text, '') ILIKE $%d ESCAPE '\'`, len(lq.Args)))
	}
	if s := strings.TrimSpace(values.Get("size")); s != "" {
		lq.Args = append(lq.Args, s)
		where = append(where, fmt.Sprintf(`COALESCE(size, '') = $%d`, len(lq.Args)))
	}
	if s := strings.TrimSpace(values.Get("woodType")); s != "" {
		lq.Args = append(lq.Args, s)
		where = append(where, fmt.Sprintf(`woodType = $%d`, len(lq.Args)))
	}
	switch values.Get("hasPicture") {
	case "":
	case "true":
		where = append(where, `COALESCE(picture, '') <> ''`)
	case "false":
		where = append(where, `COALESCE(picture, '') = ''`)
	default:
		return lq, fmt.Errorf("invalid hasPicture")
	}
	switch values.Get("delivered") {
	case "":
	case "true":
		where = append(where, `TRIM(COALESCE(deliveredTo, '')) <> ''`)
	case "false":
		where = append(where, `TRIM(COALESCE(deliveredTo, '')) = ''`)
	default:
		return lq, fmt.Errorf("invalid delivered")
	}
	switch values.Get("deliveredWithBox") {
	case "":
	case "true":
		lq.Args = append(lq.Args, true)
		where = append(where, fmt.Sprintf(`deliveredWithBox = $%d`, len(lq.Args)))
	case "false":
		lq.Args = append(lq.Args, false)
		where = append(where, fmt.Sprintf(`deliveredWithBox = $%d`, len(lq.Args)))
	default:
		return lq, fmt.Errorf("invalid deliveredWithBox")
	}
	if len(where) > 0 {
		lq.WhereSQL = strings.Join(where, " AND ")
	}
	return lq, nil
}

func likePattern(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	return "%" + s + "%"
}

func listItems(w http.ResponseWriter, r *http.Request) {
	lq, err := parseListQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items WHERE `+lq.WhereSQL, lq.Args...).Scan(&total); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + lq.PageSize - 1) / lq.PageSize
	}
	args := append(lq.Args, lq.PageSize, (lq.Page-1)*lq.PageSize)
	query := fmt.Sprintf(`SELECT %s FROM items WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		itemSelect, lq.WhereSQL, lq.OrderSQL, len(args)-1, len(args))
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, itemPage{
		Items:      items,
		Page:       lq.Page,
		PageSize:   lq.PageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

func optionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	woodTypes, err := distinctValues(`SELECT DISTINCT woodType FROM items WHERE COALESCE(woodType, '') <> '' ORDER BY 1`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	groups, err := distinctValues(`SELECT DISTINCT "group" FROM items WHERE COALESCE("group", '') <> '' ORDER BY 1`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sizes, err := distinctValues(`SELECT size FROM items WHERE COALESCE(size, '') <> '' GROUP BY size ORDER BY CASE WHEN size ~ '^[0-9]+([.][0-9]+)?$' THEN size::numeric END NULLS LAST, size`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string][]string{"woodTypes": woodTypes, "groups": groups, "sizes": sizes})
}

func distinctValues(query string) ([]string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(s scanner) (Item, error) {
	var it Item
	err := s.Scan(&it.ID, &it.DateCreated, &it.Text, &it.Size, &it.Thicknesses, &it.WoodType, &it.DeliveryDate, &it.DeliveredTo, &it.Group, &it.Price, &it.DeliveredWithBox, &it.Picture)
	return it, err
}

func loadItem(id int) (Item, error) {
	return scanItem(db.QueryRow(`SELECT `+itemSelect+` FROM items WHERE id=$1`, id))
}

// itemHandler handles requests for /api/items/{id}
func itemHandler(w http.ResponseWriter, r *http.Request) {
	// expect /api/items/{id} or /api/items/{id}/picture
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/items/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "missing id", 400)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}
	if len(parts) == 2 && parts[1] == "picture" {
		pictureHandler(w, r, id)
		return
	}
	if len(parts) != 1 {
		http.Error(w, "not found", 404)
		return
	}
	switch r.Method {
	case "GET":
		it, err := loadItem(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, it)
	case "PUT":
		if !requireAuth(w, r) {
			return
		}
		var it Item
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		res, err := db.Exec(`UPDATE items SET dateCreated=$1, text=$2, size=$3, thicknesses=$4, woodType=$5, deliveryDate=$6, deliveredTo=$7, "group"=$8, price=$9, deliveredWithBox=$10 WHERE id=$11`, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		n, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if n == 0 {
			http.Error(w, "not found", 404)
			return
		}
		stored, err := loadItem(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, stored)
	case "DELETE":
		if !requireAuth(w, r) {
			return
		}
		var picture string
		err := db.QueryRow(`SELECT COALESCE(picture, '') FROM items WHERE id=$1`, id).Scan(&picture)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		res, err := db.Exec(`DELETE FROM items WHERE id=$1`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		n, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if n == 0 {
			http.Error(w, "not found", 404)
			return
		}
		removePictureFile(picture)
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
	if !requireAuth(w, r) {
		return
	}
	path, items, err := readNomadFile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := insertItems(items, true); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("reloaded %d items from %s", len(items), path)
	writeJSON(w, map[string]string{"status": "imported"})
}

// migrateHandler applies migrations from migrations.sql to the connected database
func migrateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	if !requireAuth(w, r) {
		return
	}
	if err := applyMigrations(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]string{"status": "migrated"})
}

func applyMigrations() error {
	b, err := os.ReadFile(resolveExisting("migrations.sql"))
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if _, err := db.Exec(string(b)); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// seedIfEmpty loads NoMad.json only when the items table has no rows.
// Later starts keep whatever is already stored.
func seedIfEmpty() error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n); err != nil {
		return fmt.Errorf("count items: %w", err)
	}
	if n > 0 {
		log.Printf("database already has %d items; using stored data", n)
		return nil
	}
	path, items, err := readNomadFile()
	if err != nil {
		return err
	}
	if err := insertItems(items, false); err != nil {
		return err
	}
	log.Printf("loaded %d items from %s", len(items), path)
	return nil
}

func readNomadFile() (string, []Item, error) {
	path := resolveNomadFile()
	if path == "" {
		return "", nil, fmt.Errorf("NoMad.json not found (set NOMAD_JSON or place the file next to the repo root)")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("read %s: %w", path, err)
	}
	var raw []nomadItem
	if err := decodeJSON(data, &raw); err != nil {
		return "", nil, fmt.Errorf("parse %s: %w", path, err)
	}
	items := make([]Item, 0, len(raw))
	for _, r := range raw {
		thickness := string(r.Thickness)
		if thickness == "" {
			thickness = string(r.Thicknesses)
		}
		items = append(items, Item{
			ID:               r.ID,
			DateCreated:      r.DateCreated,
			Text:             r.Text,
			Size:             string(r.Size),
			Thicknesses:      thickness,
			WoodType:         r.WoodType,
			DeliveryDate:     r.DeliveryDate,
			DeliveredTo:      r.DeliveredTo,
			Group:            r.Group,
			Price:            r.Price,
			DeliveredWithBox: r.DeliveredWithBox,
		})
	}
	return path, items, nil
}

func insertItems(items []Item, replace bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if replace {
		if _, err := tx.Exec(`TRUNCATE items RESTART IDENTITY`); err != nil {
			return err
		}
	}
	stmt, err := tx.Prepare(`INSERT INTO items(id, dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, it := range items {
		if _, err := stmt.Exec(it.ID, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox); err != nil {
			return fmt.Errorf("insert id %d: %w", it.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if replace {
		if err := clearPictures(); err != nil {
			return err
		}
	}
	return syncIDSequence()
}

// syncIDSequence points the id sequence at the current maximum so the next
// insert does not collide with ids loaded from NoMad.json.
func syncIDSequence() error {
	_, err := db.Exec(`
		SELECT setval(
			pg_get_serial_sequence('items', 'id'),
			GREATEST(COALESCE((SELECT MAX(id) FROM items), 1), 1),
			COALESCE((SELECT MAX(id) FROM items), 0) > 0
		)`)
	if err != nil {
		return fmt.Errorf("sync id sequence: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

type nomadItem struct {
	ID               int        `json:"id"`
	DateCreated      string     `json:"dateCreated"`
	Text             string     `json:"text"`
	Size             flexString `json:"size"`
	Thickness        flexString `json:"thickness"`
	Thicknesses      flexString `json:"thicknesses"`
	WoodType         string     `json:"woodType"`
	DeliveryDate     string     `json:"deliveryDate"`
	DeliveredTo      string     `json:"deliveredTo"`
	Group            string     `json:"group"`
	Price            int        `json:"price"`
	DeliveredWithBox bool       `json:"deliveredWithBox"`
}

// flexString accepts JSON strings or numbers, which NoMad.json mixes for size.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexString(s)
		return nil
	}
	*f = flexString(b)
	return nil
}

func decodeJSON(data []byte, out interface{}) error {
	payload, err := toUTF8(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, out)
}

func toUTF8(data []byte) ([]byte, error) {
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return utf16ToUTF8(data[2:], binary.LittleEndian), nil
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return utf16ToUTF8(data[2:], binary.BigEndian), nil
	}
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:], nil
	}
	return data, nil
}

func utf16ToUTF8(data []byte, order binary.ByteOrder) []byte {
	if len(data)%2 == 1 {
		data = data[:len(data)-1]
	}
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = order.Uint16(data[i*2:])
	}
	return []byte(string(utf16.Decode(units)))
}

func resolveNomadFile() string {
	var candidates []string
	if p := os.Getenv("NOMAD_JSON"); p != "" {
		candidates = append(candidates, p)
	}
	candidates = append(candidates,
		"/NoMad.json",
		filepath.Join("..", "NoMad.json"),
		"NoMad.json",
	)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(dir, "NoMad.json"), filepath.Join(dir, "..", "NoMad.json"))
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() || st.Size() == 0 {
			continue
		}
		return p
	}
	return ""
}

func resolveExisting(name string) string {
	candidates := []string{name}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), name))
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err == nil && !st.IsDir() {
			return p
		}
	}
	return name
}
