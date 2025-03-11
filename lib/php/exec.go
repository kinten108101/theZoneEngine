package php
import (
	"os/exec"
)
func Exec(filepath string) (string, error) {
	shellCommand := "php " + filepath
	output, error := exec.Command("/usr/bin/sh", "-c", shellCommand).Output()
	stringOutput := string(output)
	return stringOutput, error
}
