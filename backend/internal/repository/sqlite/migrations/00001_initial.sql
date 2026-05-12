-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    avatar_path TEXT
);

CREATE TABLE locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id TEXT REFERENCES locations(id) ON DELETE RESTRICT,
    description TEXT NOT NULL DEFAULT '',
    qr_token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_locations_parent ON locations(parent_id);

INSERT INTO locations (id, name, parent_id, description, qr_token, created_at, updated_at) VALUES
('loc_basement', 'Basement', NULL, '', 'seed00000001locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('loc_office', 'Office', NULL, '', 'seed00000002locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('loc_livingroom', 'Livingroom', NULL, '', 'seed00000003locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('loc_bedroom', 'Bedroom', NULL, '', 'seed00000004locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('loc_bathroom', 'Bathroom', NULL, '', 'seed00000005locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('loc_garage', 'Garage', NULL, '', 'seed00000006locqr', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z');

CREATE TABLE items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    location_id TEXT NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    template_type TEXT NOT NULL,
    template_data TEXT NOT NULL DEFAULT '{}',
    additional_data TEXT NOT NULL DEFAULT '{}',
    photo_path TEXT,
    tags TEXT NOT NULL DEFAULT '',
    qr_token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_items_location ON items(location_id);
CREATE INDEX idx_items_template ON items(template_type);

CREATE TABLE item_templates (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    fields_json TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_item_templates_sort ON item_templates(sort_order, id);

INSERT INTO item_templates (id, display_name, fields_json, sort_order, created_at, updated_at) VALUES
('standard', 'Standard', '[]', 0, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('electronics', 'Electronics', '[{"key":"condition","label":"Condition","widget":"select","required":true,"options":[{"value":"working","label":"Working"},{"value":"broken","label":"Broken"}]},{"key":"power_cable","label":"Power cable needed?","widget":"select","required":true,"options":[{"value":"yes","label":"Yes"},{"value":"no","label":"No"}]}]', 1, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('clothing', 'Clothing', '[{"key":"season","label":"Season","widget":"select","required":true,"options":[{"value":"summer","label":"Summer"},{"value":"winter","label":"Winter"},{"value":"transition","label":"Transition"}]},{"key":"size","label":"Size","widget":"text","required":true,"placeholder":"S, M, L…","max_len":32}]', 2, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('documents', 'Documents', '[{"key":"year","label":"Year (YYYY)","widget":"text","required":true,"placeholder":"2024","pattern":"^\\d{4}$","min_int":1900,"max_int":2100},{"key":"category","label":"Category","widget":"select","required":true,"options":[{"value":"tax","label":"Tax"},{"value":"contracts","label":"Contracts"},{"value":"invoices","label":"Invoices"}]}]', 3, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('servers', 'Servers', '[{"key":"hostname","label":"Hostname","widget":"text","required":true,"placeholder":"db-01.example.com","max_len":253},{"key":"server_role","label":"Role","widget":"select","required":true,"options":[{"value":"physical","label":"Physical"},{"value":"virtual_machine","label":"Virtual machine"},{"value":"container_host","label":"Container host"},{"value":"hypervisor","label":"Hypervisor"}]},{"key":"os_family","label":"OS family","widget":"select","required":true,"options":[{"value":"linux","label":"Linux"},{"value":"windows","label":"Windows"},{"value":"bsd","label":"BSD"},{"value":"other","label":"Other"}]},{"key":"rack_slot","label":"Rack / slot","widget":"text","required":false,"placeholder":"A-04 U22","max_len":64}]', 4, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('iot', 'IoT', '[{"key":"device_type","label":"Device type","widget":"select","required":true,"options":[{"value":"sensor","label":"Sensor"},{"value":"actuator","label":"Actuator"},{"value":"gateway","label":"Gateway"},{"value":"module","label":"Module"},{"value":"other","label":"Other"}]},{"key":"connectivity","label":"Connectivity","widget":"select","required":true,"options":[{"value":"wifi","label":"Wi-Fi"},{"value":"ethernet","label":"Ethernet"},{"value":"cellular","label":"Cellular"},{"value":"ble","label":"Bluetooth LE"},{"value":"zigbee","label":"Zigbee"},{"value":"thread","label":"Thread"},{"value":"lora","label":"LoRa"},{"value":"other","label":"Other"}]},{"key":"firmware","label":"Firmware","widget":"text","required":false,"placeholder":"e.g. 3.2.0","max_len":64}]', 5, '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z');

CREATE TABLE labels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#64748b',
    default_template_type TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

INSERT INTO labels (id, name, color, default_template_type, created_at, updated_at) VALUES
('lbl_standard', 'Standard', '#64748b', 'standard', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('lbl_electronics', 'Electronics', '#2563eb', 'electronics', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('lbl_clothing', 'Clothing', '#db2777', 'clothing', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('lbl_documents', 'Documents', '#ca8a04', 'documents', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('lbl_servers', 'Servers', '#059669', 'servers', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z'),
('lbl_iot', 'IoT', '#7c3aed', 'iot', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z');

CREATE TABLE item_labels (
    item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    label_id TEXT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, label_id)
);
CREATE INDEX idx_item_labels_label ON item_labels(label_id);

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

CREATE TABLE invites (
    id TEXT PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS invites;
DROP TRIGGER IF EXISTS items_au;
DROP TRIGGER IF EXISTS items_ad;
DROP TRIGGER IF EXISTS items_ai;
DROP TABLE IF EXISTS items_fts;
DROP TABLE IF EXISTS item_labels;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS item_templates;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
