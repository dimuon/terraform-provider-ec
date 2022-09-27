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

type ElasticsearchExtensions types.Set

func (extensions ElasticsearchExtensions) Read(ctx context.Context, in *models.ElasticsearchConfiguration) diag.Diagnostics {
	if len(in.UserBundles) == 0 && len(in.UserPlugins) == 0 {
		return nil
	}

	exts := make([]ElasticsearchExtension, 0, len(in.UserBundles)+len(in.UserPlugins))

	for _, model := range in.UserBundles {
		var extension ElasticsearchExtension

		if diags := extension.ReadFromUserBundle(ctx, model); diags.HasError() {
			return diags
		}

		exts = append(exts, extension)
	}

	for _, model := range in.UserPlugins {
		var extension ElasticsearchExtension

		if diag := extension.ReadFromUserPlugin(ctx, model); diag != nil {
			return nil
		}
		exts = append(exts, extension)
	}

	return tfsdk.ValueFrom(ctx, exts, elasticsearchExtension().Type(), extensions)
}

func (extensions ElasticsearchExtensions) Payload(ctx context.Context, es *models.ElasticsearchConfiguration) diag.Diagnostics {
	for _, elem := range extensions.Elems {
		var extension ElasticsearchExtension

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

type ElasticsearchExtension struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

func (ext ElasticsearchExtension) ReadFromUserBundle(ctx context.Context, in *models.ElasticsearchUserBundle) diag.Diagnostics {
	ext.Type = types.String{Value: "bundle"}

	var diag diag.Diagnostics

	if in.ElasticsearchVersion == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserBundle.ElasticsearchVersion").Error())
		return diag
	}
	ext.Version = types.String{Value: *in.ElasticsearchVersion}

	if in.URL == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserBundle.URL").Error())
		return diag
	}
	ext.Url = types.String{Value: *in.URL}

	if in.Name == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserBundle.Name").Error())
		return diag
	}
	ext.Name = types.String{Value: *in.Name}

	return nil
}

func (ext ElasticsearchExtension) ReadFromUserPlugin(ctx context.Context, in *models.ElasticsearchUserPlugin) diag.Diagnostics {
	ext.Type = types.String{Value: "plugin"}

	var diag diag.Diagnostics

	if in.ElasticsearchVersion == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserPlugin.ElasticsearchVersion").Error())
		return diag
	}
	ext.Version = types.String{Value: *in.ElasticsearchVersion}

	if in.URL == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserPlugin.URL").Error())
		return diag
	}
	ext.Url = types.String{Value: *in.URL}

	if in.Name == nil {
		diag.AddError("Elasticsearch extension read error", missingField("ElasticsearchUserPlugin.Name").Error())
		return diag
	}
	ext.Name = types.String{Value: *in.Name}

	return nil
}
