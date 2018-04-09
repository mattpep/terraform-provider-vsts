package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBuildDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceBuildDefinitionCreate,
		Read:   resourceBuildDefinitionRead,
		Update: resourceBuildDefinitionUpdate,
		Delete: resourceBuildDefinitionDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceBuildDefinitionCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	d.SetId(name)
	return nil
}

func resourceBuildDefinitionRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceBuildDefinitionUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("name") {
		// if err := updateName(d, m); err != nil {
		//	return err
		// }
		d.SetPartial("name")
	}

	d.Partial(false)
	return nil
}

func resourceBuildDefinitionDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
