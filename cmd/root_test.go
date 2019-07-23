package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePageConfig(t *testing.T) {
	pageInfo, err := readFromConfigFile("sample/config.toml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pageInfo, 2)
}
