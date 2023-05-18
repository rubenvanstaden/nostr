package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Action: func(cCtx *cli.Context) error {
            if cCtx.NArg() == 0 {
                fmt.Println("ERROR: Specify the relay URL")
            }
            if cCtx.NArg() == 1 {
                url := cCtx.Args().Get(0)
                r := Relay{ url: url }
                r.Connect()
                fmt.Println("Relay URL: ", url)
            }
            if cCtx.NArg() == 2 {
                url := cCtx.Args().Get(0)
                msg := cCtx.Args().Get(1)
                fmt.Printf("\n> Publishing msg [%s] to relay [%s]\n", msg, url)
            }
            return nil
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}
