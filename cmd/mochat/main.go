package main

import (
	"log/slog"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/common"
)

func main() {
	common.ParseFlags()

	slog.Info("Starting mochat ...")
	sh := ishell.New()
	sh.SetHomeHistoryPath(".mochat_history")

	sh.AddCmd(&ishell.Cmd{
		Name: ".echo",
		Help: "echo the input",
		Func: func(c *ishell.Context) {
			c.Println("RawArgs:", c.RawArgs)
			c.Println("Args:", c.Args)
		},
	})

	sh.AddCmd(&ishell.Cmd{
		Name: ".use",
		Help: "use chat context",
		Func: func(c *ishell.Context) {
			c.Println("Use RawArgs:", c.RawArgs)
			c.Println("Use Args:", c.Args)
		},
	})

	sh.AddCmd(&ishell.Cmd{
		Name: ".",
		Help: "chat with mochat",
		Func: func(c *ishell.Context) {
			c.Println("Hello, this is mochat.")
		},
	})

	sh.Run()
}
