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
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchExtensionTF struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

type ElasticsearchExtensionsTF types.Set

type ElasticsearchExtension struct {
	Name    string `tfsdk:"name"`
	Type    string `tfsdk:"type"`
	Version string `tfsdk:"version"`
	Url     string `tfsdk:"url"`
}

type ElasticsearchExtensions []ElasticsearchExtension

func readElasticsearchExtensions(in *models.ElasticsearchConfiguration) (ElasticsearchExtensions, error) {
	if len(in.UserBundles) == 0 && len(in.UserPlugins) == 0 {
		return nil, nil
	}

	extensions := make(ElasticsearchExtensions, 0, len(in.UserBundles)+len(in.UserPlugins))

	for _, model := range in.UserBundles {
		extension, err := readFromUserBundle(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	for _, model := range in.UserPlugins {
		extension, err := readFromUserPlugin(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	return extensions, nil
}

func (extensions ElasticsearchExtensionsTF) Payload(ctx context.Context, es *models.ElasticsearchConfiguration) diag.Diagnostics {
	for _, elem := range extensions.Elems {
		var extension ElasticsearchExtensionTF

		if diags := tfsdk.ValueAs(ctx, elem, &extension); diags.HasError() {
			return diags
		}

		version := extension.Version.Value
		url := extension.Url.Value
		name := extension.Name.Value

		if extension.Type.Value == "bundle" {
			es.UserBundles = append(es.UserBundles, &models.ElasticsearchUserBundle{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}

		if extension.Type.Value == "plugin" {
			es.UserPlugins = append(es.UserPlugins, &models.ElasticsearchUserPlugin{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}
	}
	return nil
}

func readFromUserBundle(in *models.ElasticsearchUserBundle) (*ElasticsearchExtension, error) {
	var ext ElasticsearchExtension

	ext.Type = "bundle"

	if in.ElasticsearchVersion == nil {
		return nil, missingField("ElasticsearchUserBundle.ElasticsearchVersion")
	}
	ext.Version = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, missingField("ElasticsearchUserBundle.URL")
	}
	ext.Url = *in.URL

	if in.Name == nil {
		return nil, missingField("ElasticsearchUserBundle.Name")
	}
	ext.Name = *in.Name

	return &ext, nil
}

func readFromUserPlugin(in *models.ElasticsearchUserPlugin) (*ElasticsearchExtension, error) {
	var ext ElasticsearchExtension

	ext.Type = "plugin"

	if in.ElasticsearchVersion == nil {
		return nil, missingField("ElasticsearchUserPlugin.ElasticsearchVersion")
	}
	ext.Version = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, missingField("ElasticsearchUserPlugin.URL")
	}
	ext.Url = *in.URL

	if in.Name == nil {
		return nil, missingField("ElasticsearchUserPlugin.Name")
	}
	ext.Name = *in.Name

	return &ext, nil
}
