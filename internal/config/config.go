package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ActionMode string

const (
	ActionModeExec ActionMode = "exec" // takes over terminal (default)
	ActionModeRun  ActionMode = "run"  // background, result in status bar
)

type ActionInput string

const (
	ActionInputJSON     ActionInput = "json"     // todo as JSON (default)
	ActionInputMarkdown ActionInput = "markdown" // todo as markdown
	ActionInputID       ActionInput = "id"       // todo UUID
)

type ActionInputMode string

const (
	ActionInputModeStdin ActionInputMode = "stdin" // pipe to stdin (default)
	ActionInputModeArg   ActionInputMode = "arg"   // prepend to args
)

type Action struct {
	Key       string          `toml:"key"`
	Label     string          `toml:"label"`
	Command   string          `toml:"command"`
	Args      []string        `toml:"args"`
	Mode      ActionMode      `toml:"mode"`
	Input     ActionInput     `toml:"input"`
	InputMode ActionInputMode `toml:"input_mode"`
}

type Config struct {
	Actions []Action `toml:"actions"`
}

// Load reads the config from ~/.config/things-cli/config.toml.
// Returns an empty config if the file does not exist.
func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	path := filepath.Join(home, ".config", "things-cli", "config.toml")
	var cfg Config
	_, err = toml.DecodeFile(path, &cfg)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}

	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

var reservedKeys = map[string]bool{
	"j": true, "k": true, "g": true, "G": true,
	"x": true, "X": true, "e": true, "a": true, "o": true,
	"q": true, "c": true, "C": true,
	"enter": true, "esc": true,
	"up": true, "down": true, "home": true, "end": true,
	"ctrl+x": true, "ctrl+d": true, "ctrl+u": true,
}

func (c *Config) applyDefaults() {
	for i := range c.Actions {
		if c.Actions[i].Mode == "" {
			c.Actions[i].Mode = ActionModeExec
		}
		if c.Actions[i].Input == "" {
			c.Actions[i].Input = ActionInputJSON
		}
		if c.Actions[i].InputMode == "" {
			c.Actions[i].InputMode = ActionInputModeStdin
		}
	}
}

func (c Config) validate() error {
	seen := make(map[string]bool)
	for _, a := range c.Actions {
		if a.Key == "" {
			return fmt.Errorf("action %q: key is required", a.Label)
		}
		if a.Command == "" {
			return fmt.Errorf("action %q: command is required", a.Key)
		}
		if a.Mode != ActionModeExec && a.Mode != ActionModeRun {
			return fmt.Errorf("action %q: invalid mode %q (want exec or run)", a.Key, a.Mode)
		}
		if a.Input != ActionInputJSON && a.Input != ActionInputMarkdown && a.Input != ActionInputID {
			return fmt.Errorf("action %q: invalid input %q (want json, markdown, or id)", a.Key, a.Input)
		}
		if a.InputMode != ActionInputModeStdin && a.InputMode != ActionInputModeArg {
			return fmt.Errorf("action %q: invalid input_mode %q (want stdin or arg)", a.Key, a.InputMode)
		}
		if reservedKeys[a.Key] {
			return fmt.Errorf("action %q: key %q is reserved", a.Label, a.Key)
		}
		if seen[a.Key] {
			return fmt.Errorf("action %q: duplicate key %q", a.Label, a.Key)
		}
		seen[a.Key] = true
	}
	return nil
}
