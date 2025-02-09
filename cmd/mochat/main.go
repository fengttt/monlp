package main

import (
	"log/slog"
	"os"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/cmd/u"
	"github.com/matrixorigin/monlp/common"
)

func main() {
	common.ParseFlags(os.Args[1:])

	slog.Info("Starting mochat ...")
	sh := ishell.New()
	sh.SetHomeHistoryPath(".mochat_history")

	u.AddCmd(sh, "echo")
	u.AddCmd(sh, "sql")

	sh.AddCmd(&ishell.Cmd{
		Name: ".",
		Help: "chat with mochat",
		Func: func(c *ishell.Context) {
			c.Println("Hello, this is mochat.")
		},
	})

	sh.Run()
}
