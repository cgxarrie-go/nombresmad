package repo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"nombresmad/backend/internal/model"
)

const pageSize = 25

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
			return lq, model.Invalid("invalid page")
		}
		lq.Page = n
	}
	if raw := values.Get("pageSize"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 60 {
			return lq, model.Invalid("invalid pageSize")
		}
		lq.PageSize = n
	}
	sortKey := values.Get("sort")
	if sortKey == "" {
		sortKey = "id"
	}
	expr, ok := sortColumns[sortKey]
	if !ok {
		return lq, model.Invalid("invalid sort")
	}
	order := strings.ToLower(values.Get("order"))
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		return lq, model.Invalid("invalid order")
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
		return lq, model.Invalid("invalid hasPicture")
	}
	switch values.Get("delivered") {
	case "":
	case "true":
		where = append(where, `TRIM(COALESCE(deliveredTo, '')) <> ''`)
	case "false":
		where = append(where, `TRIM(COALESCE(deliveredTo, '')) = ''`)
	default:
		return lq, model.Invalid("invalid delivered")
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
		return lq, model.Invalid("invalid deliveredWithBox")
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
