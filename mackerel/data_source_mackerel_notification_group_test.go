package mackerel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceMackerelNotificationGroup(t *testing.T) {
	dsName := "data.mackerel_notification_group.foo"
	rand := acctest.RandString(5)
	name := fmt.Sprintf("tf-notification-group-%s", rand)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMackerelNotificationGroupConfig(rand, name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("notification_level"), knownvalue.StringExact("critical")),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("child_notification_group_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("child_channel_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("monitor"),
						knownvalue.SetExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":           knownvalue.NotNull(),
							"skip_default": knownvalue.Bool(false),
						})})),
					statecheck.ExpectKnownValue(dsName, tfjsonpath.New("service"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{"name": knownvalue.StringExact("tf-service-" + rand)}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{"name": knownvalue.StringExact("tf-service-" + rand + "-bar")}),
						})),
				},
			},
		},
	})
}

func testAccDataSourceMackerelNotificationGroupConfig(rand, name string) string {
	return fmt.Sprintf(`
resource "mackerel_service" "foo" {
  name = "tf-service-%s"
}

resource "mackerel_service" "bar" {
  name = "tf-service-%s-bar"
}

resource "mackerel_channel" "foo" {
  name = "tf-channel-%s"
  email {}
}

resource "mackerel_monitor" "foo" {
  name = "tf-monitor-%s"
  connectivity {}
}

resource "mackerel_notification_group" "child" {
  name = "tf-notification-group-%s-child"
}

resource "mackerel_notification_group" "foo" {
  name = "%s"
  notification_level = "critical"
  child_notification_group_ids = [
    mackerel_notification_group.child.id]
  child_channel_ids = [
    mackerel_channel.foo.id]
  monitor {
    id = mackerel_monitor.foo.id
    skip_default = false
  }
  // ignore duplicates
  monitor {
    id = mackerel_monitor.foo.id
    skip_default = false
  }
  service {
    name = mackerel_service.foo.name
  }
  // ignore duplicates
  service {
    name = mackerel_service.foo.name
  }
  service {
    name = mackerel_service.bar.name
  }
}

data "mackerel_notification_group" "foo" {
  id = mackerel_notification_group.foo.id
}
`, rand, rand, rand, rand, rand, name)
}
