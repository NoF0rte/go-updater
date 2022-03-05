//go:build !windows

package lib

import (
	"fmt"
	"os"

	builder "github.com/NoF0rte/cmd-builder"
)

func sudoExec(args string) error {
	return builder.Shell(fmt.Sprintf("sudo %s", args)).
		Stdin(os.Stdin).
		Run()
}
