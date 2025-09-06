-- Update forms table to match ContactForm structure from SDD
-- Remove old columns and add new ones

BEGIN TRANSACTION;

-- Create a temporary table with the new structure
CREATE TABLE forms_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    turnstile_secret TEXT NOT NULL,
    forward_email TEXT NOT NULL,
    form_key TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE(user_id, name),
    UNIQUE(form_key)
);

-- Copy data from old table to new table (with default values for new columns)
INSERT INTO forms_new (id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at)
SELECT 
    id, 
    user_id, 
    name,
    'example.com' as domain,
    'turnstile_secret_' || id as turnstile_secret,
    'admin@example.com' as forward_email,
    'form_key_' || id as form_key,
    created_at,
    updated_at
FROM forms;

-- Drop the old table
DROP TABLE forms;

-- Rename the new table to the original name
ALTER TABLE forms_new RENAME TO forms;

-- Recreate indexes
CREATE INDEX idx_forms_user_id ON forms(user_id);
CREATE INDEX idx_forms_form_key ON forms(form_key);

COMMIT;