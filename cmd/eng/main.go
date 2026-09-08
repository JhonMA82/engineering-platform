// Command eng is the thin M1 entrypoint over internal/cli.
package main

import (
	"os"

	"github.com/jhonma82/engineering-platform/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
