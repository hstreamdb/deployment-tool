package script

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/dev-deploy/embed"
	"os"
	"path/filepath"
	"text/template"
)

type HServerReadyCheckScript struct {
	Host    string
	Port    int
	Timeout int
}

func (m HServerReadyCheckScript) GenScript() (string, error) {
	ph := filepath.Join("script", "wait_tcp_timeout.sh.tpl")
	sh, err := embed.ReadScript(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("ServerReady").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, m); err != nil {
		return "", err
	}

	file := filepath.Join("template", "script", fmt.Sprintf("hserver_node_ready_%s_%d.sh", m.Host, m.Port))
	return file, os.WriteFile(file, content.Bytes(), 0755)
}
