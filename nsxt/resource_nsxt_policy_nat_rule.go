/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_t0nat "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/nat"
	gm_t1nat "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_1s/nat"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	t0nat "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/nat"
	t1nat "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/nat"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var policyNATRuleActionTypeValues = []string{
	model.PolicyNatRule_ACTION_SNAT,
	model.PolicyNatRule_ACTION_DNAT,
	model.PolicyNatRule_ACTION_REFLEXIVE,
	model.PolicyNatRule_ACTION_NO_SNAT,
	model.PolicyNatRule_ACTION_NO_DNAT,
	model.PolicyNatRule_ACTION_NAT64,
}
var policyNATRuleFirewallMatchTypeValues = []string{
	model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS,
	model.PolicyNatRule_FIREWALL_MATCH_MATCH_INTERNAL_ADDRESS,
	model.PolicyNatRule_FIREWALL_MATCH_BYPASS,
}

func resourceNsxtPolicyNATRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyNATRuleCreate,
		Read:   resourceNsxtPolicyNATRuleRead,
		Update: resourceNsxtPolicyNATRuleUpdate,
		Delete: resourceNsxtPolicyNATRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyNATRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"gateway_path": getPolicyGatewayPathSchema(),
			"action": {
				Type:         schema.TypeString,
				Description:  "The action for the NAT Rule",
				Required:     true,
				ValidateFunc: validation.StringInSlice(policyNATRuleActionTypeValues, false),
			},
			"destination_networks": {
				Type:        schema.TypeList,
				Description: "The destination network(s) for the NAT Rule",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrOrIPOrRange(),
				},
			},
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Description: "Enable/disable the rule",
				Optional:    true,
			},
			"firewall_match": {
				Type:         schema.TypeString,
				Description:  "Firewall match flag",
				Optional:     true,
				Default:      model.PolicyNatRule_FIREWALL_MATCH_BYPASS,
				ValidateFunc: validation.StringInSlice(policyNATRuleFirewallMatchTypeValues, false),
			},
			"logging": {
				Type:        schema.TypeBool,
				Default:     false,
				Description: "Enable/disable the logging of rule",
				Optional:    true,
			},
			"rule_priority": {
				// called 'sequence_number' in VAPI
				Type:        schema.TypeInt,
				Default:     100,
				Description: "The sequence_number decides the rule_priority of a NAT rule. Valid range [0-2147483647]",
				Optional:    true,
			},
			"service": {
				Type:         schema.TypeString,
				Description:  "Policy path of Service on which the NAT rule will be applied",
				Optional:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"source_networks": {
				Type:        schema.TypeList,
				Description: "The source network(s) for the NAT Rule",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrOrIPOrRange(),
				},
			},
			"translated_networks": {
				Type:        schema.TypeList,
				Description: "The translated network(s) for the NAT Rule",
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrOrIPOrRange(),
				},
			},
			"translated_ports": {
				Type:         schema.TypeString,
				Description:  "Port number or port range. DNAT only",
				Optional:     true,
				ValidateFunc: validatePortRange(),
			},
			"scope": {
				Type:        schema.TypeSet,
				Description: "Policy paths to interfaces or labels where the NAT Rule is enforced",
				Optional:    true,
				Computed:    true,
				Elem:        getElemPolicyPathSchema(),
			},
		},
	}
}

func deleteNsxtPolicyNATRule(connector *client.RestConnector, gwID string, isT0 bool, ruleID string, isGlobalManager bool) error {
	if isGlobalManager {
		if isT0 {
			client := gm_t0nat.NewNatRulesClient(connector)
			return client.Delete(gwID, gm_model.PolicyNat_NAT_TYPE_USER, ruleID)
		}
		client := gm_t1nat.NewNatRulesClient(connector)
		return client.Delete(gwID, gm_model.PolicyNat_NAT_TYPE_USER, ruleID)
	}
	if isT0 {
		client := t0nat.NewNatRulesClient(connector)
		return client.Delete(gwID, model.PolicyNat_NAT_TYPE_USER, ruleID)
	}
	client := t1nat.NewNatRulesClient(connector)
	return client.Delete(gwID, model.PolicyNat_NAT_TYPE_USER, ruleID)
}

func resourceNsxtPolicyNATRuleDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining NAT Rule ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	err := deleteNsxtPolicyNATRule(getPolicyConnector(m), gwID, isT0, id, isPolicyGlobalManager(m))
	if err != nil {
		return handleDeleteError("NAT Rule", id, err)
	}

	return nil
}

