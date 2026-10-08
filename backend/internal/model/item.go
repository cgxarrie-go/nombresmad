package model

import "errors"

// ErrNotFound is returned when an item row does not exist.
var ErrNotFound = errors.New("not found")

// InvalidError is a rejected request that did not touch storage.
type InvalidError struct {
	msg string
}

func (e *InvalidError) Error() string { return e.msg }

// Invalid marks a client input error.
func Invalid(msg string) error { return &InvalidError{msg: msg} }

// Item is a wooden name plate.
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

// ItemPage is one page of items.
type ItemPage struct {
	Items      []Item `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
}
