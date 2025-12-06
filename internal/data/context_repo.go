package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"whatsapp-analytics-mvp/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteContextRepo implements ContextManager.
type SQLiteContextRepo struct {
	DB *sql.DB
}

// -----------------------------------------------------------------------------
// INIT
// -----------------------------------------------------------------------------

func NewSQLiteContextRepo(dbPath string) (*SQLiteContextRepo, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		client_id TEXT PRIMARY KEY,
		name TEXT,
		lang TEXT DEFAULT 'ru',
		loyalty_level TEXT DEFAULT 'Standard',
		total_spent REAL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		sender TEXT NOT NULL,
		message_text TEXT NOT NULL,
		FOREIGN KEY(client_id) REFERENCES clients(client_id)
	);

	CREATE TABLE IF NOT EXISTS bookings (
		booking_id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL,
		booking_start TIMESTAMP NOT NULL,
		seats INTEGER DEFAULT 1,
		hours INTEGER DEFAULT 1,
		amount REAL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(client_id) REFERENCES clients(client_id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		session_id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id TEXT NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		booking_id TEXT,
		FOREIGN KEY(client_id) REFERENCES clients(client_id),
		FOREIGN KEY(booking_id) REFERENCES bookings(booking_id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_client_time ON messages(client_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_bookings_start ON bookings(booking_start);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &SQLiteContextRepo{DB: db}, nil
}

// -----------------------------------------------------------------------------
// SAVE MESSAGE
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) SaveMessage(ctx context.Context, clientID, sender, text string) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO messages (client_id, sender, message_text) VALUES (?, ?, ?)`,
		clientID, sender, text,
	)
	return err
}

// -----------------------------------------------------------------------------
// SAVE BOOKING  (новая функция для ToolsService)
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) SaveBooking(
	ctx context.Context,
	bookingID, clientID string,
	start time.Time,
	seats, hours int,
	amountStr string,
) error {

	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO bookings (booking_id, client_id, booking_start, seats, hours, amount)
		VALUES (?, ?, ?, ?, ?, ?)
	`, bookingID, clientID, start, seats, hours, amount)

	return err
}

// -----------------------------------------------------------------------------
// GET BOOKINGS AT A SPECIFIC TIME (новая функция для ToolsService)
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) GetBookingsAt(ctx context.Context, t time.Time) ([]models.Booking, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT booking_id, client_id, booking_start, seats, hours, amount
		FROM bookings
		WHERE booking_start = ?
	`, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Booking

	for rows.Next() {
		var b models.Booking
		err := rows.Scan(&b.BookingID, &b.ClientID, &b.Start, &b.Seats, &b.Hours, &b.Amount)
		if err != nil {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

// -----------------------------------------------------------------------------
// GET PROFILE
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) GetProfile(ctx context.Context, clientID string) (*models.ClientProfile, error) {
	var name, lang, loyalty string
	var spent float64

	err := r.DB.QueryRowContext(ctx, `
		SELECT name, lang, loyalty_level, total_spent
		FROM clients
		WHERE client_id = ?
	`, clientID).Scan(&name, &lang, &loyalty, &spent)

	if err == sql.ErrNoRows {
		_, err = r.DB.ExecContext(ctx,
			`INSERT INTO clients (client_id, name) VALUES (?, ?)`,
			clientID, "Client",
		)
		if err != nil {
			return nil, err
		}
		name = "Client"
		lang = "ru"
		loyalty = "Standard"
	} else if err != nil {
		return nil, err
	}

	history, err := r.getRelevantHistory(ctx, clientID)
	if err != nil {
		return nil, err
	}

	return &models.ClientProfile{
		ClientID:     clientID,
		Name:         name,
		Lang:         lang,
		LoyaltyLevel: loyalty,
		TotalSpent:   spent,
		History:      history,
	}, nil
}

// -----------------------------------------------------------------------------
// HISTORY SELECTION
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) getRelevantHistory(ctx context.Context, clientID string) (string, error) {
	cutoff := time.Now().Add(-2 * time.Hour)

	rows, err := r.DB.QueryContext(ctx, `
		SELECT timestamp, sender, message_text
		FROM messages
		WHERE client_id = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`, clientID, cutoff)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type hist struct {
		Time   time.Time `json:"time"`
		Sender string    `json:"sender"`
		Text   string    `json:"text"`
	}

	var entries []hist
	for rows.Next() {
		var h hist
		if err := rows.Scan(&h.Time, &h.Sender, &h.Text); err != nil {
			continue
		}
		entries = append(entries, h)
	}

	out, _ := json.MarshalIndent(entries, "", "  ")
	return string(out), nil
}

// -----------------------------------------------------------------------------
// CREATE OR UPDATE SESSION
// -----------------------------------------------------------------------------

// GetChatHistory — возвращает историю сообщений клиента (последние 24 часа).
func (r *SQLiteContextRepo) GetChatHistory(ctx context.Context, clientID string) ([]map[string]string, error) {
	cutoff := time.Now().Add(-24 * time.Hour)

	rows, err := r.DB.QueryContext(ctx, `
		SELECT timestamp, sender, message_text
		FROM messages
		WHERE client_id = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`, clientID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]string
	for rows.Next() {
		var ts time.Time
		var sender, text string
		if err := rows.Scan(&ts, &sender, &text); err != nil {
			continue
		}
		out = append(out, map[string]string{
			"timestamp": ts.Format("2006-01-02 15:04:05"),
			"sender":    sender,
			"text":      text,
		})
	}
	return out, nil
}

// CreateOrUpdateSession — создаёт или обновляет сессию клиента.
func (r *SQLiteContextRepo) CreateOrUpdateSession(
	ctx context.Context,
	clientID string,
	bookingID *string,
) error {
	// MVP stab: можно просто логировать или игнорировать
	// В будущем здесь будет реальная логика работы с таблицей sessions
	if bookingID != nil {
		_, err := r.DB.ExecContext(ctx, `
			INSERT OR REPLACE INTO sessions (client_id, booking_id, started_at, expires_at)
			VALUES (?, ?, CURRENT_TIMESTAMP, datetime('now', '+24 hours'))
		`, clientID, *bookingID)
		return err
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT OR REPLACE INTO sessions (client_id, started_at, expires_at)
		VALUES (?, CURRENT_TIMESTAMP, datetime('now', '+24 hours'))
	`, clientID)
	return err
}

// -----------------------------------------------------------------------------
// BUSINESS SETTINGS
// -----------------------------------------------------------------------------

func (r *SQLiteContextRepo) GetBusinessSettings(ctx context.Context) (models.BusinessSettings, error) {
	return models.BusinessSettings{
		BusinessName: "Team Racing Club",
		Address:      "г.Астана, пр.Абылай хана 27/4",
		WorkingHours: "12:00–04:00",
	}, nil
}
