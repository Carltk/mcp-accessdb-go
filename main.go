package main

import (
	"database/sql"
	"fmt"
	"log"

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
	connStr := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};DBQ=%s;", dbPath)
	return sql.Open("odbc", connStr)
}

func main() {
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport(), mcp_golang.WithName("mcp-accessdb-go"))

	// Query Tool
	err := server.RegisterTool("query", "Execute a SELECT query on an Access97 database", func(args QueryArgs) (*mcp_golang.ToolResponse, error) {
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

		// Access systemic table for MS Access is MSysObjects, but usually Type=1 and Flags=0 are user tables
		rows, err := db.Query("SELECT Name FROM MSysObjects WHERE Type=1 AND Flags=0")
		if err != nil {
			// Fallback if MSysObjects access is denied (common in some drivers)
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("Could not query MSysObjects. Try querying known table names.")), nil
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				tables = append(tables, name)
			}
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Tables: %v", tables))), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
