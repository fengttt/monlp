package dbagent

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
)

type MoDB struct {
	db *sql.DB
}

func (db *MoDB) Close() error {
	if db.db != nil {
		err := db.db.Close()
		db.db = nil
		return err
	}
	return nil
}

func ConnStr(host, port, user, passwd, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		user, passwd, host, port, dbname)
}

func PyConnStr(host, port, user, passwd, dbname string) string {
	return fmt.Sprintf("mysql+pymysql://%s:%s@%s:%s/%s",
		user, passwd, host, port, dbname)
}

func OpenDB(connstr string) (*MoDB, error) {
	db, err := sql.Open("mysql", connstr)
	if err != nil {
		return nil, err
	}
	return &MoDB{db: db}, nil
}

func (db *MoDB) Exec(sql string, params ...any) error {
	_, err := db.db.Exec(sql, params...)
	return err
}

func (db *MoDB) MustExec(sql string, params ...any) {
	err := db.Exec(sql, params...)
	if err != nil {
		panic(err)
	}
}

func (db *MoDB) Prepare(sql string) (*sql.Stmt, error) {
	return db.db.Prepare(sql)
}

func (db *MoDB) Begin() (*sql.Tx, error) {
	return db.db.Begin()
}

func (db *MoDB) QueryVal(sql string, params ...any) (string, error) {
	rows, err := db.db.Query(sql, params...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}
	var ret string
	rows.Scan(&ret)
	return ret, nil
}

func (db *MoDB) QueryIVal(sql string, params ...any) (int64, error) {
	rows, err := db.db.Query(sql, params...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}
	var ret int64
	rows.Scan(&ret)
	return ret, nil
}

func (db *MoDB) Query(sql string, params ...any) ([][]string, error) {
	rows, err := db.db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	ncol := len(cols)
	if ncol == 0 {
		return nil, nil
	}

	var ret [][]string
	row := make([]any, ncol)
	for rows.Next() {
		data := make([]string, ncol)
		for i := 0; i < ncol; i++ {
			row[i] = &data[i]
		}
		rows.Scan(row...)
		ret = append(ret, data)
	}
	return ret, nil
}

func (db *MoDB) QueryDump(sql string, params ...any) (string, error) {
	rows, err := db.db.Query(sql, params...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	ncol := len(cols)
	if ncol == 0 {
		return "", nil
	}

	sb := &strings.Builder{}
	tw := tablewriter.NewWriter(sb)
	tw.SetHeader(cols)
	tw.SetBorders(tablewriter.Border{Left: true, Right: true, Top: false, Bottom: false})
	tw.SetCenterSeparator("|")

	for rows.Next() {
		row := make([]interface{}, ncol)
		data := make([]string, ncol)
		for i := 0; i < ncol; i++ {
			row[i] = &data[i]
		}
		rows.Scan(row...)
		tw.Append(data)
	}

	tw.Render()
	return sb.String(), nil
}

func token2q(tokens []string, dict map[string]string) (string, []any) {
	var tks []string
	var params []any
	// poorman's macro
	for _, v := range tokens {
		if len(v) >= 2 && v[0] == ':' && v[len(v)-1] == ':' {
			// :FOO: will expand FOO
			vk := v[1 : len(v)-1]
			tks = append(tks, dict[vk])
		} else if len(v) >= 2 && v[0] == '?' && v[len(v)-1] == '?' {
			// ?FOO? will bind FOO as param
			vk := v[1 : len(v)-1]
			tks = append(tks, "?")
			params = append(params, dict[vk])
		} else if len(v) >= 2 && v[0] == '$' && v[len(v)-1] == '$' {
			// $FOO$ will become 'FOO', to work around ishell quote
			vk := v[1 : len(v)-1]
			tks = append(tks, "'"+vk+"'")
		} else {
			tks = append(tks, v)
		}
	}
	qry := strings.Join(tks, " ")
	return qry, params
}

func qSave(db *MoDB, sql string, params []any, f *os.File) error {
	rows, err := db.db.Query(sql, params...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	ncol := len(cols)
	if ncol == 0 {
		return nil
	}

	for rows.Next() {
		row := make([]interface{}, ncol)
		data := make([]string, ncol)
		for i := 0; i < ncol; i++ {
			row[i] = &data[i]
		}
		rows.Scan(row...)
		f.WriteString(strings.Join(data, ",") + "\n")
	}
	return nil
}
