/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDhcpV4StaticBindingHelperName = getAccTestResourceName()

var accTestPolicyDhcpV4StaticBindingCreateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform created",
	"gateway_address": "10.2.2.1",
	"hostname":        "test-create",
	"ip_address":      "10.2.2.167",
	"lease_time":      "162",
	"mac_address":     "10:0e:00:11:22:02",
}

var accTestPolicyDhcpV4StaticBindingUpdateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform updated",
	"gateway_address": "10.2.2.2",
	"hostname":        "test-update",
	"ip_address":      "10.2.2.169",
	"lease_time":      "500",
	"mac_address":     "10:ff:22:11:cc:02",
}

var testAccPolicyDhcpV4StaticBindingResourceName = "nsxt_policy_dhcp_v4_static_binding.test"

func testAccOnlyLocalManagerForFixedSegments(t *testing.T, isFixed bool) {
	if !isFixed {
		return
	}

	testAccOnlyLocalManager(t)
}

func TestAccResourceNsxtPolicyDhcpV4StaticBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV4StaticBindingBasic(t, false)
}

func TestAccResourceNsxtPolicyDhcpV4StaticBinding_fixedSegment(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV4StaticBindingBasic(t, true)
}

func testAccResourceNsxtPolicyDhcpV4StaticBindingBasic(t *testing.T, isFixed bool) {
	testResourceName := testAccPolicyDhcpV4StaticBindingResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNSXVersion(t, "3.0.0")
			testAccOnlyLocalManagerForFixedSegments(t, isFixed)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV4StaticBindingCheckDestroy(state, accTestPolicyDhcpV4StaticBindingCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingTemplate(isFixed, true),
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
				Config: testAccNsxtPolicyDhcpV4StaticBindingTemplate(isFixed, false),
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
				Config: testAccNsxtPolicyDhcpV4StaticBindingMinimalistic(isFixed),
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
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV4StaticBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV4StaticBindingMinimalistic(false),
			},
			{
				ResourceName:      testAccPolicyDhcpV4StaticBindingResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyDhcpV4StaticBindingImporterGetID,
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
		if resourceID == "" {
			return fmt.Errorf("Policy DhcpV4StaticBinding resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(resourceID, segmentPath, connector, testAccIsGlobalManager())
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
		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(resourceID, segmentPath, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DhcpV4StaticBinding %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyDhcpV4StaticBindingImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccPolicyDhcpV4StaticBindingResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding resource %s not found in resources", testAccPolicyDhcpV4StaticBindingResourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding resource ID not set in resources ")
	}
	segmentPath := rs.Primary.Attributes["segment_path"]
	if segmentPath == "" {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding Segment Path not set in resources ")
	}
	segs := strings.Split(segmentPath, "/")
	return fmt.Sprintf("%s/%s", segs[len(segs)-1], resourceID), nil
}

func testAccNsxtPolicyDhcpStaticBindingPrerequisites(isFixed bool, isIpv6 bool) string {
	helperName := accTestPolicyDhcpV4StaticBindingHelperName
	segmentResource := "nsxt_policy_segment"
	if isFixed {
		segmentResource = "nsxt_policy_fixed_segment"
	}
	address := "10.2.2.3/24"
	cidr := "10.2.2.1/24"
	version := "v4"
	if isIpv6 {
		address = "1001::3/24"
		cidr = "1001::1/64"
		version = "v6"
	}
	return testAccNsxtPolicyGatewayFabricDeps(false) + fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
  display_name = "%s"
}

resource "nsxt_policy_dhcp_server" "test" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
}

resource "%s" "test" {
  display_name        = "%s"
  connectivity_path   = nsxt_policy_tier1_gateway.test.path
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  dhcp_config_path    = nsxt_policy_dhcp_server.test.path
  subnet {
    cidr = "%s"
    dhcp_%s_config {
        server_address = "%s"
    }
  }
}`, helperName, helperName, segmentResource, helperName, cidr, version, address)
}

func testAccNsxtPolicyGetSegmentResourceName(isFixed bool) string {
	if isFixed {
		return "nsxt_policy_fixed_segment"
	}

	return "nsxt_policy_segment"
}

func testAccNsxtPolicyDhcpV4StaticBindingTemplate(isFixed bool, createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDhcpV4StaticBindingCreateAttributes
	} else {
		attrMap = accTestPolicyDhcpV4StaticBindingUpdateAttributes
	}
	return testAccNsxtPolicyDhcpStaticBindingPrerequisites(isFixed, false) + fmt.Sprintf(`

resource "nsxt_policy_dhcp_v4_static_binding" "test" {
  segment_path    = %s.test.path
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
`, testAccNsxtPolicyGetSegmentResourceName(isFixed), attrMap["display_name"], attrMap["description"], attrMap["gateway_address"], attrMap["hostname"], attrMap["ip_address"], attrMap["lease_time"], attrMap["mac_address"])
}

func testAccNsxtPolicyDhcpV4StaticBindingMinimalistic(isFixed bool) string {
	attrMap := accTestPolicyDhcpV4StaticBindingUpdateAttributes
	return testAccNsxtPolicyDhcpStaticBindingPrerequisites(isFixed, false) + fmt.Sprintf(`
resource "nsxt_policy_dhcp_v4_static_binding" "test" {
  segment_path = %s.test.path
  display_name = "%s"
  ip_address   = "%s"
  mac_address  = "%s"
}`, testAccNsxtPolicyGetSegmentResourceName(isFixed), attrMap["display_name"], attrMap["ip_address"], attrMap["mac_address"])
}
