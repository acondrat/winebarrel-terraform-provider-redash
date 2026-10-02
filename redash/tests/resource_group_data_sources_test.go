package test

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	redashgo "github.com/winebarrel/redash-go/v2"
)

func TestAccGroupDataSources_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourcesOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"redash_group_data_sources.my_group", "data_source.*.data_source_id",
						"redash_data_source.my_data_source", "id",
					),
					resource.TestCheckTypeSetElemNestedAttrs("redash_group_data_sources.my_group", "data_source.*", map[string]string{
						"name":      "my-data-source",
						"view_only": "false",
					}),
				),
			},
			{
				Config: testAccGroupDataSourcesTwo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(
						"redash_group_data_sources.my_group", "data_source.*.data_source_id",
						"redash_data_source.my_data_source", "id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"redash_group_data_sources.my_group", "data_source.*.data_source_id",
						"redash_data_source.extra", "id",
					),
				),
			},
			{
				Config: testAccGroupDataSourcesOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"redash_group_data_sources.my_group", "data_source.*.data_source_id",
						"redash_data_source.my_data_source", "id",
					),
				),
			},
			{
				Config: testAccGroupDataSourcesViewOnly,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("redash_group_data_sources.my_group", "data_source.*", map[string]string{
						"view_only": "true",
					}),
				),
			},
			{
				ResourceName:      "redash_group_data_sources.my_group",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupDataSourcesEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "0"),
				),
			},
		},
	})
}

func TestAccGroupDataSources_removesExtra(t *testing.T) {
	var groupId, extraId int

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourcesOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "1"),
					func(s *terraform.State) error {
						var err error
						groupId, err = strconv.Atoi(s.RootModule().Resources["redash_group.my_group"].Primary.ID)
						if err != nil {
							return err
						}
						extraId, err = strconv.Atoi(s.RootModule().Resources["redash_data_source.extra"].Primary.ID)
						return err
					},
				),
			},
			{
				// Add the extra link before this step. Refresh then plans to remove it.
				// Doing this in the previous step's Check makes the post-apply plan non-empty.
				PreConfig: func() {
					client := testAccProvider.Meta().(*redashgo.Client)
					if _, err := client.AddGroupDataSource(context.Background(), groupId, extraId); err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccGroupDataSourcesOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redash_group_data_sources.my_group", "data_source.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"redash_group_data_sources.my_group", "data_source.*.data_source_id",
						"redash_data_source.my_data_source", "id",
					),
				),
			},
		},
	})
}

const testAccGroupDataSourcesBase = testAccGroupConfigBasic + testAccDataSourceConfigBasicPg + `
resource "redash_data_source" "extra" {
	name = "extra-data-source"
	type = "pg"
	options = jsonencode({
		dbname = "postgres"
		host   = "postgres"
		port   = 5432
		user   = "postgres"
	})
}
`

const testAccGroupDataSourcesOne = testAccGroupDataSourcesBase + `
resource "redash_group_data_sources" "my_group" {
	group_id = redash_group.my_group.id

	data_source {
		data_source_id = redash_data_source.my_data_source.id
	}
}
`

const testAccGroupDataSourcesTwo = testAccGroupDataSourcesBase + `
resource "redash_group_data_sources" "my_group" {
	group_id = redash_group.my_group.id

	data_source {
		data_source_id = redash_data_source.my_data_source.id
	}

	data_source {
		data_source_id = redash_data_source.extra.id
	}
}
`

const testAccGroupDataSourcesViewOnly = testAccGroupDataSourcesBase + `
resource "redash_group_data_sources" "my_group" {
	group_id = redash_group.my_group.id

	data_source {
		data_source_id = redash_data_source.my_data_source.id
		view_only      = true
	}
}
`

const testAccGroupDataSourcesEmpty = testAccGroupDataSourcesBase + `
resource "redash_group_data_sources" "my_group" {
	group_id = redash_group.my_group.id
}
`
