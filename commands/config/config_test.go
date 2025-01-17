package config

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
)

func TestRead(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	config.SetFilename("./testdata/mock_config.yaml")

	cmd := NewCmd(db, nil)
	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed reading config: %v", err)
	}
}

func TestReadError(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	config.SetFilename("")

	cmd := NewCmd(db, nil)
	if err := cmd.Execute(); err == nil {
		t.Error("Expected an error and got nil")
	}
}
