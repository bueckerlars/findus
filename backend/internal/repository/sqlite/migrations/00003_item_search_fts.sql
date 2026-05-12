-- +goose Up
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN search_labels TEXT NOT NULL DEFAULT '';
ALTER TABLE items ADD COLUMN search_location TEXT NOT NULL DEFAULT '';
ALTER TABLE items ADD COLUMN search_body TEXT NOT NULL DEFAULT '';

DROP TRIGGER IF EXISTS items_au;
DROP TRIGGER IF EXISTS items_ad;
DROP TRIGGER IF EXISTS items_ai;

DROP TABLE IF EXISTS items_fts_substr;
DROP TABLE IF EXISTS items_fts;

CREATE VIRTUAL TABLE items_fts USING fts5(
    name,
    description,
    template_data,
    additional_data,
    search_labels,
    search_location,
    template_type,
    tokenize = 'unicode61 remove_diacritics 2',
    content='items',
    content_rowid='rowid'
);

CREATE TRIGGER items_ai AFTER INSERT ON items BEGIN
    INSERT INTO items_fts(rowid, name, description, template_data, additional_data, search_labels, search_location, template_type)
    VALUES (new.rowid, new.name, new.description, new.template_data, new.additional_data, new.search_labels, new.search_location, new.template_type);
END;

CREATE TRIGGER items_ad AFTER DELETE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, template_data, additional_data, search_labels, search_location, template_type)
    VALUES ('delete', old.rowid, old.name, old.description, old.template_data, old.additional_data, old.search_labels, old.search_location, old.template_type);
END;

CREATE TRIGGER items_au AFTER UPDATE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, name, description, template_data, additional_data, search_labels, search_location, template_type)
    VALUES ('delete', old.rowid, old.name, old.description, old.template_data, old.additional_data, old.search_labels, old.search_location, old.template_type);
    INSERT INTO items_fts(rowid, name, description, template_data, additional_data, search_labels, search_location, template_type)
    VALUES (new.rowid, new.name, new.description, new.template_data, new.additional_data, new.search_labels, new.search_location, new.template_type);
END;

INSERT INTO items_fts(items_fts) VALUES('rebuild');

-- Trigram substring index (no SQLite triggers: UPDATE triggers on trigram FTS are unreliable with rowid sync).
-- Rows are maintained from Go via upsertSubstrFTS (see UpdateItemSearchDenorm and Delete).
CREATE VIRTUAL TABLE items_fts_substr USING fts5(
    body,
    tokenize = 'trigram'
);

CREATE TRIGGER items_substr_ai AFTER INSERT ON items BEGIN
    INSERT INTO items_fts_substr(rowid, body) VALUES (
        new.rowid,
        new.name || ' ' || new.description || ' ' || new.template_data || ' ' || new.additional_data || ' '
            || new.search_labels || ' ' || new.search_location || ' ' || new.template_type || ' ' || new.search_body
    );
END;

INSERT INTO items_fts_substr(rowid, body)
SELECT rowid,
    name || ' ' || description || ' ' || template_data || ' ' || additional_data || ' '
        || search_labels || ' ' || search_location || ' ' || template_type || ' ' || search_body
FROM items;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS items_substr_ai;
DROP TABLE IF EXISTS items_fts_substr;

DROP TRIGGER IF EXISTS items_au;
DROP TRIGGER IF EXISTS items_ad;
DROP TRIGGER IF EXISTS items_ai;
DROP TABLE IF EXISTS items_fts;

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

-- SQLite cannot DROP COLUMN easily in older versions; recreate items without search_* columns is heavy.
-- Down migration: leave search_* columns on items (harmless) for simplicity, or use recreate.
-- Goose down: we drop only FTS rebuild state; keep columns for compatibility with app expecting them.
-- +goose StatementEnd
