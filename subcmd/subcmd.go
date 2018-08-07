package subcmd

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	subCommands = make([]SubCommand, 0)
}

var subCommands []SubCommand

func AddSubCommand(sc SubCommand) {
	subCommands = append(subCommands, sc)
}

type SubCommand interface {
	Cmd() string
	Help() string
	Register(cc *kingpin.CmdClause)
	Run()
}

type register struct {
	CC *kingpin.CmdClause
	SC SubCommand
}

func RunApp(cmdName, help string) {
	app := kingpin.New(cmdName, help)

	regs := make([]register, 0)

	for _, sc := range subCommands {
		cc := app.Command(sc.Cmd(), sc.Help())
		sc.Register(cc)

		regs = append(regs, register{cc, sc})
	}

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	for _, reg := range regs {
		if reg.CC.FullCommand() == cmd {
			reg.SC.Run()
			break
		}
	}
}
