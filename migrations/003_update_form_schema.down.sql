-- Revert forms table to original structure

BEGIN TRANSACTION;

-- Create a temporary table with the old structure
CREATE TABLE forms_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    redirect_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

-- Copy data from new table to old table (with default values for old columns)
INSERT INTO forms_old (id, user_id, name, title, description, redirect_url, created_at, updated_at)
SELECT 
    id, 
    user_id, 
    name,
    name as title,
    '' as description,
    '' as redirect_url,
    created_at,
    updated_at
FROM forms;

-- Drop the new table
DROP TABLE forms;

-- Rename the old table to the original name
ALTER TABLE forms_old RENAME TO forms;

-- Recreate indexes
CREATE INDEX idx_forms_user_id ON forms(user_id);

COMMIT;