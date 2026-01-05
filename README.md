# mcp-accessdb-go

A Model Context Protocol (MCP) server written in Go for interacting with legacy Microsoft Access (.mdb) databases.

## Features

- **query**: Execute SELECT statements.
- **execute**: Execute INSERT, UPDATE, DELETE, or CREATE statements.
- **list_tables**: List user tables in the database using robust ADOX collection iteration.
- **get_table_schema**: Retrieve all field names and the primary key for a specified table.

## Requirements

- **Windows Architecture**: This server must be built as a **32-bit (x86)** application to ensure compatibility with legacy `Microsoft Access Driver (*.mdb)` ODBC and OLEDB providers.
- **ODBC Driver**: Ensure the "Microsoft Access Driver (*.mdb, *.accdb)" or the older legacy driver is installed.

## Configuration

The server looks for a `config.yaml` in its executable directory. It supports relative paths (resolved against the executable location).

### Example `config.yaml`
```yaml
# Directory for log files
logDir: "./log"
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

## Project Structure

- `main.go`: Server entry point and configuration loading.
- `tools.go`: MCP tool definitions and handlers.
- `schema.go`: Database metadata discovery using OLE/ADOX.
- `config.go`: Configuration file management.

## Development

To build the 32-bit executable:
```bash
$env:GOARCH="386"
go build -o mcp-accessdb.exe main.go tools.go schema.go config.go
```
