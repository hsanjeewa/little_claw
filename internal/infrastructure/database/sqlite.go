package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/devops/agent/internal/domain/agent"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to open database: %v", err))
	}

	query := `
	CREATE TABLE IF NOT EXISTS system_audit_logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME,
		host TEXT,
		command TEXT,
		status TEXT,
		output TEXT
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to create schema: %v", err))
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) SaveLog(ctx context.Context, log agent.ExecutionLog) error {
	query := `
	INSERT INTO system_audit_logs (id, timestamp, host, command, status, output)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	
	outputVal := sql.NullString{
		String: log.Output,
		Valid:  log.Output != "",
	}

	_, err := r.db.ExecContext(ctx, query, log.ID, log.Timestamp, log.Host, log.Command, log.Status, outputVal)
	if err != nil {
		return fmt.Errorf("context: %w", fmt.Errorf("failed to insert log: %v", err))
	}

	return nil
}

func (r *SQLiteRepository) GetLogs(ctx context.Context, host string) ([]agent.ExecutionLog, error) {
	query := `
	SELECT id, timestamp, host, command, status, output
	FROM system_audit_logs
	WHERE host = ?
	ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, host)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to query logs: %v", err))
	}
	defer rows.Close()

	var logs []agent.ExecutionLog
	for rows.Next() {
		var log agent.ExecutionLog
		var statusStr string
		var outputVal sql.NullString

		err := rows.Scan(&log.ID, &log.Timestamp, &log.Host, &log.Command, &statusStr, &outputVal)
		if err != nil {
			return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to scan log row: %v", err))
		}

		log.Status = agent.TaskStatus(statusStr)
		if outputVal.Valid {
			log.Output = outputVal.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("error iterating log rows: %v", err))
	}

	return logs, nil
}
