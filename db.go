package main

import (
	"database/sql"

	_ "github.com/lib/pq"
  "fmt"

  "errors"
  "time"
  "reflect"
)

type QueryResult struct {
  columns []string
  values  [][]string
  err error
}

type DBInfo struct {
  name, user, pass, host, port string
  ssl bool
}

func ConnectDb(info *DBInfo) (*sql.DB, error) {
  connStr := ""
  connStr += " host=" + info.host
  connStr += " port=" + info.port
  connStr += " user=" + info.user
  connStr += " password=" + info.pass
  connStr += " dbname=" + info.name

  if !info.ssl {
    connStr += " sslmode=disable"
  }

  db, err := sql.Open("postgres", connStr)
  if err != nil {
    return nil, err
  }

  err = db.Ping()
  return db, err
}

func GetQueryResult(db *sql.DB, query string) QueryResult {
  result := QueryResult{columns: []string{}, values: [][]string{}, err: nil}

	rows, err := db.Query(query)

  if err != nil {
    result.err = err
    return result
  }

  defer rows.Close()

  colNames, err := rows.Columns()
  if err != nil {
    result.err = err
    return result
  }

  var myMap = make(map[string]interface{})

  cols := make([]interface{}, len(colNames))
  colPtrs := make([]interface{}, len(colNames))
  for i := 0; i < len(colNames); i++ {
    colPtrs[i] = &cols[i]
  }

  var keys = make(map[string]int)
  result.columns = make([]string, len(colNames))
  result.values = [][]string{}

  for rows.Next() {
    err = rows.Scan(colPtrs...)
    if err != nil {
      result.err = err
      return result
    }
    for i, col := range cols {
      myMap[colNames[i]] = col
      result.columns[i] = fmt.Sprint(colNames[i])
      keys[colNames[i]] = i
    }

    row := make([]string, len(colNames))

    for k, val := range myMap {
      i := keys[k]

      if val == nil {
        row[i] = "nil"
        continue
      }

      t, isTime := val.(time.Time)
      if isTime {
        row[i] = fmt.Sprint(t.Format(time.RFC3339))
      } else if reflect.TypeOf(val).Kind() == reflect.Slice {
        v, ok := val.([]byte)
        if ok {
          row[i] = string(v)
        } else {
          row[i] = "-*-*-"
        }
      } else {
        row[i] = fmt.Sprint(val)
      }
    }

    result.values = append(result.values, row)
  }

  if len(result.values) == 0 {
    for i := 0; i < len(colNames); i++ {
      result.columns[i] = fmt.Sprint(colNames[i])
    }
  }

  return result
}

func GetOID(db *sql.DB, tablename string) (string, error) {
  query :=
    `SELECT c.oid, n.nspname, c.relname
     FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     WHERE n.nspname = 'public' and c.relname = '%s'
     ORDER BY 2, 3;`

  query = fmt.Sprintf(query, tablename)

  result := GetQueryResult(db, query)

  if result.err != nil {
    return "", result.err
  }

  if len(result.values) == 0 {
    return "", errors.New("oid relation doesn't exist.")
  }

  return result.values[0][0], nil
}

func GetConstraints(db *sql.DB, oid string) QueryResult {
  query :=
    `SELECT c2.relname, i.indisprimary, i.indisunique,
            --i.indisclustered, c2.reltablespace, i.indisvalid,
            pg_catalog.pg_get_indexdef(i.indexrelid, 0, true),
            pg_size_pretty(pg_relation_size(c2.oid)) as index_size
     FROM pg_catalog.pg_index i
     INNER JOIN  pg_catalog.pg_class c1 ON i.indrelid = c1.oid
     INNER JOIN  pg_catalog.pg_class c2 ON i.indexrelid = c2.oid
     WHERE c1.oid = ` + oid + `
     ORDER BY i.indisprimary DESC, i.indisunique DESC, c2.relname;`

  return GetQueryResult(db, query)
}

func GetTables(db *sql.DB) QueryResult {
  query := `SELECT table_name FROM information_schema.tables
            WHERE table_schema = 'public'`
  return GetQueryResult(db, query)
}

func GetColumns(db *sql.DB, tablename string) QueryResult {
  query :=
    `SELECT cols.column_name, cols.data_type, cols.character_maximum_length,
            cols.numeric_precision, cols.numeric_scale,
            cols.column_default, cols.is_nullable, NOT (cs.constraint_name is NULL) AS pk
    FROM information_schema.columns cols                                                                                       
    LEFT JOIN
    (
      SELECT tco.constraint_name, kcu.column_name AS key_column
      FROM information_schema.table_constraints tco
      JOIN information_schema.key_column_usage kcu ON
        kcu.constraint_name = tco.constraint_name     AND
        kcu.constraint_schema = tco.constraint_schema AND
        kcu.constraint_name = tco.constraint_name
      WHERE tco.constraint_type = 'PRIMARY KEY' AND kcu.table_name = '%s'
    ) cs
    ON cs.key_column = cols.column_name
    WHERE table_name = '%s';
    `
  query = fmt.Sprintf(query, tablename, tablename)

  return GetQueryResult(db, query)
}
