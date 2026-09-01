package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// The chain is read once per process. Every subcommand needs it and none of them may
// see a different answer than another: a run where `check` used one vault_path and the
// `drift` it invokes used another would produce reports about two different vaults.
var (
	cfgOnce sync.Once
	cfgVal  *config.Config
	cfgErr  error
)

func loadConfig() (*config.Config, error) {
	cfgOnce.Do(func() { cfgVal, cfgErr = config.Load(config.Options{}) })
	return cfgVal, cfgErr
}

// resolveVault applies the flag-over-config-over-cwd order. An explicit --vault always
// wins, because a user who typed a path is answering the question directly.
//
// The error path matters more than the happy one: a config that assigns a model engine
// to recall, write or index fails here, before any command does work. That is the
// invariant's enforcement point for the whole binary — refuse to start, never silently
// override.
func resolveVault(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	if cfg.VaultPath != "" {
		return cfg.VaultPath, nil
	}
	return ".", nil
}

// vaultOrExit is the shape every subcommand wants: a path, or a printed error and a
// non-zero code it can return directly.
func vaultOrExit(name, flagVal string) (string, int) {
	v, err := resolveVault(flagVal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge %s: %v\n", name, err)
		return "", 2
	}
	return v, 0
}

// configOrExit is vaultOrExit's sibling for commands that need more than the path.
func configOrExit(name string) (*config.Config, int) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge %s: %v\n", name, err)
		return nil, 2
	}
	return cfg, 0
}
