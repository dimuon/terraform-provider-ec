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
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchConfigs []ElasticsearchConfig

func (cfgs *ElasticsearchConfigs) fromModel(in *models.ElasticsearchConfiguration) error {
	var cfg ElasticsearchConfig
	cfg.fromModel(in)

	*cfgs = nil

	if !cfg.isEmpty() {
		*cfgs = []ElasticsearchConfig{cfg}
	}
	return nil
}

type ElasticsearchConfig struct {
	Plugins                  []string     `tfsdk:"plugins"`
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

func (c *ElasticsearchConfig) isEmpty() bool {
	return reflect.ValueOf(*c).IsZero()
}

func (c *ElasticsearchConfig) fromModel(in *models.ElasticsearchConfiguration) error {
	if in == nil {
		return nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		c.Plugins = append(c.Plugins, in.EnabledBuiltInPlugins...)
	}

	if in.UserSettingsYaml != "" {
		c.UserSettingsYaml.Value = in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		c.UserSettingsOverrideYaml.Value = in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			c.UserSettingsJson.Value = string(b)
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			c.UserSettingsOverrideJson.Value = string(b)
		}
	}

	if in.DockerImage != "" {
		c.DockerImage.Value = in.DockerImage
	}

	return nil
}
