package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/devops/agent/internal/domain/agent"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
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
		output TEXT,
		error TEXT
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
	INSERT INTO system_audit_logs (id, timestamp, host, command, status, output, error)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	outputVal := sql.NullString{
		String: log.Output,
		Valid:  log.Output != "",
	}
	
	errorVal := sql.NullString{
		String: log.Error,
		Valid:  log.Error != "",
	}

	_, err := r.db.ExecContext(ctx, query, log.ID, log.Timestamp, log.Host, log.Command, log.Status, outputVal, errorVal)
	if err != nil {
		return fmt.Errorf("context: %w", fmt.Errorf("failed to insert log: %v", err))
	}

	return nil
}

func (r *SQLiteRepository) GetLogs(ctx context.Context, host string) ([]agent.ExecutionLog, error) {
	query := `
	SELECT id, timestamp, host, command, status, output, error
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
		var errorVal sql.NullString

		err := rows.Scan(&log.ID, &log.Timestamp, &log.Host, &log.Command, &statusStr, &outputVal, &errorVal)
		if err != nil {
			return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to scan log row: %v", err))
		}

		log.Status = agent.TaskStatus(statusStr)
		if outputVal.Valid {
			log.Output = outputVal.String
		}
		if errorVal.Valid {
			log.Error = errorVal.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("error iterating log rows: %v", err))
	}

	return logs, nil
}

func (r *SQLiteRepository) GetFailedLogs(ctx context.Context) ([]agent.ExecutionLog, error) {
	query := `
	SELECT id, timestamp, host, command, status, output, error
	FROM system_audit_logs
	WHERE status = 'FAILED'
	ORDER BY timestamp DESC
	LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to query failed logs: %v", err))
	}
	defer rows.Close()

	var logs []agent.ExecutionLog
	for rows.Next() {
		var log agent.ExecutionLog
		var statusStr string
		var outputVal sql.NullString
		var errorVal sql.NullString

		err := rows.Scan(&log.ID, &log.Timestamp, &log.Host, &log.Command, &statusStr, &outputVal, &errorVal)
		if err != nil {
			return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to scan log row: %v", err))
		}

		log.Status = agent.TaskStatus(statusStr)
		if outputVal.Valid {
			log.Output = outputVal.String
		}
		if errorVal.Valid {
			log.Error = errorVal.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("error iterating log rows: %v", err))
	}

	return logs, nil
}
