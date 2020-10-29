/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDhcpV4StaticBindingCreateAttributes = map[string]string{
	"display_name":    "terra-test",
	"description":     "terraform created",
	"gateway_address": "10.2.2.1",
	"hostname":        "test-create",
	"ip_address":      "10.2.2.167",
	"lease_time":      "162",
	"mac_address":     "10:0e:00:11:22:02",
}

var accTestPolicyDhcpV4StaticBindingUpdateAttributes = map[string]string{
	"display_name":    "terra-test-updated",
	"description":     "terraform updated",
	"gateway_address": "10.20.0.1",
	"hostname":        "test-update",
	"ip_address":      "10.20.0.167",
	"lease_time":      "500",
	"mac_address":     "10:ff:22:11:cc:02",
}

func TestAccResourceNsxtPolicyDhcpV4StaticBinding_basic(t *testing.T) {
	testResourceName := "nsxt_policy_dhcp_v4_static_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV4StaticBindingCheckDestroy(state, accTestPolicyDhcpV4StaticBindingCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV4StaticBindingExists(accTestPolicyDhcpV4StaticBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpV4StaticBindingCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpV4StaticBindingCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_address", accTestPolicyDhcpV4StaticBindingCreateAttributes["gateway_address"]),
					resource.TestCheckResourceAttr(testResourceName, "hostname", accTestPolicyDhcpV4StaticBindingCreateAttributes["hostname"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyDhcpV4StaticBindingCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestPolicyDhcpV4StaticBindingCreateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestPolicyDhcpV4StaticBindingCreateAttributes["mac_address"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV4StaticBindingExists(accTestPolicyDhcpV4StaticBindingUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpV4StaticBindingUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpV4StaticBindingUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_address", accTestPolicyDhcpV4StaticBindingUpdateAttributes["gateway_address"]),
					resource.TestCheckResourceAttr(testResourceName, "hostname", accTestPolicyDhcpV4StaticBindingUpdateAttributes["hostname"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyDhcpV4StaticBindingUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestPolicyDhcpV4StaticBindingUpdateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestPolicyDhcpV4StaticBindingUpdateAttributes["mac_address"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV4StaticBindingExists(accTestPolicyDhcpV4StaticBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDhcpV4StaticBinding_importBasic(t *testing.T) {
	name := "terra-test-import"
	testResourceName := "nsxt_policy_dhcp_v4_static_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV4StaticBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyDhcpV4StaticBindingExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DhcpV4StaticBinding resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		segmentPath := rs.Primary.Attributes["segment_path"]
		segmentID := getPolicyIDFromPath(segmentPath)
		if resourceID == "" {
			return fmt.Errorf("Policy DhcpV4StaticBinding resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(resourceID, segmentID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy DhcpV4StaticBinding %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDhcpV4StaticBindingCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_dhcp_v4_static_binding" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		segmentPath := rs.Primary.Attributes["segment_path"]
		segmentID := getPolicyIDFromPath(segmentPath)
		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(resourceID, segmentID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DhcpV4StaticBinding %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDhcpV4StaticBindingTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDhcpV4StaticBindingCreateAttributes
	} else {
		attrMap = accTestPolicyDhcpV4StaticBindingUpdateAttributes
	}
	return testAccNSXPolicyTransportZoneReadTemplate(getOverlayTransportZoneName(), false, true) + testAccNsxtPolicyEdgeCluster(getEdgeClusterName()) + fmt.Sprintf(`
resource "nsxt_policy_dhcp_server" "test" {
  display_name      = "terraform-test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  server_addresses  = ["110.64.0.1/16"]
}

resource "nsxt_policy_segment" "test" {
  display_name        = "terraform-test"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  dhcp_config_path    = nsxt_policy_dhcp_server.test.path
  subnet {
    cidr = "10.2.2.1/24"
  }
}

resource "nsxt_policy_dhcp_v4_static_binding" "test" {
  segment_path    = nsxt_policy_segment.test.path
  display_name    = "%s"
  description     = "%s"
  gateway_address = "%s"
  hostname        = "%s"
  ip_address      = "%s"
  lease_time      = %s
  mac_address     = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_dhcp_v4_static_binding.test.path
}`, attrMap["display_name"], attrMap["description"], attrMap["gateway_address"], attrMap["hostname"], attrMap["ip_address"], attrMap["lease_time"], attrMap["mac_address"])
}

func testAccNsxtPolicyDhcpV4StaticBindingMinimalistic() string {
	return testAccNSXPolicyTransportZoneReadTemplate(getOverlayTransportZoneName(), false, true) + fmt.Sprintf(`
resource "nsxt_policy_segment" "test" {
  display_name        = "terraform-test"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
}

resource "nsxt_policy_dhcp_v4_static_binding" "test" {
  segment_path    = nsxt_policy_segment.test.path
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_dhcp_v4_static_binding.test.path
}`, accTestPolicyDhcpV4StaticBindingUpdateAttributes["display_name"])
}
