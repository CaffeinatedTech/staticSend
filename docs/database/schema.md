# Database Schema Design

## Overview
This document describes the database schema for staticSend, a form submission management system.

## Tables

### users
Stores user accounts for form management
- `id` - Primary key, auto-increment
- `email` - Unique user email
- `password_hash` - Hashed password
- `created_at` - Account creation timestamp
- `updated_at` - Last update timestamp

### forms
Stores contact form configurations
- `id` - Primary key, auto-increment
- `user_id` - Foreign key to users
- `name` - Form name/identifier
- `title` - Display title
- `description` - Form description
- `redirect_url` - URL to redirect after submission
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### submissions
Stores form submissions
- `id` - Primary key, auto-increment
- `form_id` - Foreign key to forms
- `ip_address` - Submitter IP address
- `user_agent` - Browser user agent
- `submitted_data` - JSON blob of form data
- `created_at` - Submission timestamp
- `processed_at` - When email was sent (nullable)
- `status` - Submission status (pending, processed, failed)

### submission_emails
Tracks email sending for submissions
- `id` - Primary key, auto-increment
- `submission_id` - Foreign key to submissions
- `sent_at` - When email was sent
- `status` - Delivery status (sent, failed)
- `error_message` - Error if delivery failed

## Relationships
- One user can have multiple forms
- One form can have multiple submissions
- One submission has one email tracking record

## Indexes
- `users.email` - Unique index for login
- `forms.user_id` - For user form queries
- `submissions.form_id` - For form submission queries
- `submissions.created_at` - For time-based queries
- `submission_emails.submission_id` - For email tracking