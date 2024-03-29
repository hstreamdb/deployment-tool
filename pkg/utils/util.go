package utils

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
		log.Errorf("scp command error: %s", err.Error())
		os.Exit(1)
	}
	return position
}

func ScpDirFiles(originPath, remotePath string) []executor.Position {
	position := []executor.Position{}

	if err := filepath.WalkDir(originPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			position = append(position, executor.Position{LocalDir: path, RemoteDir: filepath.Join(remotePath, d.Name())})
		}
		return nil
	}); err != nil {
		log.Errorf("scp command error: %s", err.Error())
		os.Exit(1)
	}
	return position
}

func CpDir(src, dst string) error {
	cmd := exec.Command("rm", "-rf", dst)
	log.Debugf("exec %s", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Errorf("rm %s error: %s", dst, err.Error())
		return err
	}
	cmd = exec.Command("cp", "-r", src, dst)
	log.Debugf("exec %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

var (
	// Version082 hserver version <= v0.8.2 or == 0.8.4 should
	// start without `--seed` argument
	Version082 = Version{0, 8, 2, false}
	Version084 = Version{0, 8, 4, false}
	// Version090 hserver version >= v0.9.0 need call server init
	Version090 = Version{0, 9, 0, false}
	// Version095 hserver version > v0.9.5 should replace argument
	// `--zkuri` to `--meta-store` when start
	Version095 = Version{0, 9, 5, false}
	// Version096 hserver version > v0.9.6 should replace argument
	// `--meta-store` to `--metastore-uri`
	Version096 = Version{0, 9, 6, false}
	// Version0100 hserver version >= v0.10.0 should support rqlite
	// as meta store
	Version0100 = Version{0, 10, 0, false}
	// Version0101 hserver version >= v0.10.1 should replace argument
	// `--host` and `--address` to `--bind-address` and
	// `--advertised-address`, `--compression` should be removed
	Version0101 = Version{0, 10, 1, false}

	// ElkVersion800 Kibana saved objects format for 8.*.*
	ElkVersion800 = Version{8, 0, 0, false}
	// ElkVersion760 Kibana saved objects format for 7.8.*
	ElkVersion760 = Version{7, 6, 0, false}
	// ElkVersion7130 Kibana `ServerShutdownTimeout` option
	ElkVersion7130 = Version{7, 13, 0, false}
)

type Version struct {
	Major, Minor, Patch int
	IsLatest            bool
}

func CreateVersion(ver string) Version {
	if len(ver) == 0 || ver == "latest" {
		return Version{IsLatest: true}
	}

	ver = strings.TrimSpace(ver)
	ver = strings.TrimPrefix(ver, "v")
	fragment := strings.Split(ver, ".")
	codes := make([]int, 3)
	for idx, c := range fragment {
		code, _ := strconv.Atoi(c)
		codes[idx] = code
	}
	return Version{
		Major: codes[0],
		Minor: codes[1],
		Patch: codes[2],
	}
}

func CompareVersion(lh, rh Version) int {
	if lh.IsLatest && rh.IsLatest {
		return 0
	}
	if lh.IsLatest {
		return 1
	}
	if rh.IsLatest {
		return -1
	}
	if res := compareSegment(lh.Major, rh.Major); res != 0 {
		return res
	}
	if res := compareSegment(lh.Minor, rh.Minor); res != 0 {
		return res
	}
	return compareSegment(lh.Patch, rh.Patch)
}

func compareSegment(l, r int) int {
	if l < r {
		return -1
	}
	if l > r {
		return 1
	}
	return 0
}

type DirCfg struct {
	Path string
	Perm fs.FileMode
}

func MakeDirs(dirs []DirCfg) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir.Path, dir.Perm); err != nil {
			return fmt.Errorf("create %s error: %s\n", dir.Path, err.Error())
		}
	}
	return nil
}
