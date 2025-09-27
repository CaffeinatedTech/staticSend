-- Add optional callback_url to forms
ALTER TABLE forms ADD COLUMN callback_url TEXT;
