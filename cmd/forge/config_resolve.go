package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// The chain is read once per process.
var (
	cfgOnce sync.Once
	cfgVal  *config.Config
	cfgErr  error
)

func loadConfig() (*config.Config, error) {
	cfgOnce.Do(func() { cfgVal, cfgErr = config.Load(config.Options{}) })
	return cfgVal, cfgErr
}

// resolveVault applies the flag-over-config-over-cwd order.
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
