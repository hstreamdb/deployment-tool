package utils

import (
	"fmt"
	"github.com/hstreamdb/dev-deploy/pkg/executor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"io/fs"
	"os"
	"path/filepath"
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

func ScpDir(originPath, remotePath string) []executor.Position {
	position := []executor.Position{}

	if err := filepath.WalkDir(originPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths := strings.Split(path, "/")
			n := len(paths)
			if strings.HasSuffix(remotePath, paths[n-2]) {
				position = append(position, executor.Position{LocalDir: path, RemoteDir: filepath.Join(remotePath, paths[n-1])})
			} else {
				position = append(position, executor.Position{LocalDir: path, RemoteDir: filepath.Join(remotePath, paths[n-2], paths[n-1])})
			}

		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("scp command error: %s", err.Error()))
	}
	return position
}