func getNsxtPolicyNATRuleByID(connector *client.RestConnector, gwID string, isT0 bool, ruleID string, isGlobalManager bool) (model.PolicyNatRule, error) {
	if isGlobalManager {
		var obj model.PolicyNatRule
		var gmObj gm_model.PolicyNatRule
		var rawObj interface{}
		var err error
		if isT0 {
			client := gm_t0nat.NewNatRulesClient(connector)
			gmObj, err = client.Get(gwID, gm_model.PolicyNat_NAT_TYPE_USER, ruleID)
		} else {
			client := gm_t1nat.NewNatRulesClient(connector)
			gmObj, err = client.Get(gwID, gm_model.PolicyNat_NAT_TYPE_USER, ruleID)
		}
		if err != nil {
			return obj, err
		}
		rawObj, err = convertModelBindingType(gmObj, gm_model.PolicyNatRuleBindingType(), model.PolicyNatRuleBindingType())
		if err != nil {
			return obj, err
		}
		return rawObj.(model.PolicyNatRule), err
	}
	if isT0 {
		client := t0nat.NewNatRulesClient(connector)
		return client.Get(gwID, model.PolicyNat_NAT_TYPE_USER, ruleID)
	}
	client := t1nat.NewNatRulesClient(connector)
	return client.Get(gwID, model.PolicyNat_NAT_TYPE_USER, ruleID)
}

func patchNsxtPolicyNATRule(connector *client.RestConnector, gwID string, rule model.PolicyNatRule, isT0 bool, isGlobalManager bool) error {
	if isGlobalManager {
		rawObj, err := convertModelBindingType(rule, model.PolicyNatRuleBindingType(), gm_model.PolicyNatRuleBindingType())
		if err != nil {
			return err
		}
		if isT0 {
			client := gm_t0nat.NewNatRulesClient(connector)
			return client.Patch(gwID, model.PolicyNat_NAT_TYPE_USER, *rule.Id, rawObj.(gm_model.PolicyNatRule))
		}
		client := gm_t1nat.NewNatRulesClient(connector)
		return client.Patch(gwID, model.PolicyNat_NAT_TYPE_USER, *rule.Id, rawObj.(gm_model.PolicyNatRule))
	}
	if isT0 {
		client := t0nat.NewNatRulesClient(connector)
		return client.Patch(gwID, model.PolicyNat_NAT_TYPE_USER, *rule.Id, rule)
	}
	client := t1nat.NewNatRulesClient(connector)
	return client.Patch(gwID, model.PolicyNat_NAT_TYPE_USER, *rule.Id, rule)
}

func resourceNsxtPolicyNATRuleRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining NAT Rule ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	obj, err := getNsxtPolicyNATRuleByID(connector, gwID, isT0, id, isPolicyGlobalManager(m))
	if err != nil {
		return handleReadError(d, "NAT Rule", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("action", obj.Action)
	if obj.DestinationNetwork != nil {
		d.Set("destination_networks", commaSeparatedStringToStringList(*obj.DestinationNetwork))
	}
	d.Set("enabled", obj.Enabled)
	d.Set("firewall_match", obj.FirewallMatch)
	d.Set("logging", obj.Logging)
	d.Set("rule_priority", obj.SequenceNumber)
	d.Set("service", obj.Service)
	if obj.SourceNetwork != nil {
		d.Set("source_networks", commaSeparatedStringToStringList(*obj.SourceNetwork))
	}
	if obj.TranslatedNetwork != nil {
		d.Set("translated_networks", commaSeparatedStringToStringList(*obj.TranslatedNetwork))
	}
	d.Set("translated_ports", obj.TranslatedPorts)
	d.Set("scope", obj.Scope)

	d.SetId(id)

	return nil
}

func resourceNsxtPolicyNATRuleCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	isGlobalManager := isPolicyGlobalManager(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := getNsxtPolicyNATRuleByID(connector, gwID, isT0, id, isGlobalManager)
		if err == nil {
			return fmt.Errorf("NAT Rule with nsx_id '%s' already exists", id)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	action := d.Get("action").(string)
	enabled := d.Get("enabled").(bool)
	fwMatch := d.Get("firewall_match").(string)
	logging := d.Get("logging").(bool)
	priority := int64(d.Get("rule_priority").(int))
	service := d.Get("service").(string)
	ports := d.Get("translated_ports").(string)
	dNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("destination_networks").([]interface{})))
	sNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("source_networks").([]interface{})))
	tNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("translated_networks").([]interface{})))
	scope := getStringListFromSchemaSet(d, "scope")
	tags := getPolicyTagsFromSchema(d)

	ruleStruct := model.PolicyNatRule{
		Id:                 &id,
		DisplayName:        &displayName,
		Description:        &description,
		Tags:               tags,
		Action:             &action,
		DestinationNetwork: &dNets,
		Enabled:            &enabled,
		Logging:            &logging,
		SequenceNumber:     &priority,
		Service:            &service,
		TranslatedNetwork:  &tNets,
		Scope:              scope,
	}

	// handle values that can't be an empty string
	if fwMatch != "" {
		ruleStruct.FirewallMatch = &fwMatch
	}
	if ports != "" {
		ruleStruct.TranslatedPorts = &ports
	}

	if len(sNets) > 0 {
		ruleStruct.SourceNetwork = &sNets
	}

	log.Printf("[INFO] Creating NAT Rule with ID %s", id)

	err := patchNsxtPolicyNATRule(connector, gwID, ruleStruct, isT0, isGlobalManager)
	if err != nil {
		return handleCreateError("NAT Rule", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyNATRuleRead(d, m)
}

func resourceNsxtPolicyNATRuleUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining NAT Rule ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	action := d.Get("action").(string)
	enabled := d.Get("enabled").(bool)
	logging := d.Get("logging").(bool)
	priority := int64(d.Get("rule_priority").(int))
	service := d.Get("service").(string)
	dNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("destination_networks").([]interface{})))
	sNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("source_networks").([]interface{})))
	tNets := stringListToCommaSeparatedString(interfaceListToStringList(d.Get("translated_networks").([]interface{})))
	tags := getPolicyTagsFromSchema(d)
	scope := getStringListFromSchemaSet(d, "scope")

	ruleStruct := model.PolicyNatRule{
		Id:                 &id,
		DisplayName:        &displayName,
		Description:        &description,
		Tags:               tags,
		Action:             &action,
		DestinationNetwork: &dNets,
		Enabled:            &enabled,
		Logging:            &logging,
		SequenceNumber:     &priority,
		Service:            &service,
		TranslatedNetwork:  &tNets,
		Scope:              scope,
	}

	// handle values that can't be an empty string
	fwMatch := d.Get("firewall_match").(string)
	if fwMatch != "" {
		ruleStruct.FirewallMatch = &fwMatch
	}
	tPorts := d.Get("translated_ports").(string)
	if tPorts != "" {
		ruleStruct.TranslatedPorts = &tPorts
	}
	if len(sNets) > 0 {
		ruleStruct.SourceNetwork = &sNets
	}

	log.Printf("[INFO] Updating NAT Rule with ID %s", id)
	err := patchNsxtPolicyNATRule(connector, gwID, ruleStruct, isT0, isPolicyGlobalManager(m))
	if err != nil {
		return handleUpdateError("NAT Rule", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyNATRuleRead(d, m)
}

func resourceNsxtPolicyNATRuleImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <gateway-id>/<nat-rule-id> as an input")
	}

	gwID := s[0]
	connector := getPolicyConnector(m)
	if isPolicyGlobalManager(m) {
		t0Client := gm_infra.NewTier0sClient(connector)
		t0gw, err := t0Client.Get(gwID)
		if err != nil {
			if !isNotFoundError(err) {
				return nil, err
			}
			t1Client := gm_infra.NewTier1sClient(connector)
			t1gw, err := t1Client.Get(gwID)
			if err != nil {
				return nil, err
			}
			d.Set("gateway_path", t1gw.Path)
		} else {
			d.Set("gateway_path", t0gw.Path)
		}
	} else {
		t0Client := infra.NewTier0sClient(connector)
		t0gw, err := t0Client.Get(gwID)
		if err != nil {
			if !isNotFoundError(err) {
				return nil, err
			}
			t1Client := infra.NewTier1sClient(connector)
			t1gw, err := t1Client.Get(gwID)
			if err != nil {
				return nil, err
			}
			d.Set("gateway_path", t1gw.Path)
		} else {
			d.Set("gateway_path", t0gw.Path)
		}
	}
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil

}
