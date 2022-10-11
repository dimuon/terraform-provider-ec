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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnterpriseSearchConfigTF struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type EnterpriseSearchConfig struct {
	DockerImage              *string `tfsdk:"docker_image"`
	UserSettingsJson         *string `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson *string `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         *string `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml *string `tfsdk:"user_settings_override_yaml"`
}

func readEnterpriseSearchConfig(in *models.EnterpriseSearchConfiguration) (*EnterpriseSearchConfig, error) {
	var cfg EnterpriseSearchConfig

	if in == nil {
		return nil, nil
	}

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

	if cfg == (EnterpriseSearchConfig{}) {
		return nil, nil
	}

	return &cfg, nil
}

func (cfg EnterpriseSearchConfigTF) Payload(ctx context.Context, res *models.EnterpriseSearchConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics

	if cfg.UserSettingsJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &res.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding enterprise_search user_settings_json", err.Error())
		}
	}
	if cfg.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &res.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding enterprise_search user_settings_override_json", err.Error())
		}
	}
	if !cfg.UserSettingsYaml.IsNull() {
		res.UserSettingsYaml = cfg.UserSettingsYaml.Value
	}
	if !cfg.UserSettingsOverrideYaml.IsNull() {
		res.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
	}

	if !cfg.DockerImage.IsNull() {
		res.DockerImage = cfg.DockerImage.Value
	}

	return diags
}
