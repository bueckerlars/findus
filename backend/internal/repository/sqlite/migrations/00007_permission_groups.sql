-- +goose Up
-- +goose StatementBegin
CREATE TABLE perm_groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE perm_group_permissions (
    group_id TEXT NOT NULL REFERENCES perm_groups(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (group_id, permission)
);

CREATE TABLE user_perm_groups (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id TEXT NOT NULL REFERENCES perm_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE INDEX idx_user_perm_groups_group_id ON user_perm_groups(group_id);

INSERT INTO perm_groups (id, name, created_at, updated_at) VALUES
('grp_seed_editors', 'Editors', '2020-01-01T00:00:00Z', '2020-01-01T00:00:00Z');

INSERT INTO perm_group_permissions (group_id, permission) VALUES
('grp_seed_editors', 'items.write'),
('grp_seed_editors', 'labels.write'),
('grp_seed_editors', 'locations.write');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_perm_groups;
DROP TABLE IF EXISTS perm_group_permissions;
DROP TABLE IF EXISTS perm_groups;
-- +goose StatementEnd
