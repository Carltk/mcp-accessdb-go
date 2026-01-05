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

	registerTools(server)

	if err := server.Serve(); err != nil {
		log.Printf("Server error: %v", err)
		log.Fatal(err)
	}
}
