package script

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type HStoreReadyCheckScript struct {
	Host    string
	Port    int
	Timeout int
}

func (m HStoreReadyCheckScript) GenScript() (string, error) {
	ph := filepath.Join("script", "wait_store_node_ready.sh.tpl")
	sh, err := embed.ReadScript(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("HStore").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, m); err != nil {
		return "", err
	}

	file := filepath.Join("template", "script", fmt.Sprintf("wait_store_node_ready_%s_%d.sh", m.Host, m.Port))
	return file, os.WriteFile(file, content.Bytes(), 0755)
}

type HStoreMountDiskScript struct {
	Host    string
	Shard   uint
	Disk    uint
	DataDir string
}

func (h HStoreMountDiskScript) GenScript() (string, error) {
	ph := filepath.Join("script", "store_mount_disk.sh.tpl")
	sh, err := embed.ReadScript(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("HStoreMount").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, h); err != nil {
		return "", err
	}

	file := filepath.Join("template", "script", fmt.Sprintf("store_mount_disk_%s.sh", h.Host))
	return file, os.WriteFile(file, content.Bytes(), 0755)
}
