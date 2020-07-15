/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_locale_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/locale_services"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyBgpConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyBgpConfigUpdate,
		Read:   resourceNsxtPolicyBgpConfigRead,
		Update: resourceNsxtPolicyBgpConfigUpdate,
		Delete: resourceNsxtPolicyBgpConfigDelete,

		Schema: getPolicyBGPConfigSchema(),
	}
}

func findTier0LocaleService(connector *client.RestConnector, gwID string, sitePath string) (string, error) {
	localeServices, err := listPolicyTier0GatewayLocaleServices(connector, gwID, true)
	if err != nil {
		return "", err
	}
	return getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices, sitePath, gwID)
}

func resourceNsxtPolicyBgpConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	gwID := getPolicyIDFromPath(gwPath)
	sitePath := d.Get("site_path").(string)
	if !isPolicyGlobalManager(m) {
		return fmt.Errorf("This resource is not supported for local manager")
	}

	serviceID, err := findTier0LocaleService(connector, gwID, sitePath)
	if err != nil {
		return err
	}
	client := gm_locale_services.NewDefaultBgpClient(connector)
	gmObj, err := client.Get(gwID, serviceID)
	if err != nil {
		return handleReadError(d, "BGP Config", serviceID, err)
	}
	lmObj, convErr := convertModelBindingType(gmObj, gm_model.BgpRoutingConfigBindingType(), model.BgpRoutingConfigBindingType())
	if convErr != nil {
		return convErr
	}
	lmRoutingConfig := lmObj.(model.BgpRoutingConfig)

	data := initPolicyTier0BGPConfigMap(&lmRoutingConfig)

	for key, value := range data {
		d.Set(key, value)
	}

	return nil
}

func resourceNsxtPolicyBgpConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	gwID := getPolicyIDFromPath(gwPath)
	sitePath := d.Get("site_path").(string)
	ecmp := d.Get("ecmp").(bool)
	enabled := d.Get("enabled").(bool)
	interSrIbgp := d.Get("inter_sr_ibgp").(bool)
	localAsNum := d.Get("local_as_num").(string)
	multipathRelax := d.Get("multipath_relax").(bool)
	restartMode := d.Get("graceful_restart_mode").(string)
	restartTimer := int64(d.Get("graceful_restart_timer").(int))
	staleTimer := int64(d.Get("graceful_restart_stale_route_timer").(int))
	//tags := getPolicyTagsFromSchema(d)

	serviceID, err := findTier0LocaleService(connector, gwID, sitePath)
	if err != nil {
		return err
	}
	var aggregationStructs []gm_model.RouteAggregationEntry
	routeAggregations := d.Get("route_aggregation").([]interface{})
	if len(routeAggregations) > 0 {
		for _, agg := range routeAggregations {
			data := agg.(map[string]interface{})
			prefix := data["prefix"].(string)
			summary := data["summary_only"].(bool)
			elem := gm_model.RouteAggregationEntry{
				Prefix:      &prefix,
				SummaryOnly: &summary,
			}

			aggregationStructs = append(aggregationStructs, elem)
		}
	}

	restartTimerStruct := gm_model.BgpGracefulRestartTimer{
		RestartTimer:    &restartTimer,
		StaleRouteTimer: &staleTimer,
	}

	restartConfigStruct := gm_model.BgpGracefulRestartConfig{
		Mode:  &restartMode,
		Timer: &restartTimerStruct,
	}

	routeStruct := gm_model.BgpRoutingConfig{
		Ecmp:              &ecmp,
		Enabled:           &enabled,
		RouteAggregations: aggregationStructs,
		//		Tags:                  tags,
		InterSrIbgp:           &interSrIbgp,
		LocalAsNum:            &localAsNum,
		MultipathRelax:        &multipathRelax,
		GracefulRestartConfig: &restartConfigStruct,
	}

	client := gm_locale_services.NewDefaultBgpClient(connector)
	err = client.Patch(gwID, serviceID, routeStruct)
	if err != nil {
		return handleUpdateError("BgpRoutingConfig", serviceID, err)
	}
	d.SetId("bgp")

	return resourceNsxtPolicyBgpConfigRead(d, m)
}

func resourceNsxtPolicyBgpConfigDelete(d *schema.ResourceData, m interface{}) error {
	// TODO: revert to default
	return nil
}
