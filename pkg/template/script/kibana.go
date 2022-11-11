package script

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type KibanaReadyCheck struct {
	KibanaHost string
	KibanaPort string
	FilePath   string
	Timeout    string
}

func (m KibanaReadyCheck) GenScript() (string, error) {
	ph := filepath.Join("script", "wait_kibana_timeout.sh.tpl")
	sh, err := embed.ReadScript(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("MetaStore").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, m); err != nil {
		return "", err
	}

	file := filepath.Join("template", "script", fmt.Sprintf("wait_kibana_timeout%s_%s.sh", m.KibanaHost, m.KibanaPort))
	return file, os.WriteFile(file, content.Bytes(), 0755)
}
