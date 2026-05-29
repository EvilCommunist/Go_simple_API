-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT NULL REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(parent_id, name)
);

CREATE INDEX idx_departments_parent ON departments(parent_id);

-- +goose Down
DROP TABLE departments;