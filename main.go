package main

import (
	"github.com/urfave/cli"
	"time"
	"os"
	"log"
	"github.com/lpisces/cnpostcode/cmd/scrape"
	"github.com/lpisces/cnpostcode/cmd/api"
)

func main() {
	app := cli.NewApp()
	app.Name = "postcode"
	app.Usage = "scrape cn postcode and provide query api"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "zebrapool",
			Email: "iamalazyrat@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "scrape",
			Aliases: []string{"s"},
			Usage:   "scrape cn postcode",
			Action:  scrape.Run,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "debug, d",
					Usage: "show debug info",
				},
				cli.Int64Flag{
					Name:  "number, n",
					Value: 1,
					Usage: "spider number",
				},
				cli.StringFlag{
					Name: "key, k",
					Usage: "access key",
					Value: "",
				},
				cli.StringFlag{
					Name: "c, cache",
					Usage: "cache dir path",
					Value: "./cache",
				},
				cli.StringFlag{
					Name: "o, output",
					Usage: "output dir path",
					Value: "./data",
				},
			},
		},
		{
			Name:    "api",
			Aliases: []string{"a"},
			Usage:   "serve postcode query api",
			Action:  api.Run,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "debug, d",
					Usage: "show debug info",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
