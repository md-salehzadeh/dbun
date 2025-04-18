package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/md-salehzadeh/dbun/src/config"
	"github.com/md-salehzadeh/dbun/src/model"

	_ "github.com/go-sql-driver/mysql"
)

// Manager handles database operations
type Manager struct {
	db     *sql.DB
	config config.DBConfig
}

// NewManager creates a new database manager
func NewManager(config config.DBConfig) (*Manager, error) {
	db, err := initDB(config)
	if err != nil {
		return nil, err
	}

	return &Manager{
		db:     db,
		config: config,
	}, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// Initialize the database connection
func initDB(config config.DBConfig) (*sql.DB, error) {
	// Format DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Database)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

// GetTableNames fetches all table names from the database
func (m *Manager) GetTableNames() ([]string, error) {
	query := "SHOW TABLES"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %v", err)
	}

	return tables, nil
}

// GetTableMetadata fetches column metadata for a specific table
func (m *Manager) GetTableMetadata(tableName string) ([]model.ColumnMetadata, error) {
	query := fmt.Sprintf("DESCRIBE %s", tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching table metadata: %v", err)
	}
	defer rows.Close()

	var columns []model.ColumnMetadata
	for rows.Next() {
		var field, dataType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &dataType, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("error scanning column metadata: %v", err)
		}

		column := model.ColumnMetadata{
			Name:     field.String,
			Type:     dataType.String,
			Nullable: null.String == "YES",
			Key:      key.String,
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %v", err)
	}

	return columns, nil
}

// GetTableIndices fetches indices for a specific table
func (m *Manager) GetTableIndices(tableName string) ([]string, error) {
	query := fmt.Sprintf("SHOW INDEX FROM %s", tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching table indices: %v", err)
	}
	defer rows.Close()

	// Get column names from result set
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting index columns: %v", err)
	}

	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	indexMap := make(map[string]bool) // Use map to avoid duplicates

	// For each row, scan all columns
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, fmt.Errorf("error scanning index data: %v", err)
		}

		// Extract the key name (usually in position 2)
		keyNameIdx := 2 // Default position for key_name

		// Find the key_name column position for flexibility
		for i, colName := range columns {
			if strings.EqualFold(colName, "Key_name") {
				keyNameIdx = i
				break
			}
		}

		// Get the key name if it exists and is not null
		if keyNameIdx < len(values) && values[keyNameIdx] != nil {
			switch v := values[keyNameIdx].(type) {
			case []byte:
				indexMap[string(v)] = true
			case string:
				indexMap[v] = true
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indices: %v", err)
	}

	// Convert map keys to slice
	indices := make([]string, 0, len(indexMap))
	for key := range indexMap {
		indices = append(indices, key)
	}

	return indices, nil
}

// GetTableData fetches data for a specific table (limited to a reasonable number of rows)
func (m *Manager) GetTableData(tableName string, limit int) ([]model.RowData, error) {
	// Get columns first to handle the results properly
	columns, err := m.GetTableMetadata(tableName)
	if err != nil {
		return nil, err
	}

	// Build query with column names with proper backtick escaping for MySQL
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		// Escape column names with backticks to handle reserved words and special characters
		columnNames[i] = fmt.Sprintf("`%s`", col.Name)
	}

	query := fmt.Sprintf("SELECT %s FROM `%s` LIMIT %d",
		strings.Join(columnNames, ", "), tableName, limit)

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer rows.Close()

	// Get column types to properly handle NULL values
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error getting column types: %v", err)
	}

	var result []model.RowData

	// For each row
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columnNames))
		// Create a slice of pointers to the values
		scanArgs := make([]interface{}, len(columnNames))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Scan the row into the slice of interface{}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Create a map to hold the row data
		rowData := make(model.RowData)

		// For each column in the row
		for i, col := range columns {
			colName := col.Name

			// Handle different types
			var v interface{}
			val := values[i]

			// Handle NULL values
			if val == nil {
				v = nil
			} else {
				// Handle different types based on the column type
				switch colTypes[i].DatabaseTypeName() {
				case "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT":
					switch val.(type) {
					case []byte:
						v, _ = model.ParseInt(string(val.([]byte)))
					default:
						v = val
					}
				case "DECIMAL", "FLOAT", "DOUBLE":
					switch val.(type) {
					case []byte:
						v, _ = model.ParseFloat(string(val.([]byte)))
					default:
						v = val
					}
				default:
					// For TEXT, VARCHAR, etc. convert []byte to string
					switch val.(type) {
					case []byte:
						v = string(val.([]byte))
					default:
						v = val
					}
				}
			}

			rowData[colName] = v
		}

		result = append(result, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return result, nil
}
