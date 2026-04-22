package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/rackn/provision/v4/api"
)

type Config struct {
	token    string
	username string
	password string
	endpoint string

	session *api.Client
}

func (c *Config) validateAndConnect(ctx context.Context) error {
	tflog.Debug(ctx, "Configuring the DRP API client")

	if c.session != nil {
		return nil
	}
	var err error
	if c.token != "" {
		c.session, err = api.TokenSession(c.endpoint, c.token)
	} else {
		c.session, err = api.UserSession(c.endpoint, c.username, c.password)
	}
	if err != nil {
		tflog.Error(ctx, "Error creating session", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error creating session: %w", err)
	}

	return nil
}
