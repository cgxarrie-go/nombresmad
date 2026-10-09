package service

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf16"

	"nombresmad/backend/internal/model"
)

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

func readNomadFile() (string, []model.Item, error) {
	path := resolveNomadFile()
	if path == "" {
		return "", nil, fmt.Errorf("initial_load.json not found (set NOMAD_JSON or place the file in backend/seed_data)")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("read %s: %w", path, err)
	}
	var raw []nomadItem
	if err := decodeJSON(data, &raw); err != nil {
		return "", nil, fmt.Errorf("parse %s: %w", path, err)
	}
	items := make([]model.Item, 0, len(raw))
	for _, r := range raw {
		thickness := string(r.Thickness)
		if thickness == "" {
			thickness = string(r.Thicknesses)
		}
		items = append(items, model.Item{
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

func readMigrations() ([]byte, error) {
	return os.ReadFile(resolveExisting("migrations.sql"))
}

func decodeJSON(data []byte, out any) error {
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
		filepath.Join("seed_data", "initial_load.json"),
		filepath.Join("backend", "seed_data", "initial_load.json"),
		filepath.Join("..", "seed_data", "initial_load.json"),
		filepath.Join("..", "..", "seed_data", "initial_load.json"),
	)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "seed_data", "initial_load.json"),
			filepath.Join(dir, "..", "seed_data", "initial_load.json"),
		)
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
