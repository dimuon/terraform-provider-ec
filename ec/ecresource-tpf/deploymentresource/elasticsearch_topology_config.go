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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyConfig struct {
	Plugins                  types.Set    `tfsdk:"plugins"`
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

func (c *ElasticsearchTopologyConfig) fromModel(cfg *models.ElasticsearchConfiguration) error {
	if cfg == nil {
		return nil
	}

	if len(cfg.EnabledBuiltInPlugins) > 0 {
		c.Plugins.ElemType = types.StringType
		c.Plugins.Elems = make([]attr.Value, 0, len(cfg.EnabledBuiltInPlugins))
		for _, plugin := range cfg.EnabledBuiltInPlugins {
			c.Plugins.Elems = append(c.Plugins.Elems, types.String{Value: plugin})
		}
	}

	if cfg.UserSettingsYaml != "" {
		c.UserSettingsYaml.Value = cfg.UserSettingsYaml
	}

	if cfg.UserSettingsOverrideYaml != "" {
		c.UserSettingsOverrideYaml.Value = cfg.UserSettingsOverrideYaml
	}

	if o := cfg.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			c.UserSettingsJson.Value = string(b)
		}
	}

	if o := cfg.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			c.UserSettingsOverrideJson.Value = string(b)
		}
	}

	if cfg.DockerImage != "" {
		c.DockerImage.Value = cfg.DockerImage
	}

	return nil
}
