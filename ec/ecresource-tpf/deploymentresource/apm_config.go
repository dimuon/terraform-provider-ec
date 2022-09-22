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
	"encoding/json"
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApmConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ApmConfigs []*ApmConfig

func NewApmConfigs(in *models.ApmConfiguration) (ApmConfigs, error) {
	var cfg ApmConfig

	if in == nil {
		return nil, nil
	}

	cfg.UserSettingsYaml.Value = in.UserSettingsYaml

	cfg.UserSettingsOverrideYaml.Value = in.UserSettingsOverrideYaml

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsJson.Value = string(b)
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsOverrideJson.Value = string(b)
		}
	}

	cfg.DockerImage.Value = in.DockerImage

	if in.SystemSettings != nil {
		if in.SystemSettings.DebugEnabled != nil {
			cfg.DebugEnabled.Value = *in.SystemSettings.DebugEnabled
		}
	}

	if cfg != (ApmConfig{}) {
		return []*ApmConfig{&cfg}, nil
	}

	return nil, nil
}

func (cfgs ApmConfigs) Payload(model *models.ApmConfiguration) error {
	for _, cfg := range cfgs {
		if !cfg.DebugEnabled.IsNull() {
			if model.SystemSettings == nil {
				model.SystemSettings = &models.ApmSystemSettings{}
			}
			model.SystemSettings.DebugEnabled = &cfg.DebugEnabled.Value
		}

		if cfg.UserSettingsJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
				return fmt.Errorf("failed expanding apm user_settings_json: %w", err)
			}
		}

		if cfg.UserSettingsOverrideJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
				return fmt.Errorf("failed expanding apm user_settings_override_json: %w", err)
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
