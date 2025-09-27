-- Remove callback_url from forms (requires table rebuild in SQLite)
BEGIN TRANSACTION;

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

INSERT INTO forms_new (id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at)
SELECT id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at
FROM forms;

DROP TABLE forms;
ALTER TABLE forms_new RENAME TO forms;

CREATE INDEX idx_forms_user_id ON forms(user_id);
CREATE INDEX idx_forms_form_key ON forms(form_key);

COMMIT;
