package drpv4

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

type TemplateResult struct {
	Template *models.Template

	Available bool
	Errors    []string
}

func resourceTemplate() *schema.Resource {
	r := &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Update: resourceTemplateUpdate,
		Delete: resourceTemplateDelete,

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:        schema.TypeString,
				Description: "Template id",
				Required:    true,
				ForceNew:    true,
				Optional:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Template description",
				ForceNew:    false,
				Optional:    true,
			},
			"contents": {
				Type:        schema.TypeString,
				Description: "Template contents",
				ForceNew:    false,
				Optional:    true,
			},
			"start_delimiter": {
				Type:        schema.TypeString,
				Description: "Template start delimiter",
				ForceNew:    false,
				Optional:    true,
			},
			"end_delimiter": {
				Type:        schema.TypeString,
				Description: "Template end delimiter",
				ForceNew:    false,
				Optional:    true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}

	return r
}

func resourceTemplateCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	template := models.Template{
		ID:             d.Get("template_id").(string),
		Description:    d.Get("description").(string),
		Contents:       d.Get("contents").(string),
		StartDelimiter: d.Get("start_delimiter").(string),
		EndDelimiter:   d.Get("end_delimiter").(string),
	}

	template.Validate()
	if template.Error() != "" {
		return fmt.Errorf("template validation failed: %v", template.Error())
	}

	d.SetId(d.Get("template_id").(string))

	log.Printf("Creating template %s", d.Id())

	req := c.session.Req().Post(template).UrlFor("templates")
	if err := req.Do(&template); err != nil {
		return fmt.Errorf("error creating template: %s", err)
	}

	if template.Error() != "" {
		return fmt.Errorf("error creating template: %v", template.Error())
	}

	return nil
}

func resourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	to, err := c.session.GetModel("templates", d.Id())
	if err != nil {
		return fmt.Errorf("error reading template: %s", err)
	}

	templateObject := to.(*models.Template)

	if err := d.Set("template_id", templateObject.ID); err != nil {
		return fmt.Errorf("error setting template ID: %s", err)
	}

	if err := d.Set("description", templateObject.Description); err != nil {
		return fmt.Errorf("error setting template description: %s", err)
	}

	if err := d.Set("contents", templateObject.Contents); err != nil {
		return fmt.Errorf("error setting template contents: %s", err)
	}

	if err := d.Set("start_delimiter", templateObject.StartDelimiter); err != nil {
		return fmt.Errorf("error setting template start delimiter: %s", err)
	}

	if err := d.Set("end_delimiter", templateObject.EndDelimiter); err != nil {
		return fmt.Errorf("error setting template end delimiter: %s", err)
	}

	return nil
}

func resourceTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	var templateResult TemplateResult
	template := models.Template{
		ID:             d.Get("template_id").(string),
		Description:    d.Get("description").(string),
		Contents:       d.Get("contents").(string),
		StartDelimiter: d.Get("start_delimiter").(string),
		EndDelimiter:   d.Get("end_delimiter").(string),
	}

	template.Validate()
	if template.Error() != "" {
		return fmt.Errorf("template validation failed: %v", template.Error())
	}

	req := c.session.Req().Put(template).UrlFor("templates", d.Id())
	if err := req.Do(&templateResult); err != nil {
		return fmt.Errorf("error updating template: %s", err)
	}

	return nil
}

func resourceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	_, err := c.session.DeleteModel("templates", d.Id())
	if err != nil {
		return fmt.Errorf("error deleting template: %s", err)
	}
	d.SetId("")

	return nil
}
