// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationsServerConfigTF struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type IntegrationsServerConfig struct {
	DockerImage              *string `tfsdk:"docker_image"`
	DebugEnabled             *bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         *string `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson *string `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         *string `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml *string `tfsdk:"user_settings_override_yaml"`
}

func readIntegrationsServerConfig(in *models.IntegrationsServerConfiguration) (*IntegrationsServerConfig, error) {
	var cfg IntegrationsServerConfig

	if in.UserSettingsYaml != "" {
		cfg.UserSettingsYaml = &in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		cfg.UserSettingsOverrideYaml = &in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsJson = ec.String(string(b))
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsOverrideJson = ec.String(string(b))
		}
	}

	if in.DockerImage != "" {
		cfg.DockerImage = &in.DockerImage
	}

	if in.SystemSettings != nil {
		if in.SystemSettings.DebugEnabled != nil {
			cfg.DebugEnabled = in.SystemSettings.DebugEnabled
		}
	}

	if cfg == (IntegrationsServerConfig{}) {
		return nil, nil
	}

	return &cfg, nil
}

func payloadIntegrationsServerConfig(ctx context.Context, cfg types.Object, res *models.IntegrationsServerConfiguration) diag.Diagnostics {
	if cfg.IsNull() {
		return nil
	}

	var diags diag.Diagnostics

	var cfgTF IntegrationsServerConfigTF

	if diags = tfsdk.ValueAs(ctx, cfg, &cfgTF); diags.HasError() {
		return diags
	}

	if !cfgTF.DebugEnabled.IsNull() {
		if res.SystemSettings == nil {
			res.SystemSettings = &models.IntegrationsServerSystemSettings{}
		}
		res.SystemSettings.DebugEnabled = &cfgTF.DebugEnabled.Value
	}

	if cfgTF.UserSettingsJson.Value != "" {
		if err := json.Unmarshal([]byte(cfgTF.UserSettingsJson.Value), &res.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding IntegrationsServer user_settings_json", err.Error())
		}
	}

	if cfgTF.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(cfgTF.UserSettingsOverrideJson.Value), &res.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding IntegrationsServer user_settings_override_json", err.Error())
		}
	}

	if !cfgTF.UserSettingsYaml.IsNull() {
		res.UserSettingsYaml = cfgTF.UserSettingsYaml.Value
	}

	if !cfgTF.UserSettingsOverrideYaml.IsNull() {
		res.UserSettingsOverrideYaml = cfgTF.UserSettingsOverrideYaml.Value
	}

	if !cfgTF.DockerImage.IsNull() {
		res.DockerImage = cfgTF.DockerImage.Value
	}

	return diags
}
