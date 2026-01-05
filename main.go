package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/alexbrainman/odbc"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

type QueryArgs struct {
	DbPath string `json:"dbPath" jsonschema:"required,description=Absolute path to the .mdb file"`
	SQL    string `json:"sql" jsonschema:"required,description=The SELECT SQL query to execute"`
}

type UpdateArgs struct {
	DbPath string `json:"dbPath" jsonschema:"required,description=Absolute path to the .mdb file"`
	SQL    string `json:"sql" jsonschema:"required,description=The INSERT/UPDATE/DELETE/CREATE SQL statement"`
}

type ListTablesArgs struct {
	DbPath string `json:"dbPath" jsonschema:"required,description=Absolute path to the .mdb file"`
}

func getConn(dbPath string) (*sql.DB, error) {
	// For Access 97 MDB files, the older 32-bit driver is often required.
	// We use the driver name specifically registered for MDB.
	connStr := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb)};DBQ=%s;", dbPath)
	return sql.Open("odbc", connStr)
}

func main() {
	cfg := LoadConfig()

	// Setup logging to configured or default folder.
	logPath := filepath.Join(cfg.LogDir, "mcp-accessdb-go.log")
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}

	server := mcp_golang.NewServer(stdio.NewStdioServerTransport(), mcp_golang.WithName("mcp-accessdb-go"))

	// Query Tool
	err = server.RegisterTool("query", "Execute a SELECT query on an Access97 database", func(args QueryArgs) (*mcp_golang.ToolResponse, error) {
		db, err := getConn(args.DbPath)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		rows, err := db.Query(args.SQL)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		var results []map[string]interface{}
		for rows.Next() {
			columnPointers := make([]interface{}, len(cols))
			columnValues := make([]interface{}, len(cols))
			for i := range columnValues {
				columnPointers[i] = &columnValues[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				return nil, err
			}

			m := make(map[string]interface{})
			for i, colName := range cols {
				val := columnValues[i]
				if b, ok := val.([]byte); ok {
					m[colName] = string(b)
				} else {
					m[colName] = val
				}
			}
			results = append(results, m)
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("%+v", results))), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Update Tool
	err = server.RegisterTool("execute", "Execute an INSERT, UPDATE, DELETE or CREATE statement", func(args UpdateArgs) (*mcp_golang.ToolResponse, error) {
		db, err := getConn(args.DbPath)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		res, err := db.Exec(args.SQL)
		if err != nil {
			return nil, err
		}

		affected, _ := res.RowsAffected()
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Rows affected: %d", affected))), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// List Tables Tool (Schema discovery)
	err = server.RegisterTool("list_tables", "List all user tables in the database", func(args ListTablesArgs) (*mcp_golang.ToolResponse, error) {
		db, err := getConn(args.DbPath)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		// Attempt discovery using MSysObjects (requires Read Design permissions)
		rows, err := db.Query("SELECT Name FROM MSysObjects WHERE Type=1 AND Name NOT LIKE 'MSys*' AND Name NOT LIKE 'f_*'")
		if err == nil {
			defer rows.Close()
			var tables []string
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err == nil {
					tables = append(tables, name)
				}
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Tables: %v", tables))), nil
		}

		// Fallback: If MSysObjects is restricted, use the driver's own metadata if possible.
		// Since database/sql doesn't expose SQLTables directly, and we can't easily cast to internal odbc.Conn types,
		// we'll try a common Access-specific probing logic or report the permission issue clearly.
		log.Printf("MSysObjects access denied: %v. Database is connected.", err)
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Connected to database, but MSysObjects is restricted. You may need to grant 'Read Design' permissions on system tables in Access. Error: %v", err))), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Printf("Server error: %v", err)
		log.Fatal(err)
	}
}
