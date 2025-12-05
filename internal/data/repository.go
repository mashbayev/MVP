package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"whatsapp-analytics-mvp/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite3 database and creates necessary tables.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS dialog_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT,
			timestamp TEXT,
			message_text TEXT,
			intent TEXT,
			lead_source TEXT,
			sentiment TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS client_profiles (
			client_id TEXT PRIMARY KEY,
			first_contact TEXT,
			last_contact TEXT,
			admin_assigned_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS business_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			business_name TEXT,
			address TEXT,
			working_hours TEXT
		);`,
		// Insert default settings if table is empty
		`INSERT OR IGNORE INTO business_settings (id, business_name, address, working_hours)
		 VALUES (1, 'My Business', '123 Main St', '9:00-18:00');`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}

// SQLiteRepo implements core.AnalyticsAdapter and core.SettingsRepository.
type SQLiteRepo struct {
	DB *sql.DB
}

// SaveLog saves a DialogLog entry to the database.
func (r *SQLiteRepo) SaveLog(entry models.DialogLog) error {
	query := `INSERT INTO dialog_logs (client_id, timestamp, message_text, intent, lead_source, sentiment)
              VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.DB.Exec(query,
		entry.ClientID,
		entry.Timestamp.Format(time.RFC3339),
		entry.MessageText,
		entry.Intent,
		entry.LeadSource,
		entry.Sentiment,
	)
	if err != nil {
		log.Printf("Error saving log: %v", err)
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

// GetBusinessSettings retrieves the AI settings from the database.
func (r *SQLiteRepo) GetBusinessSettings() (models.BusinessSettings, error) {
	query := `SELECT business_name, address, working_hours FROM business_settings WHERE id = 1 LIMIT 1`

	var settings models.BusinessSettings
	err := r.DB.QueryRow(query).Scan(&settings.BusinessName, &settings.Address, &settings.WorkingHours)
	if err == sql.ErrNoRows {
		return models.BusinessSettings{
			BusinessName: "My Business",
			Address:      "123 Main St",
			WorkingHours: "9:00-18:00",
		}, nil
	}
	if err != nil {
		return models.BusinessSettings{}, fmt.Errorf("failed to get business settings: %w", err)
	}
	return settings, nil
}
