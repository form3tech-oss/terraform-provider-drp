package drpv4

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceProfileParam represents a profile param in DRP system
//
//		resource "drp_profile_param" "profile_param" {
//	   profile = "test"
//		  name = "test"
//		  schema = {
//		    type = "string"
//		  }
//		}
func resourceProfileParam() *schema.Resource {
	r := &schema.Resource{
		Create: resourceProfileParamCreate,
		Read:   resourceProfileParamRead,
		Update: resourceProfileParamUpdate,
		Delete: resourceProfileParamDelete,

		Schema: map[string]*schema.Schema{
			"profile": {
				Type:        schema.TypeString,
				Description: "Profile name",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Param name",
				Required:    true,
				ForceNew:    true,
			},
			"schema": {
				Type:        schema.TypeMap,
				Description: "Param schema",
				Default:     `{"type":"string"}`,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

	return r
}

// resourceProfileParamCreate creates a profile param in the DRP system
func resourceProfileParamCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	profile := d.Get("profile").(string)
	name := d.Get("name").(string)
	paramSchema := d.Get("schema").(map[string]interface{})

	log.Printf("Creating profile param %s", name)

	err := c.session.Req().Post(paramSchema).UrlFor("profiles", profile, "params", name).Do(nil)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", profile, name))

	return resourceProfileParamRead(d, m)
}

// resourceProfileParamRead reads a profile param from the DRP system
func resourceProfileParamRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	profile := d.Get("profile").(string)
	name := d.Get("name").(string)

	log.Printf("Reading profile param %s", name)

	var p map[string]interface{}
	if err := c.session.Req().UrlFor("profiles", profile, "params", name).Do(&p); err != nil {
		return err
	}

	d.Set("name", name)
	d.Set("profile", profile)
	d.Set("schema", p)

	return nil
}

// resourceProfileParamUpdate updates a profile param in the DRP system
func resourceProfileParamUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	name := d.Get("name").(string)
	profile := d.Get("profile").(string)
	paramSchema := d.Get("schema").(map[string]interface{})

	log.Printf("Updating profile param %s", name)

	if err := c.session.Req().Post(paramSchema).UrlFor("profiles", profile, "params", name).Do(nil); err != nil {
		return err
	}

	return resourceProfileParamRead(d, m)
}

// resourceProfileParamDelete deletes a profile param from the DRP system
func resourceProfileParamDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	name := d.Get("name").(string)
	profile := d.Get("profile").(string)

	log.Printf("Deleting profile param %s", name)

	if err := c.session.Req().Del().UrlFor("profiles", profile, "params", name).Do(nil); err != nil {
		return err
	}

	return nil
}
