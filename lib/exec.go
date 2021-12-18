package lib

import (
	"fmt"
	"os"
	"os/exec"
)

func sudoExec(args string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sudo %s", args))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
