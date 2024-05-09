package script

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type EsIndexScript struct {
	Host string
	Port int
}

func (e EsIndexScript) GenScript() (string, error) {
	ph := filepath.Join("script", "config_elasticsearch_index.sh.tpl")
	sh, err := embed.ReadScript(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("SetEsIndex").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, e); err != nil {
		return "", err
	}

	file := filepath.Join("template", "script", fmt.Sprintf("config_es_index_%s_%d.sh", e.Host, e.Port))
	return file, os.WriteFile(file, content.Bytes(), 0755)
}
