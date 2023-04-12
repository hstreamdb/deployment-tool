package service

import "testing"
import "github.com/stretchr/testify/assert"

const (
	available800 = "8.0.0"
	available760 = "7.6.0"
)

func Test_isVersionCompatible(t *testing.T) {
	ret, err := isVersionCompatible("8.5.0", 8, 0, 0)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, ret)

	ret, err = isVersionCompatible("7.10.2", 7, 6, 0)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, ret)

	ret, err = isVersionCompatible("7.5.9", 7, 6, 0)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, ret)
}

func Test_whichIndexPatternToUse(t *testing.T) {
	assert.Equal(t, available800, whichIndexPatternToUse("8.5.0"))
	assert.Equal(t, available760, whichIndexPatternToUse("7.10.2"))
	assert.Equal(t, available800, whichIndexPatternToUse("latest"))
	assert.Equal(t, available800, whichIndexPatternToUse(""))
	//assert.Panics(t, func() { _ = whichIndexPatternToUse("7.5.9") }) should exit 1
}
