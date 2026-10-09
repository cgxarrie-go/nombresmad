package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"nombresmad/backend/internal/model"
)

const itemSelect = `id, COALESCE(dateCreated, ''), COALESCE(text, ''), COALESCE(size, ''), COALESCE(thicknesses, ''), COALESCE(woodType, ''), COALESCE(deliveryDate, ''), COALESCE(deliveredTo, ''), COALESCE("group", ''), COALESCE(price, 0), COALESCE(deliveredWithBox, false), COALESCE(picture, '')`

// ItemRepo stores name plates in Postgres.
type ItemRepo struct {
	db *sql.DB
}

// NewItemRepo uses the database handle supplied by the caller.
func NewItemRepo(db *sql.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(s scanner) (model.Item, error) {
	var it model.Item
	err := s.Scan(&it.ID, &it.DateCreated, &it.Text, &it.Size, &it.Thicknesses, &it.WoodType, &it.DeliveryDate, &it.DeliveredTo, &it.Group, &it.Price, &it.DeliveredWithBox, &it.Picture)
	return it, err
}

// List returns one filtered page of items.
func (r *ItemRepo) List(values url.Values) (model.ItemPage, error) {
	lq, err := parseListQuery(values)
	if err != nil {
		return model.ItemPage{}, err
	}
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM items WHERE `+lq.WhereSQL, lq.Args...).Scan(&total); err != nil {
		return model.ItemPage{}, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + lq.PageSize - 1) / lq.PageSize
	}
	args := append(lq.Args, lq.PageSize, (lq.Page-1)*lq.PageSize)
	query := fmt.Sprintf(`SELECT %s FROM items WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		itemSelect, lq.WhereSQL, lq.OrderSQL, len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return model.ItemPage{}, err
	}
	defer rows.Close()
	items := []model.Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return model.ItemPage{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return model.ItemPage{}, err
	}
	return model.ItemPage{
		Items:      items,
		Page:       lq.Page,
		PageSize:   lq.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// Get loads one item by id.
func (r *ItemRepo) Get(id int) (model.Item, error) {
	it, err := scanItem(r.db.QueryRow(`SELECT `+itemSelect+` FROM items WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Item{}, model.ErrNotFound
	}
	return it, err
}

// Create inserts an item and returns it with the new id.
func (r *ItemRepo) Create(it model.Item) (model.Item, error) {
	err := r.db.QueryRow(`INSERT INTO items(dateCreated, text, size, thicknesses, woodType, deliveryDate, deliveredTo, "group", price, deliveredWithBox) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox).Scan(&it.ID)
	if err != nil {
		return model.Item{}, err
	}
	return it, nil
}

// Update writes editable fields and returns the stored row.
func (r *ItemRepo) Update(id int, it model.Item) (model.Item, error) {
	res, err := r.db.Exec(`UPDATE items SET dateCreated=$1, text=$2, size=$3, thicknesses=$4, woodType=$5, deliveryDate=$6, deliveredTo=$7, "group"=$8, price=$9, deliveredWithBox=$10 WHERE id=$11`, it.DateCreated, it.Text, it.Size, it.Thicknesses, it.WoodType, it.DeliveryDate, it.DeliveredTo, it.Group, it.Price, it.DeliveredWithBox, id)
	if err != nil {
		return model.Item{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return model.Item{}, err
	}
	if n == 0 {
		return model.Item{}, model.ErrNotFound
	}
	return r.Get(id)
}

// Delete removes the row and returns the picture file name it held.
func (r *ItemRepo) Delete(id int) (string, error) {
	var picture string
	err := r.db.QueryRow(`SELECT COALESCE(picture, '') FROM items WHERE id=$1`, id).Scan(&picture)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	res, err := r.db.Exec(`DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return "", err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", model.ErrNotFound
	}
	return picture, nil
}

// SetPicture stores the picture file name for an item.
func (r *ItemRepo) SetPicture(id int, name string) error {
	res, err := r.db.Exec(`UPDATE items SET picture=$1 WHERE id=$2`, name, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// Options lists the distinct filter values shown in the UI.
func (r *ItemRepo) Options() (woodTypes, groups, sizes []string, err error) {
	woodTypes, err = r.distinct(`SELECT DISTINCT woodType FROM items WHERE COALESCE(woodType, '') <> '' ORDER BY 1`)
	if err != nil {
		return nil, nil, nil, err
	}
	groups, err = r.distinct(`SELECT DISTINCT "group" FROM items WHERE COALESCE("group", '') <> '' ORDER BY 1`)
	if err != nil {
		return nil, nil, nil, err
	}
	sizes, err = r.distinct(`SELECT size FROM items WHERE COALESCE(size, '') <> '' GROUP BY size ORDER BY CASE WHEN size ~ '^[0-9]+([.][0-9]+)?$' THEN size::numeric END NULLS LAST, size`)
	if err != nil {
		return nil, nil, nil, err
	}
	return woodTypes, groups, sizes, nil
}

func (r *ItemRepo) distinct(query string) ([]string, error) {
	rows, err := r.db.Query(query)
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

// Count returns the number of stored items.
func (r *ItemRepo) Count() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n)
	return n, err
}

// Insert writes items in one transaction. replace truncates the table first.
func (r *ItemRepo) Insert(items []model.Item, replace bool) error {
	tx, err := r.db.Begin()
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
	return tx.Commit()
}

// Exec runs a SQL script, used for migrations.
func (r *ItemRepo) Exec(query string) error {
	_, err := r.db.Exec(query)
	return err
}

// SyncIDSequence points the id sequence at the current maximum so the next
// insert does not collide with ids loaded from initial_load.json.
func (r *ItemRepo) SyncIDSequence() error {
	_, err := r.db.Exec(`
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
