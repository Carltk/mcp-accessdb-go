# mcp-accessdb-go

A Model Context Protocol (MCP) server written in Go for interacting with legacy Microsoft Access (.mdb) databases.

## Features

- **query**: Execute SELECT statements.
- **execute**: Execute INSERT, UPDATE, DELETE, or CREATE statements.
- **list_tables**: List user tables (Type 1) in the database, with compatibility fixes for Access 97.

## Requirements

- **Windows Architecture**: This server is built as a **32-bit (x86)** application to ensure compatibility with the legacy `Microsoft Access Driver (*.mdb)` ODBC driver.
- **ODBC Driver**: Ensure the "Microsoft Access Driver (*.mdb, *.accdb)" or the older legacy driver is installed.

## Configuration

The server looks for a `config.yaml` in its executable directory. If not found, it defaults to using the system temporary directory for logging.

### Example `config.yaml`
```yaml
logDir: "C:\\Custom\\Logs"
```

## Usage in Droid / Factory

Add the following to your MCP configuration file (typically `C:\Users\<User>\.factory\mcp.json`):

```json
{
  "mcpServers": {
    "mcp-accessdb-go": {
      "command": "D:\\mydev\\mcp-accessdb-go\\mcp-accessdb.exe",
      "args": [],
      "disabled": false,
      "type": "stdio"
    }
  }
}
```

## Development

To build the 32-bit executable:
```bash
$env:GOARCH="386"
go build -o mcp-accessdb.exe main.go config.go
```
