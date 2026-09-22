CREATE TABLE IF NOT EXISTS items (
    id integer PRIMARY KEY,
    dateCreated text,
    text text,
    size text,
    thicknesses text,
    woodType text,
    deliveryDate text,
    deliveredTo text,
    "group" text,
    price integer,
    deliveredWithBox boolean
);
