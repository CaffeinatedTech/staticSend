package models

import (
	"database/sql"
	"time"
)

// SubmissionEmail represents email tracking for a submission
type SubmissionEmail struct {
	ID            int64      `json:"id"`
	SubmissionID  int64      `json:"submission_id"`
	SentAt        time.Time  `json:"sent_at"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"error_message"`
}

// CreateSubmissionEmail creates a new email tracking record
func CreateSubmissionEmail(db *sql.DB, submissionID int64, status, errorMessage string) (*SubmissionEmail, error) {
	result, err := db.Exec(
		"INSERT INTO submission_emails (submission_id, status, error_message) VALUES (?, ?, ?)",
		submissionID, status, errorMessage,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetSubmissionEmailByID(db, id)
}

// GetSubmissionEmailByID retrieves an email record by its ID
func GetSubmissionEmailByID(db *sql.DB, id int64) (*SubmissionEmail, error) {
	var email SubmissionEmail
	err := db.QueryRow(
		"SELECT id, submission_id, sent_at, status, error_message FROM submission_emails WHERE id = ?",
		id,
	).Scan(&email.ID, &email.SubmissionID, &email.SentAt, &email.Status, &email.ErrorMessage)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &email, nil
}

// GetSubmissionEmailBySubmissionID retrieves the email record for a specific submission
func GetSubmissionEmailBySubmissionID(db *sql.DB, submissionID int64) (*SubmissionEmail, error) {
	var email SubmissionEmail
	err := db.QueryRow(
		"SELECT id, submission_id, sent_at, status, error_message FROM submission_emails WHERE submission_id = ?",
		submissionID,
	).Scan(&email.ID, &email.SubmissionID, &email.SentAt, &email.Status, &email.ErrorMessage)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &email, nil
}

// UpdateSubmissionEmailStatus updates the status of an email record
func UpdateSubmissionEmailStatus(db *sql.DB, id int64, status, errorMessage string) error {
	_, err := db.Exec(
		"UPDATE submission_emails SET status = ?, error_message = ? WHERE id = ?",
		status, errorMessage, id,
	)
	return err
}