package drpv4

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

type ParamResult struct {
	Param *models.Param

	Available bool
	Errors    []string
}

type ParamSchema struct {
	Type string `json:"type"`
}

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

func resourceParamCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	name := d.Get("name").(string)
	if name == "" {
		return fmt.Errorf("param name is required")
	}

	param := models.Param{
		Name:          name,
		Description:   d.Get("description").(string),
		Documentation: d.Get("documentation").(string),
		Schema:        d.Get("schema").(map[string]interface{}),
		Secure:        d.Get("secure").(bool),
	}

	param.Validate()
	if param.Validation.Error() != "" {
		return fmt.Errorf("error validating param: %v", param.Error())
	}

	d.SetId(param.Key())

	var paramResult *ParamResult

	log.Printf("Creating param: %+v", param)

	req := c.session.Req().Post(param).UrlFor("params")
	if err := req.Do(&paramResult); err != nil {
		return fmt.Errorf("error creating param: %s", err)
	}

	if len(paramResult.Errors) > 0 {
		return fmt.Errorf("error creating param: %v", paramResult.Errors)
	}

	return nil
}

func resourceParamRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	po, err := c.session.GetModel("params", d.Id())
	if err != nil {
		return fmt.Errorf("error reading param: %s", err)
	}

	paramObject := po.(*models.Param)

	if err := d.Set("name", paramObject.Name); err != nil {
		return fmt.Errorf("error setting param name: %s", err)
	}

	if err := d.Set("description", paramObject.Description); err != nil {
		return fmt.Errorf("error setting param description: %s", err)
	}

	if err := d.Set("documentation", paramObject.Documentation); err != nil {
		return fmt.Errorf("error setting param documentation: %s", err)
	}

	if err := d.Set("schema", paramObject.Schema.(map[string]interface{})); err != nil {
		return fmt.Errorf("error setting param type: %s", err)
	}

	if err := d.Set("secure", paramObject.Secure); err != nil {
		return fmt.Errorf("error setting param secure: %s", err)
	}

	return nil
}

func resourceParamUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("Updating param: %+v", d)

	c := m.(*Config)

	var paramResult *ParamResult
	param := models.Param{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Documentation: d.Get("documentation").(string),
		Schema:        d.Get("schema").(map[string]interface{}),
		Secure:        d.Get("secure").(bool),
	}

	param.Validate()
	if param.Validation.Error() != "" {
		return fmt.Errorf("error validating param: %v", param.Error())
	}

	req := c.session.Req().Put(param).UrlFor("params", d.Id())
	if err := req.Do(&paramResult); err != nil {
		return fmt.Errorf("error updating param: %s", err)
	}

	if len(paramResult.Errors) > 0 {
		return fmt.Errorf("error updating param: %v", paramResult.Errors)
	}

	return nil
}

func resourceParamDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("Deleting param: %+v", d.State().Attributes)

	c := m.(*Config)

	_, err := c.session.DeleteModel("params", d.Id())
	if err != nil {
		return fmt.Errorf("error deleting param: %s", err)
	}
	d.SetId("")

	return nil
}
