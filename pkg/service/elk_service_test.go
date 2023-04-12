package service

import (
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"testing"
)
import "github.com/stretchr/testify/assert"

func Test_isVersionCompatible(t *testing.T) {
	ret := isVersionCompatible("8.5.0", utils.ElkVersion800)
	assert.True(t, ret)

	ret = isVersionCompatible("7.10.2", utils.ElkVersion760)
	assert.True(t, ret)

	ret = isVersionCompatible("7.5.9", utils.ElkVersion760)
	assert.False(t, ret)
}

func Test_whichIndexPatternToUse(t *testing.T) {
	assert.Equal(t, available800, whichIndexPatternToUse("8.5.0"))
	assert.Equal(t, available760, whichIndexPatternToUse("7.10.2"))
	assert.Equal(t, available800, whichIndexPatternToUse("latest"))
	assert.Equal(t, available800, whichIndexPatternToUse(""))
	//assert.Panics(t, func() { _ = whichIndexPatternToUse("7.5.9") }) should exit 1
}
