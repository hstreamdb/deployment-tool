package service

import (
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"gotest.tools/v3/assert"
	"testing"
)

func TestCheckNeedSeedNodes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input    utils.Version
		expected bool
	}{
		"less than 0.8.2": {
			input:    utils.Version{0, 7, 1, false},
			expected: false,
		},
		"equal 0.8.2": {
			input:    utils.Version{0, 8, 2, false},
			expected: false,
		},
		"equal 0.8.3": {
			input:    utils.Version{0, 8, 3, false},
			expected: true,
		},
		"equal 0.8.4": {
			input:    utils.Version{0, 8, 4, false},
			expected: false,
		},
		"greater than 0.8.4": {
			input:    utils.Version{0, 9, 1, false},
			expected: true,
		},
		"latest": {
			input:    utils.Version{IsLatest: true},
			expected: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			get := needSeedNodes(tc.input)
			assert.Equal(t, get, tc.expected)
		})
	}
}

func TestParseImage(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input   string
		image   string
		version utils.Version
	}{
		"standard name": {
			input:   "hstreamdb/hstream:v0.10.0",
			image:   "hstreamdb/hstream",
			version: utils.Version{0, 10, 0, false},
		},
		"image lack version": {
			input:   "hstreamdb/hstream",
			image:   "hstreamdb/hstream",
			version: utils.Version{IsLatest: true},
		},
		"unexpected image name": {
			input:   "hstreamdb/hstream:rqlite",
			image:   "hstreamdb/hstream:rqlite",
			version: utils.Version{IsLatest: true},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			image, version := parseImage(tc.input)
			assert.Equal(t, image, tc.image)
			assert.Equal(t, version, tc.version)
		})
	}
}

func TestGetMetaStoreUrl(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input string
		tp    spec.MetaStoreType
		want  string
	}{
		"zk url": {
			input: "host1:2181,host2:2181",
			tp:    spec.ZK,
			want:  "zk://host1:2181,host2:2181",
		},
		"rqlite url": {
			input: "http://host1:2181,http://host2:2181",
			tp:    spec.RQLITE,
			want:  "rq://host1:2181,host2:2181",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			url := getMetaStoreUrl(tc.tp, tc.input)
			assert.Equal(t, url, tc.want)
		})
	}
}
