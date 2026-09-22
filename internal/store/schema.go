package store

import (
	"database/sql"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

var CodeSchema = faults.Code("ocrecent.store.schema")

// Inspect returns the column set of table.
func Inspect(db *sql.DB, table string) (model.Cols, error) {
	rows, err := db.Query("PRAGMA table_info(" + quoteIdent(table) + ")")
	if err != nil {
		return nil, faults.Wrap(CodeSchema, "inspect table", err, "table", table)
	}
	defer rows.Close()

	cols := model.Cols{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, faults.Wrap(CodeSchema, "scan table_info", err, "table", table)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, faults.Wrap(CodeSchema, "iterate table_info", err, "table", table)
	}
	return cols, nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
