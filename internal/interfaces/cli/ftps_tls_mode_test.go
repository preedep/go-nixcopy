package cli

import (
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

// ---- CLI flag: --source-tls-mode ----

func TestApplyCliFlags_SourceFTPS_TLSMode_Explicit(t *testing.T) {
	resetTransferFlags()
	sourceType = "ftps"
	sourceHost = "ftps.example.com"
	sourceTLSMode = "explicit"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
	}
	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("TLSMode = %q, want explicit", cfg.Source.FTPS.TLSMode)
	}
}

func TestApplyCliFlags_SourceFTPS_TLSMode_Implicit(t *testing.T) {
	resetTransferFlags()
	sourceType = "ftps"
	sourceHost = "ftps.example.com"
	sourceTLSMode = "implicit"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
	}
	if cfg.Source.FTPS.TLSMode != "implicit" {
		t.Errorf("TLSMode = %q, want implicit", cfg.Source.FTPS.TLSMode)
	}
}

func TestApplyCliFlags_SourceFTPS_TLSMode_Empty_DoesNotOverride(t *testing.T) {
	// When --source-tls-mode is not set, a value already in the config (e.g. from
	// a config file or env var) must not be cleared.
	resetTransferFlags()
	sourceType = "ftps"
	sourceHost = "ftps.example.com"
	// sourceTLSMode intentionally left ""

	cfg := config.DefaultConfig()
	cfg.Source.FTPS = &config.FTPSConfig{Host: "ftps.example.com", TLSMode: "explicit"}
	applyCliFlags(cfg)

	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("TLSMode = %q, want explicit (empty CLI flag must not clear config-file value)", cfg.Source.FTPS.TLSMode)
	}
}

// ---- CLI flag: --dest-tls-mode ----

func TestApplyCliFlags_DestFTPS_TLSMode_Explicit(t *testing.T) {
	resetTransferFlags()
	destType = "ftps"
	destHost = "ftps-dest.example.com"
	destTLSMode = "explicit"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.FTPS == nil {
		t.Fatal("Destination.FTPS is nil")
	}
	if cfg.Destination.FTPS.TLSMode != "explicit" {
		t.Errorf("TLSMode = %q, want explicit", cfg.Destination.FTPS.TLSMode)
	}
}

func TestApplyCliFlags_DestFTPS_TLSMode_Implicit(t *testing.T) {
	resetTransferFlags()
	destType = "ftps"
	destHost = "ftps-dest.example.com"
	destTLSMode = "implicit"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.FTPS == nil {
		t.Fatal("Destination.FTPS is nil")
	}
	if cfg.Destination.FTPS.TLSMode != "implicit" {
		t.Errorf("TLSMode = %q, want implicit", cfg.Destination.FTPS.TLSMode)
	}
}

// ---- validateConfig: TLS mode validation ----

func TestValidateConfig_SourceFTPS_InvalidTLSMode(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{Host: "ftps.example.com", TLSMode: "starttls"},
		},
	}
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid source TLS mode")
	}
}

func TestValidateConfig_DestFTPS_InvalidTLSMode(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{Host: "sftp.example.com", Username: "user"},
		},
		Destination: config.DestinationConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{Host: "ftps-dest.example.com", TLSMode: "none"},
		},
	}
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid destination TLS mode")
	}
}

func TestValidateConfig_SourceFTPS_TLSMode_Explicit_Valid(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{Host: "ftps.example.com", TLSMode: "explicit"},
		},
		Destination: config.DestinationConfig{
			Type:  config.StorageTypeLocal,
			Local: &config.LocalConfig{BasePath: "/tmp"},
		},
	}
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for valid explicit TLS mode: %v", err)
	}
}

func TestValidateConfig_SourceFTPS_TLSMode_Implicit_Valid(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{Host: "ftps.example.com", TLSMode: "implicit"},
		},
		Destination: config.DestinationConfig{
			Type:  config.StorageTypeLocal,
			Local: &config.LocalConfig{BasePath: "/tmp"},
		},
	}
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for valid implicit TLS mode: %v", err)
	}
}

func TestValidateConfig_SourceFTPS_TLSMode_Empty_Valid(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{Host: "ftps.example.com", TLSMode: ""},
		},
		Destination: config.DestinationConfig{
			Type:  config.StorageTypeLocal,
			Local: &config.LocalConfig{BasePath: "/tmp"},
		},
	}
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for empty TLS mode (defaults to implicit): %v", err)
	}
}

// ---- Env var: NIXCOPY_SOURCE/DEST_TLS_MODE ----

func TestEnvVar_SourceFTPS_TLSMode_Explicit(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TLS_MODE", "explicit")

	resetTransferFlags()
	sourceType = "ftps"

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
	}
	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("TLSMode = %q, want explicit (from NIXCOPY_SOURCE_TLS_MODE)", cfg.Source.FTPS.TLSMode)
	}
}

func TestEnvVar_DestFTPS_TLSMode_Implicit(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_TLS_MODE", "implicit")

	resetTransferFlags()
	destType = "ftps"

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Destination.FTPS == nil {
		t.Fatal("Destination.FTPS is nil")
	}
	if cfg.Destination.FTPS.TLSMode != "implicit" {
		t.Errorf("TLSMode = %q, want implicit (from NIXCOPY_DEST_TLS_MODE)", cfg.Destination.FTPS.TLSMode)
	}
}

// ---- Precedence: CLI flag beats env var ----

func TestCLIFlag_TLSMode_WinsOver_EnvVar(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TLS_MODE", "implicit")

	resetTransferFlags()
	sourceType = "ftps"
	sourceTLSMode = "explicit" // CLI must beat env var

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
	}
	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("TLSMode = %q, want explicit (CLI flag must win over NIXCOPY_SOURCE_TLS_MODE=implicit)", cfg.Source.FTPS.TLSMode)
	}
}
