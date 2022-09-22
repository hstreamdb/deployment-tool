package service

import (
	"github.com/hstreamdb/dev-deploy/pkg/utils"
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
