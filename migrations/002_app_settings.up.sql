-- Add app_settings table for application configuration
CREATE TABLE app_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO app_settings (key, value, description) VALUES
('registration_enabled', 'true', 'Whether new user registration is enabled (true/false)'),
('site_title', 'staticSend', 'The title of the application'),
('site_description', 'A simple contact form service', 'The description of the application');

-- Create index for faster lookups
CREATE INDEX idx_app_settings_key ON app_settings(key);