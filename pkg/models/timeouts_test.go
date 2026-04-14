package models

import (
	"testing"
	"time"
)

func TestEffectiveTimeoutsDefault(t *testing.T) {
	cfg := &Config{}

	if got := cfg.EffectiveDBConnectTimeout(); got != DefaultDBConnectTimeout {
		t.Fatalf("EffectiveDBConnectTimeout() = %v, want %v", got, DefaultDBConnectTimeout)
	}
	if got := cfg.EffectiveFTPConnectTimeout(); got != DefaultFTPConnectTimeout {
		t.Fatalf("EffectiveFTPConnectTimeout() = %v, want %v", got, DefaultFTPConnectTimeout)
	}
	if got := cfg.EffectivePipelineLoadTimeout(); got != DefaultPipelineLoadTimeout {
		t.Fatalf("EffectivePipelineLoadTimeout() = %v, want %v", got, DefaultPipelineLoadTimeout)
	}
	if got := cfg.EffectiveCLIRunTimeout(); got != DefaultCLIRunTimeout {
		t.Fatalf("EffectiveCLIRunTimeout() = %v, want %v", got, DefaultCLIRunTimeout)
	}
	if got := cfg.EffectiveWebhookReportHTTPTimeout(); got != DefaultWebhookReportHTTPTimeout {
		t.Fatalf("EffectiveWebhookReportHTTPTimeout() = %v, want %v", got, DefaultWebhookReportHTTPTimeout)
	}
	if got := cfg.EffectiveWebhookReportResultWaitTimeout(); got != DefaultWebhookReportResultWaitTimeout {
		t.Fatalf("EffectiveWebhookReportResultWaitTimeout() = %v, want %v", got, DefaultWebhookReportResultWaitTimeout)
	}
	if got := cfg.EffectiveHTTPReadHeaderTimeout(); got != DefaultHTTPReadHeaderTimeout {
		t.Fatalf("EffectiveHTTPReadHeaderTimeout() = %v, want %v", got, DefaultHTTPReadHeaderTimeout)
	}
	if got := cfg.EffectiveHTTPReadTimeout(); got != DefaultHTTPReadTimeout {
		t.Fatalf("EffectiveHTTPReadTimeout() = %v, want %v", got, DefaultHTTPReadTimeout)
	}
	if got := cfg.EffectiveHTTPWriteTimeout(); got != DefaultHTTPWriteTimeout {
		t.Fatalf("EffectiveHTTPWriteTimeout() = %v, want %v", got, DefaultHTTPWriteTimeout)
	}
	if got := cfg.EffectiveHTTPIdleTimeout(); got != DefaultHTTPIdleTimeout {
		t.Fatalf("EffectiveHTTPIdleTimeout() = %v, want %v", got, DefaultHTTPIdleTimeout)
	}
	if got := cfg.EffectiveShutdownTimeout(); got != DefaultShutdownTimeout {
		t.Fatalf("EffectiveShutdownTimeout() = %v, want %v", got, DefaultShutdownTimeout)
	}
}

func TestEffectiveTimeoutsOverride(t *testing.T) {
	cfg := &Config{
		DBConnectTimeout:               7 * time.Second,
		FTPConnectTimeout:              8 * time.Second,
		PipelineLoadTimeout:            9 * time.Minute,
		CLIRunTimeout:                  10 * time.Minute,
		WebhookReportHTTPTimeout:       11 * time.Second,
		WebhookReportResultWaitTimeout: 12 * time.Second,
		HTTPReadHeaderTimeout:          13 * time.Second,
		HTTPReadTimeout:                14 * time.Second,
		HTTPWriteTimeout:               15 * time.Second,
		HTTPIdleTimeout:                16 * time.Second,
		ShutdownTimeout:                17 * time.Second,
	}

	if got := cfg.EffectiveDBConnectTimeout(); got != 7*time.Second {
		t.Fatalf("EffectiveDBConnectTimeout() = %v, want 7s", got)
	}
	if got := cfg.EffectiveFTPConnectTimeout(); got != 8*time.Second {
		t.Fatalf("EffectiveFTPConnectTimeout() = %v, want 8s", got)
	}
	if got := cfg.EffectivePipelineLoadTimeout(); got != 9*time.Minute {
		t.Fatalf("EffectivePipelineLoadTimeout() = %v, want 9m", got)
	}
	if got := cfg.EffectiveCLIRunTimeout(); got != 10*time.Minute {
		t.Fatalf("EffectiveCLIRunTimeout() = %v, want 10m", got)
	}
	if got := cfg.EffectiveWebhookReportHTTPTimeout(); got != 11*time.Second {
		t.Fatalf("EffectiveWebhookReportHTTPTimeout() = %v, want 11s", got)
	}
	if got := cfg.EffectiveWebhookReportResultWaitTimeout(); got != 12*time.Second {
		t.Fatalf("EffectiveWebhookReportResultWaitTimeout() = %v, want 12s", got)
	}
	if got := cfg.EffectiveHTTPReadHeaderTimeout(); got != 13*time.Second {
		t.Fatalf("EffectiveHTTPReadHeaderTimeout() = %v, want 13s", got)
	}
	if got := cfg.EffectiveHTTPReadTimeout(); got != 14*time.Second {
		t.Fatalf("EffectiveHTTPReadTimeout() = %v, want 14s", got)
	}
	if got := cfg.EffectiveHTTPWriteTimeout(); got != 15*time.Second {
		t.Fatalf("EffectiveHTTPWriteTimeout() = %v, want 15s", got)
	}
	if got := cfg.EffectiveHTTPIdleTimeout(); got != 16*time.Second {
		t.Fatalf("EffectiveHTTPIdleTimeout() = %v, want 16s", got)
	}
	if got := cfg.EffectiveShutdownTimeout(); got != 17*time.Second {
		t.Fatalf("EffectiveShutdownTimeout() = %v, want 17s", got)
	}
}
