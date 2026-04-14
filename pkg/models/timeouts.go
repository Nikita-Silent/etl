package models

import "time"

const (
	DefaultDBConnectTimeout               = 5 * time.Second
	DefaultFTPConnectTimeout              = 5 * time.Second
	DefaultPipelineLoadTimeout            = 60 * time.Minute
	DefaultCLIRunTimeout                  = 30 * time.Minute
	DefaultOperationStaleTimeout          = 2 * time.Hour
	DefaultWebhookReportHTTPTimeout       = 30 * time.Second
	DefaultWebhookReportResultWaitTimeout = 5 * time.Second
	DefaultHTTPReadHeaderTimeout          = 5 * time.Second
	DefaultHTTPReadTimeout                = 15 * time.Second
	DefaultHTTPWriteTimeout               = 30 * time.Second
	DefaultHTTPIdleTimeout                = 60 * time.Second
	DefaultShutdownTimeout                = 30 * time.Second
)

func (c *Config) EffectiveDBConnectTimeout() time.Duration {
	if c == nil || c.DBConnectTimeout <= 0 {
		return DefaultDBConnectTimeout
	}
	return c.DBConnectTimeout
}

func (c *Config) EffectiveFTPConnectTimeout() time.Duration {
	if c == nil || c.FTPConnectTimeout <= 0 {
		return DefaultFTPConnectTimeout
	}
	return c.FTPConnectTimeout
}

func (c *Config) EffectivePipelineLoadTimeout() time.Duration {
	if c == nil || c.PipelineLoadTimeout <= 0 {
		return DefaultPipelineLoadTimeout
	}
	return c.PipelineLoadTimeout
}

func (c *Config) EffectiveCLIRunTimeout() time.Duration {
	if c == nil || c.CLIRunTimeout <= 0 {
		return DefaultCLIRunTimeout
	}
	return c.CLIRunTimeout
}

func (c *Config) EffectiveOperationStaleTimeout() time.Duration {
	if c == nil || c.OperationStaleTimeout <= 0 {
		return DefaultOperationStaleTimeout
	}
	return c.OperationStaleTimeout
}

func (c *Config) EffectiveWebhookReportHTTPTimeout() time.Duration {
	if c == nil || c.WebhookReportHTTPTimeout <= 0 {
		return DefaultWebhookReportHTTPTimeout
	}
	return c.WebhookReportHTTPTimeout
}

func (c *Config) EffectiveWebhookReportResultWaitTimeout() time.Duration {
	if c == nil || c.WebhookReportResultWaitTimeout <= 0 {
		return DefaultWebhookReportResultWaitTimeout
	}
	return c.WebhookReportResultWaitTimeout
}

func (c *Config) EffectiveHTTPReadHeaderTimeout() time.Duration {
	if c == nil || c.HTTPReadHeaderTimeout <= 0 {
		return DefaultHTTPReadHeaderTimeout
	}
	return c.HTTPReadHeaderTimeout
}

func (c *Config) EffectiveHTTPReadTimeout() time.Duration {
	if c == nil || c.HTTPReadTimeout <= 0 {
		return DefaultHTTPReadTimeout
	}
	return c.HTTPReadTimeout
}

func (c *Config) EffectiveHTTPWriteTimeout() time.Duration {
	if c == nil || c.HTTPWriteTimeout <= 0 {
		return DefaultHTTPWriteTimeout
	}
	return c.HTTPWriteTimeout
}

func (c *Config) EffectiveHTTPIdleTimeout() time.Duration {
	if c == nil || c.HTTPIdleTimeout <= 0 {
		return DefaultHTTPIdleTimeout
	}
	return c.HTTPIdleTimeout
}

func (c *Config) EffectiveShutdownTimeout() time.Duration {
	if c == nil || c.ShutdownTimeout <= 0 {
		return DefaultShutdownTimeout
	}
	return c.ShutdownTimeout
}
