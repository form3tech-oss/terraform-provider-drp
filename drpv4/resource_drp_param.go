package drpv4

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceParam is the Terraform resource for a param
//
//	resource "drp_param" "test" {
//	  name = "test"
//	  description = "test"
//	  documentation = "test"
//	  schema = {
//		 type = "string"
//	  }
//	  secure = false
//	}
func resourceParam() *schema.Resource {
	r := &schema.Resource{
		Create: resourceParamCreate,
		Read:   resourceParamRead,
		Update: resourceParamUpdate,
		Delete: resourceParamDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Param name",
				Required:    true,
				ForceNew:    true,
				Optional:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Param description",
				ForceNew:    false,
				Optional:    true,
			},
			"documentation": {
				Type:        schema.TypeString,
				Description: "Param documentation",
				ForceNew:    false,
				Optional:    true,
			},
			"schema": {
				Type:        schema.TypeMap,
				Description: "Param schema",
				Default:     `{"type":"string"}`,
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secure": {
				Type:        schema.TypeBool,
				Description: "Param secure",
				ForceNew:    false,
				Optional:    true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}

	return r
}

// flattenParam converts a param object to a Terraform resource
func flattenParam(d *schema.ResourceData, param *models.Param) error {
	if err := d.Set("name", param.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}
	if err := d.Set("description", param.Description); err != nil {
		return fmt.Errorf("error setting description: %s", err)
	}
	if err := d.Set("documentation", param.Documentation); err != nil {
		return fmt.Errorf("error setting documentation: %s", err)
	}
	if err := d.Set("schema", param.Schema); err != nil {
		return fmt.Errorf("error setting schema: %s", err)
	}
	if err := d.Set("secure", param.Secure); err != nil {
		return fmt.Errorf("error setting secure: %s", err)
	}

	return nil
}

// expandParam converts a Terraform resource to a param object
func expandParam(d *schema.ResourceData) *models.Param {
	param := &models.Param{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Documentation: d.Get("documentation").(string),
		Schema:        d.Get("schema").(map[string]interface{}),
		Secure:        d.Get("secure").(bool),
	}

	return param
}

func resourceParamCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Creating param: %+v", d)

	param := expandParam(d)

	log.Printf("Creating param: %+v", param)

	if err := c.session.CreateModel(param); err != nil {
		return fmt.Errorf("error creating param: %s", err)
	}

	d.SetId(param.Key())

	return resourceParamRead(d, m)
}

func resourceParamRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Reading param: %+v", d)

	po, err := c.session.GetModel("params", d.Id())
	if err != nil {
		return fmt.Errorf("error reading param: %s", err)
	}

	return flattenParam(d, po.(*models.Param))
}

func resourceParamUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Updating param: %+v", d)

	param := expandParam(d)

	if err := c.session.PutModel(param); err != nil {
		return fmt.Errorf("error updating param: %s", err)
	}

	return resourceParamRead(d, m)
}

func resourceParamDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Deleting param: %s", d.Id())

	if _, err := c.session.DeleteModel("params", d.Id()); err != nil {
		return fmt.Errorf("error deleting param: %s", err)
	}

	d.SetId("")

	return nil
}
