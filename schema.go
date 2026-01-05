package main

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type TableSchema struct {
	Fields     []string `json:"fields"`
	PrimaryKey []string `json:"primaryKey"`
}

func getTableMetadata(dbPath string, tableName string) (*TableSchema, error) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("ADOX.Catalog")
	if err != nil {
		return nil, fmt.Errorf("failed to create ADOX object: %v", err)
	}
	catalog, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer catalog.Release()

	providers := []string{
		"Microsoft.Jet.OLEDB.4.0",
		"Microsoft.ACE.OLEDB.12.0",
		"Microsoft.ACE.OLEDB.16.0",
	}

	connected := false
	for _, p := range providers {
		connStr := fmt.Sprintf("Provider=%s;Data Source=%s;", p, dbPath)
		_, err = oleutil.PutProperty(catalog, "ActiveConnection", connStr)
		if err == nil {
			connected = true
			break
		}
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect via ADOX")
	}

	tables := oleutil.MustGetProperty(catalog, "Tables").ToIDispatch()
	defer tables.Release()

	table := oleutil.MustGetProperty(tables, "Item", tableName).ToIDispatch()
	defer table.Release()

	schema := &TableSchema{
		Fields:     []string{},
		PrimaryKey: []string{},
	}

	// Get Columns
	columns := oleutil.MustGetProperty(table, "Columns").ToIDispatch()
	colCount := int(oleutil.MustGetProperty(columns, "Count").Val)
	for i := 0; i < colCount; i++ {
		col := oleutil.MustGetProperty(columns, "Item", i).ToIDispatch()
		schema.Fields = append(schema.Fields, oleutil.MustGetProperty(col, "Name").ToString())
		col.Release()
	}
	columns.Release()

	// Get Primary Key from Indexes
	indexes := oleutil.MustGetProperty(table, "Indexes").ToIDispatch()
	idxCount := int(oleutil.MustGetProperty(indexes, "Count").Val)
	for i := 0; i < idxCount; i++ {
		index := oleutil.MustGetProperty(indexes, "Item", i).ToIDispatch()
		isPK := oleutil.MustGetProperty(index, "PrimaryKey").Value().(bool)
		if isPK {
			pkColsDisp := oleutil.MustGetProperty(index, "Columns").ToIDispatch()
			pkColCount := int(oleutil.MustGetProperty(pkColsDisp, "Count").Val)
			for j := 0; j < pkColCount; j++ {
				pkCol := oleutil.MustGetProperty(pkColsDisp, "Item", j).ToIDispatch()
				schema.PrimaryKey = append(schema.PrimaryKey, oleutil.MustGetProperty(pkCol, "Name").ToString())
				pkCol.Release()
			}
			pkColsDisp.Release()
		}
		index.Release()
	}
	indexes.Release()

	return schema, nil
}

func listAllTables(dbPath string) ([]string, error) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("ADOX.Catalog")
	if err != nil {
		return nil, err
	}
	catalog, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer catalog.Release()

	providers := []string{"Microsoft.Jet.OLEDB.4.0", "Microsoft.ACE.OLEDB.12.0", "Microsoft.ACE.OLEDB.16.0"}
	connected := false
	for _, p := range providers {
		connStr := fmt.Sprintf("Provider=%s;Data Source=%s;", p, dbPath)
		_, err = oleutil.PutProperty(catalog, "ActiveConnection", connStr)
		if err == nil {
			connected = true
			break
		}
	}
	if !connected {
		return nil, fmt.Errorf("failed to connect via ADOX")
	}

	tablesValue := oleutil.MustGetProperty(catalog, "Tables").ToIDispatch()
	defer tablesValue.Release()

	count := int(oleutil.MustGetProperty(tablesValue, "Count").Val)
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
	return tables, nil
}
