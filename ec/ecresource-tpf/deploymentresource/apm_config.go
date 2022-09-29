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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApmConfigTF struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ApmConfig struct {
	DockerImage              string `tfsdk:"docker_image"`
	DebugEnabled             bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         string `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson string `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         string `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml string `tfsdk:"user_settings_override_yaml"`
}

type ApmConfigsTF types.List

type ApmConfigs []ApmConfig

func ReadApmConfigs(in *models.ApmConfiguration) (ApmConfigs, error) {
	var cfg ApmConfig

	cfg.UserSettingsYaml = in.UserSettingsYaml

	cfg.UserSettingsOverrideYaml = in.UserSettingsOverrideYaml

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsJson = string(b)
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsOverrideJson = string(b)
		}
	}

	cfg.DockerImage = in.DockerImage

	if in.SystemSettings != nil {
		if in.SystemSettings.DebugEnabled != nil {
			cfg.DebugEnabled = *in.SystemSettings.DebugEnabled
		}
	}

	if cfg == (ApmConfig{}) {
		return nil, nil
	}

	return []ApmConfig{cfg}, nil
}

func (cfgs ApmConfigsTF) Payload(ctx context.Context, model *models.ApmConfiguration) diag.Diagnostics {
	for _, elem := range cfgs.Elems {
		var cfg ApmConfigTF
		if diags := tfsdk.ValueAs(ctx, elem, &cfg); diags.HasError() {
			return diags
		}
		if !cfg.DebugEnabled.IsNull() {
			if model.SystemSettings == nil {
				model.SystemSettings = &models.ApmSystemSettings{}
			}
			model.SystemSettings.DebugEnabled = &cfg.DebugEnabled.Value
		}

		var diags diag.Diagnostics
		if cfg.UserSettingsJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
				diags.AddError("failed expanding apm user_settings_json", err.Error())
				return diags
			}
		}

		if cfg.UserSettingsOverrideJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
				diags.AddError("failed expanding apm user_settings_override_json", err.Error())
				return diags
			}
		}

		if !cfg.UserSettingsYaml.IsNull() {
			model.UserSettingsYaml = cfg.UserSettingsYaml.Value
		}

		if !cfg.UserSettingsOverrideYaml.IsNull() {
			model.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
		}

		if !cfg.DockerImage.IsNull() {
			model.DockerImage = cfg.DockerImage.Value
		}
	}

	return nil
}
