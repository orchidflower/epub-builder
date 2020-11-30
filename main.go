package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sort"
)

var builder EPubBuilder

func main() {
	app := cli.App{
		Name:    "epub-builder",
		Usage:   "Build ePub from text file.",
		Version: "1.0.0",
	}

	app.Commands = []*cli.Command{
		{
			Name:   "build",
			Usage:  "build a epub file from text.",
			Action: builder.Build,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "i",
					Aliases:     []string{"input"},
					Usage:       "Input text file",
					Destination: &builder.FileName,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "b",
					Aliases:     []string{"bookName"},
					Usage:       "Book Name",
					Destination: &builder.BookName,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "c",
					Aliases:     []string{"cover"},
					Usage:       "Cover image",
					Destination: &builder.Cover,
					Required:    true,
				},
			},
		},
		{
			Name:    "split",
			Aliases: []string{"send", "sendmail"},
			Usage:   "Split input text file to chapters.",
			Action:  builder.Split,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "i",
					Aliases:     []string{"input"},
					Usage:       "Input text file",
					Destination: &builder.FileName,
				},
			},
		},
	}
	app.Before = builder.Before
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
