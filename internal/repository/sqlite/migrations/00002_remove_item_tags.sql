-- +goose Up
-- +goose StatementBegin
DROP TRIGGER IF EXISTS items_au;
DROP TRIGGER IF EXISTS items_ad;
DROP TRIGGER IF EXISTS items_ai;

DROP TABLE IF EXISTS items_fts;

ALTER TABLE items DROP COLUMN tags;

CREATE VIRTUAL TABLE items_fts USING fts5(
    name,
    description,
    additional_data,
    content='items',
    content_rowid='rowid'
);

CREATE TRIGGER items_ai AFTER INSERT ON items BEGIN
    INSERT INTO items_fts(rowid, name, description, additional_data)
    VALUES (new.rowid, new.name, new.description, new.additional_data);
END;

CREATE TRIGGER items_ad AFTER DELETE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, additional_data)
    VALUES ('delete', old.rowid, old.name, old.description, old.additional_data);
END;

CREATE TRIGGER items_au AFTER UPDATE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, additional_data)
    VALUES ('delete', old.rowid, old.name, old.description, old.additional_data);
    INSERT INTO items_fts(rowid, name, description, additional_data)
    VALUES (new.rowid, new.name, new.description, new.additional_data);
END;

INSERT INTO items_fts(items_fts) VALUES('rebuild');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS items_au;
DROP TRIGGER IF EXISTS items_ad;
DROP TRIGGER IF EXISTS items_ai;

DROP TABLE IF EXISTS items_fts;

ALTER TABLE items ADD COLUMN tags TEXT NOT NULL DEFAULT '';

CREATE VIRTUAL TABLE items_fts USING fts5(
    name,
    description,
    tags,
    additional_data,
    content='items',
    content_rowid='rowid'
);

CREATE TRIGGER items_ai AFTER INSERT ON items BEGIN
    INSERT INTO items_fts(rowid, name, description, tags, additional_data)
    VALUES (new.rowid, new.name, new.description, new.tags, new.additional_data);
END;

CREATE TRIGGER items_ad AFTER DELETE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, tags, additional_data)
    VALUES ('delete', old.rowid, old.name, old.description, old.tags, old.additional_data);
END;

CREATE TRIGGER items_au AFTER UPDATE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, tags, additional_data)
    VALUES ('delete', old.rowid, old.name, old.description, old.tags, old.additional_data);
    INSERT INTO items_fts(rowid, name, description, tags, additional_data)
    VALUES (new.rowid, new.name, new.description, new.tags, new.additional_data);
END;

INSERT INTO items_fts(items_fts) VALUES('rebuild');
-- +goose StatementEnd
