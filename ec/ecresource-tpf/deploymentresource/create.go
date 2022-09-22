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

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var config Deployment
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan Deployment
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//
	// model, err := plan.ToModel(plan.Id.Value)

	// // deploymentResource, errors := Create(ctx, r.provider.GetClient(), &cfg, &plan)

	// if len(errors) > 0 {
	// 	for _, err := range errors {
	// 		resp.Diagnostics.AddError(
	// 			"Cannot create deployment resource",
	// 			err.Error(),
	// 		)
	// 	}
	// 	return
	// }

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.CreateExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	// data.Id = types.String{Value: "example-id"}

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	// tflog.Trace(ctx, "created a resource")

	// diags = resp.State.Set(ctx, &deploymentResource)
	// resp.Diagnostics.Append(diags...)
}

// func create(ctx context.Context, client *api.API, plan Deployment, config Deployment) (*Deployment, error) {
// 	reqID := deploymentapi.RequestID(plan.RequestId.Value)

// 	req, err := plan.Model(d, client)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}

// 	res, err := deploymentapi.Create(deploymentapi.CreateParams{
// 		API:       client,
// 		RequestID: reqID,
// 		Request:   req,
// 		Overrides: &deploymentapi.PayloadOverrides{
// 			Name:    d.Get("name").(string),
// 			Version: d.Get("version").(string),
// 			Region:  d.Get("region").(string),
// 		},
// 	})
// 	if err != nil {
// 		merr := multierror.NewPrefixed("failed creating deployment", err)
// 		return diag.FromErr(merr.Append(newCreationError(reqID)))
// 	}

// 	if err := WaitForPlanCompletion(client, *res.ID); err != nil {
// 		merr := multierror.NewPrefixed("failed tracking create progress", err)
// 		return diag.FromErr(merr.Append(newCreationError(reqID)))
// 	}

// 	d.SetId(*res.ID)

// 	// Since before the deployment has been read, there's no real state
// 	// persisted, it'd better to handle each of the errors by appending
// 	// it to the `diag.Diagnostics` since it has support for it.
// 	var diags diag.Diagnostics
// 	if err := handleRemoteClusters(d, client); err != nil {
// 		diags = append(diags, diag.FromErr(err)...)
// 	}

// 	if diag := readResource(ctx, d, meta); diag != nil {
// 		diags = append(diags, diags...)
// 	}

// 	if err := parseCredentials(d, res.Resources); err != nil {
// 		diags = append(diags, diag.FromErr(err)...)
// 	}

// 	return diags
// }
