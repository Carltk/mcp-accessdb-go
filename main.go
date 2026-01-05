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

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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
		ole.CoInitialize(0)
		defer ole.CoUninitialize()

		unknown, err := oleutil.CreateObject("ADOX.Catalog")
		if err != nil {
			return nil, fmt.Errorf("failed to create ADOX object: %v", err)
		}
		catalog, _ := unknown.QueryInterface(ole.IID_IDispatch)
		defer catalog.Release()

		// Try both common providers
		providers := []string{
			"Microsoft.Jet.OLEDB.4.0",
			"Microsoft.ACE.OLEDB.12.0",
			"Microsoft.ACE.OLEDB.16.0",
		}

		connected := false
		for _, p := range providers {
			connStr := fmt.Sprintf("Provider=%s;Data Source=%s;", p, args.DbPath)
			_, err = oleutil.PutProperty(catalog, "ActiveConnection", connStr)
			if err == nil {
				connected = true
				break
			}
		}

		if !connected {
			return nil, fmt.Errorf("failed to connect to database via ADOX: connection parameters restricted or driver missing")
		}

		tablesValue := oleutil.MustGetProperty(catalog, "Tables").ToIDispatch()
		defer tablesValue.Release()

		countVar := oleutil.MustGetProperty(tablesValue, "Count")
		count := int(countVar.Val)

		var tables []string
		for i := 0; i < count; i++ {
			table := oleutil.MustGetProperty(tablesValue, "Item", i).ToIDispatch()
			name := oleutil.MustGetProperty(table, "Name").ToString()
			typ := oleutil.MustGetProperty(table, "Type").ToString()
			if typ == "TABLE" {
				tables = append(tables, name)
			}
			table.Release()
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Tables: %v", tables))), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Printf("Server error: %v", err)
		log.Fatal(err)
	}
}
