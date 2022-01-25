package validation

import (
	"errors"
	"fmt"

	"github.com/toppr-systems/dops/dops"
)

// ValidateConfig performs validation on the Config object
func ValidateConfig(cfg *dops.Config) error {
	// TODO: return field.ErrorList instead?
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return fmt.Errorf("unsupported log format: %s", cfg.LogFormat)
	}
	if len(cfg.Sources) == 0 {
		return errors.New("no sources specified")
	}
	if cfg.Provider == "" {
		return errors.New("no provider specified")
	}

	if len(cfg.TXTPrefix) > 0 && len(cfg.TXTSuffix) > 0 {
		return errors.New("txt-prefix and txt-suffix are mutually exclusive")
	}

	return nil
}
