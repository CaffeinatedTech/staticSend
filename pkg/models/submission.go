package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Submission represents a form submission
type Submission struct {
	ID            int64           `json:"id"`
	FormID        int64           `json:"form_id"`
	IPAddress     string          `json:"ip_address"`
	UserAgent     string          `json:"user_agent"`
	SubmittedData json.RawMessage `json:"submitted_data"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at"`
	Status        string          `json:"status"`
}

// CreateSubmission creates a new form submission
func CreateSubmission(db *sql.DB, formID int64, ipAddress, userAgent string, submittedData json.RawMessage) (*Submission, error) {
	result, err := db.Exec(
		"INSERT INTO submissions (form_id, ip_address, user_agent, submitted_data) VALUES (?, ?, ?, ?)",
		formID, ipAddress, userAgent, string(submittedData),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetSubmissionByID(db, id)
}

// GetSubmissionByID retrieves a submission by its ID
func GetSubmissionByID(db *sql.DB, id int64) (*Submission, error) {
	var submission Submission
	var processedAt sql.NullTime
	var submittedData string

	err := db.QueryRow(
		"SELECT id, form_id, ip_address, user_agent, submitted_data, created_at, processed_at, status FROM submissions WHERE id = ?",
		id,
	).Scan(&submission.ID, &submission.FormID, &submission.IPAddress, &submission.UserAgent, &submittedData, &submission.CreatedAt, &processedAt, &submission.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert string back to JSON raw message
	submission.SubmittedData = json.RawMessage(submittedData)

	// Handle nullable processed_at
	if processedAt.Valid {
		submission.ProcessedAt = &processedAt.Time
	}

	return &submission, nil
}

// GetSubmissionsByFormID retrieves all submissions for a specific form
func GetSubmissionsByFormID(db *sql.DB, formID int64) ([]Submission, error) {
	rows, err := db.Query(
		"SELECT id, form_id, ip_address, user_agent, submitted_data, created_at, processed_at, status FROM submissions WHERE form_id = ? ORDER BY created_at DESC",
		formID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var submission Submission
		var processedAt sql.NullTime
		var submittedData string

		if err := rows.Scan(&submission.ID, &submission.FormID, &submission.IPAddress, &submission.UserAgent, &submittedData, &submission.CreatedAt, &processedAt, &submission.Status); err != nil {
			return nil, err
		}

		// Convert string back to JSON raw message
		submission.SubmittedData = json.RawMessage(submittedData)

		// Handle nullable processed_at
		if processedAt.Valid {
			submission.ProcessedAt = &processedAt.Time
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// UpdateSubmissionStatus updates the status and processed_at timestamp of a submission
func UpdateSubmissionStatus(db *sql.DB, id int64, status string) error {
	var processedAt interface{}
	if status == "processed" {
		processedAt = time.Now()
	} else {
		processedAt = nil
	}

	_, err := db.Exec(
		"UPDATE submissions SET status = ?, processed_at = ? WHERE id = ?",
		status, processedAt, id,
	)
	return err
}

// GetSubmissionCountByFormID returns the number of submissions for a form
func GetSubmissionCountByFormID(db *sql.DB, formID int64) (int, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM submissions WHERE form_id = ?",
		formID,
	).Scan(&count)

	return count, err
}