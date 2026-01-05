package main

import (
	"fmt"
	"log"

	"github.com/metoro-io/mcp-golang"
)

type ListFieldsArgs struct {
	DbPath    string `json:"dbPath" jsonschema:"required,description=Absolute path to the .mdb file"`
	TableName string `json:"tableName" jsonschema:"required,description=Name of the table to inspect"`
}

func registerTools(server *mcp_golang.Server, cfg *Config) {
	// Query Tool
	server.RegisterTool("query", "Execute a SELECT query on an Access97 database", func(args QueryArgs) (*mcp_golang.ToolResponse, error) {
		if cfg.Debug {
			log.Printf("Tool 'query' called with arguments: %+v", args)
		}
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

		resp := mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("%+v", results)))
		if cfg.Debug {
			log.Printf("Tool 'query' result: %+v", resp)
		}
		return resp, nil
	})

	// Execute Tool
	server.RegisterTool("execute", "Execute an INSERT, UPDATE, DELETE or CREATE statement", func(args UpdateArgs) (*mcp_golang.ToolResponse, error) {
		if cfg.Debug {
			log.Printf("Tool 'execute' called with arguments: %+v", args)
		}
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
		resp := mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Rows affected: %d", affected)))
		if cfg.Debug {
			log.Printf("Tool 'execute' result: %+v", resp)
		}
		return resp, nil
	})

	// List Tables Tool
	server.RegisterTool("list_tables", "List all user tables in the database using robust ADOX inspection", func(args ListTablesArgs) (*mcp_golang.ToolResponse, error) {
		if cfg.Debug {
			log.Printf("Tool 'list_tables' called with arguments: %+v", args)
		}
		tables, err := listAllTables(args.DbPath)
		if err != nil {
			return nil, err
		}
		resp := mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Tables: %v", tables)))
		if cfg.Debug {
			log.Printf("Tool 'list_tables' result: %+v", resp)
		}
		return resp, nil
	})

	// List Fields / Schema Tool
	server.RegisterTool("get_table_schema", "List all fields and the primary key for a specific table", func(args ListFieldsArgs) (*mcp_golang.ToolResponse, error) {
		if cfg.Debug {
			log.Printf("Tool 'get_table_schema' called with arguments: %+v", args)
		}
		schema, err := getTableMetadata(args.DbPath, args.TableName)
		if err != nil {
			return nil, err
		}
		resp := mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Schema for %s: %+v", args.TableName, schema)))
		if cfg.Debug {
			log.Printf("Tool 'get_table_schema' result: %+v", resp)
		}
		return resp, nil
	})
}
