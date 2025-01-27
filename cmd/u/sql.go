package u

import (
	"fmt"
	"path"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/common"
)

var (
	db *dbagent.MoDB
)

func openDB(args []string) error {
	if db != nil {
		return nil
	}

	var err error
	var driver string
	var connstr string

	if len(args) == 0 {
		driver = "mysql"
	} else {
		driver = args[0]
	}

	if len(args) < 2 {
		switch driver {
		case "mysql":
			connstr = dbagent.ConnStr("localhost", "6001", "dump", "111", "monlp")
		case "dslite", "sqlite":
			connstr = path.Join(common.WorkingDir, "monlp.db")
		default:
			return fmt.Errorf("unsupported driver: %s", driver)
		}
	} else {
		connstr = args[1]
	}

	db, err = dbagent.OpenDB(driver, connstr)
	return err
}

func SqlDriverCmd(c *ishell.Context) {
	if db != nil {
		db.Close()
		db = nil
	}
	if err := openDB(c.Args); err != nil {
		c.Println(err)
	}
}

func SqlCmd(c *ishell.Context) {
	if len(c.Args) == 0 {
		c.Println("No sql to execute")
		return
	}

	if err := openDB(nil); err != nil {
		c.Println(err)
		return
	}

	sql := strings.Join(c.Args, " ")
	res, err := db.QueryDump(sql)
	if err != nil {
		c.Println(err)
		return
	} else {
		c.Println(res)
	}
}
