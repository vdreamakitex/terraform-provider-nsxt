/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gm_locale_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/locale_services"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
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

func resourceNsxtPolicyBgpConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	sitePath := d.Get("site_path").(string)
	var obj model.BgpRoutingConfig
	if isPolicyGlobalManager(m) {
                localeServices, err := listPolicyTier0GatewayLocaleServices(connector, gwID, true)
                if err != nil {
                    return err
                }
                serviceID, err := getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices, sitePath, gwPath)
                if err != nil {
                    return err
                }

		client := gm_locale_services.NewDefaultBgpClient(connector)
		gmObj, err := client.Get(t0ID, serviceID)
		if err != nil {
			return handleReadError(d, "BGP Config", serviceID, err)
		}
		lmObj, convErr := convertModelBindingType(gmObj, gm_model.BgpRoutingConfigBindingType(), model.BgpRoutingConfigBindingType())
		if convErr != nil {
			return convErr
		}
		obj = lmObj.(model.BgpRoutingConfig)
	} else {
                getPolicyTier0GatewayLocaleServiceWithEdgeCluster(gwPath, connector)
		client := locale_services.NewDefaultBgpClient(connector)
		var err error
		obj, err = client.Get(t0ID, serviceID)
		if err != nil {
			return handleReadError(d, "BGP Config", serviceID, err)
		}

	}
	data := initPolicyTier0BGPConfigMap(&obj)

	for key, value := range data {
		d.Set(key, value)
	}

	return nil
}

func resourceNsxtPolicyBgpConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	ecmp := d.Get("ecmp").(bool)
	enabled := d.Get("enabled").(bool)
	interSrIbgp := d.Get("inter_sr_ibgp").(bool)
	localAsNum := d.Get("local_as_num").(string)
	multipathRelax := d.Get("multipath_relax").(bool)
	restartMode := d.Get("graceful_restart_mode").(string)
	restartTimer := int64(d.Get("graceful_restart_timer").(int))
	staleTimer := int64(d.Get("graceful_restart_stale_route_timer").(int))
	tags := getPolicyTagsFromSchema(d)

	var aggregationStructs []model.RouteAggregationEntry
	routeAggregations := d.Get("route_aggregation").([]interface{})
	if len(routeAggregations) > 0 {
		for _, agg := range routeAggregations {
			data := agg.(map[string]interface{})
			prefix := data["prefix"].(string)
			summary := data["summary_only"].(bool)
			elem := model.RouteAggregationEntry{
				Prefix:      &prefix,
				SummaryOnly: &summary,
			}

			aggregationStructs = append(aggregationStructs, elem)
		}
	}

	restartTimerStruct := model.BgpGracefulRestartTimer{
		RestartTimer:    &restartTimer,
		StaleRouteTimer: &staleTimer,
	}

	restartConfigStruct := model.BgpGracefulRestartConfig{
		Mode:  &restartMode,
		Timer: &restartTimerStruct,
	}

	localeServicePath := d.Get("locale_service_path").(string)
	t0ID, serviceID := resourceNsxtPolicyBgpConfigParseIDs(localeServicePath)
	if t0ID == "" || serviceID == "" {
		return fmt.Errorf("Invalid locale service path %s", localeServicePath)
	}

	routeStruct := model.BgpRoutingConfig{
		Ecmp:                  &ecmp,
		Enabled:               &enabled,
		RouteAggregations:     aggregationStructs,
		Tags:                  tags,
		InterSrIbgp:           &interSrIbgp,
		LocalAsNum:            &localAsNum,
		MultipathRelax:        &multipathRelax,
		GracefulRestartConfig: &restartConfigStruct,
	}

	var err error
	if isPolicyGlobalManager(m) {
		gmObj, err1 := convertModelBindingType(routeStruct, model.GroupBindingType(), gm_model.GroupBindingType())
		if err1 != nil {
			return err1
		}
		gmRouteStruct := gmObj.(gm_model.BgpRoutingConfig)
		client := gm_locale_services.NewDefaultBgpClient(connector)
		err = client.Patch(t0ID, serviceID, gmRouteStruct)
	} else {
		client := locale_services.NewDefaultBgpClient(connector)
		err = client.Patch(t0ID, serviceID, routeStruct)

	}
	if err != nil {
		return handleUpdateError("BgpRoutingConfig", serviceID, err)
	}

	return resourceNsxtPolicyBgpConfigRead(d, m)
}

func resourceNsxtPolicyBgpConfigDelete(d *schema.ResourceData, m interface{}) error {
	// TODO: revert to default
	return nil
}
