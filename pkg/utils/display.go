package utils

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type Show interface {
	// Display will return components info with following columns:
	// | component name | host | ports | Image | container name | paths |
	Display() map[string]DisplayedComponent
}

type DisplayedComponent struct {
	Name          string
	Host          string
	Ports         string
	ContainerName string
	Image         string
	Paths         string
}

func ShowComponents(mp map[string]DisplayedComponent) {
	componentsArr := []string{"metaStore", "hstore", "hserver", "httpServer", "nodeExporter",
		"cadVisor", "hstreamExporter", "prometheus", "grafana", "alertManager"}
	header := []string{"Component", "Host", "Ports", "Image", "ContainerName", "Dirs"}

	data := make([][]string, 0, len(componentsArr))
	for _, k := range componentsArr {
		if component, ok := mp[k]; ok {
			row := make([]string, 0, 6)
			row = append(row, component.Name, component.Host, component.Ports,
				component.Image, component.ContainerName, component.Paths)
			data = append(data, row)
		}
	}

	res, err := RenderTable(header, data)
	if err != nil {
		log.Errorf("render table err: %s\n", err.Error())
		return
	}
	fmt.Println(res)
}

func calculateColLen(headers []string, data [][]string) ([]int, error) {
	headerSize := len(headers)
	res := make([]int, headerSize)
	for i, header := range headers {
		res[i] = len(header)
	}

	for _, cols := range data {
		if len(cols) != headerSize {
			return nil, errors.New("data and header size not match")
		}
		for idx, col := range cols {
			res[idx] = max(res[idx], len(col))
		}
	}
	return res, nil
}

func renderLine(colLengths []int) []byte {
	content := []byte{}
	for _, v := range colLengths {
		content = append(content, '+')
		content = append(content, bytes.Repeat([]byte{'-'}, v+2)...)
	}
	content = append(content, []byte("+\n")...)
	return content
}

func renderData(data []string, colLengths []int) []byte {
	content := []byte{}
	for i, v := range data {
		content = append(content, []byte("| ")...)
		content = append(content, []byte(v)...)
		content = append(content, bytes.Repeat([]byte{' '}, colLengths[i]-len(v)+1)...)
	}
	content = append(content, []byte("|\n")...)
	return content
}

func RenderTable(headers []string, data [][]string) (string, error) {
	colLengths, err := calculateColLen(headers, data)
	if err != nil {
		return "", err
	}

	table := []byte{}
	// render upper bodder
	table = append(table, renderLine(colLengths)...)
	// render header
	table = append(table, renderData(headers, colLengths)...)
	// render divider
	table = append(table, renderLine(colLengths)...)
	// render data
	for _, data := range data {
		table = append(table, renderData(data, colLengths)...)
	}
	// render bottom bodder
	table = append(table, renderLine(colLengths)...)
	return string(table), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
