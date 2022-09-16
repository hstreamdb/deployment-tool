package utils

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

func CheckSSHAuthentication(identityFile string, usePassword bool) (string, string, error) {
	if usePassword {
		fmt.Println("Input SSH password: ")
		input, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", "", err
		}
		password := strings.TrimSpace(strings.Trim(string(input), "\n"))
		return "", password, nil
	}

	if len(identityFile) != 0 {
		buf, err := os.ReadFile(identityFile)
		if err != nil {
			return "", "", fmt.Errorf("failed to read identity file %s: %s", identityFile, err.Error())
		}

		if _, err := ssh.ParsePrivateKey(buf); err != nil {
			return "", "", fmt.Errorf("unable to parse identity file %s: %s", identityFile, err.Error())
		}
		return identityFile, "", nil
	}

	//if len(identityFile) == 0 || !CheckExist(identityFile) {
	//	return "", "", fmt.Errorf("need to specify identify-file or password")
	//}
	return "", "", nil
}

func CheckExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
