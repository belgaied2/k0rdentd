package cli

import (
	"github.com/belgaied2/k0rdentd/pkg/ui"
	"github.com/urfave/cli/v2"
)

var ExposeUICommand = &cli.Command{
	Name:      "expose-ui",
	Aliases:   []string{"ui"},
	Usage:     "Expose K0rdent UI via ingress",
	UsageText: "k0rdentd expose-ui [options]",
	Action:    exposeUIAction,
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Value:   0,
			Usage:   "Timeout for waiting for deployment readiness (0 = use default 5m)",
		},
	},
}

func exposeUIAction(c *cli.Context) error {
	return ui.ExposeUI()
}
