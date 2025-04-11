package drpv4

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

func resourceProfile() *schema.Resource {
	r := &schema.Resource{
		Create: resourceProfileCreate,
		Read:   resourceProfileRead,
		Update: resourceProfileUpdate,
		Delete: resourceProfileDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Profile name",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Profile description",
				Optional:    true,
				Sensitive:   false,
			},
		},
	}

	return r
}

func flattenProfile(d *schema.ResourceData, profile *models.Profile) error {
	if err := d.Set("name", profile.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err := d.Set("description", profile.Description); err != nil {
		return fmt.Errorf("error setting description: %s", err)
	}

	return nil
}

func expandProfile(d *schema.ResourceData) *models.Profile {
	profile := &models.Profile{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	return profile
}

func resourceProfileCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	profile := expandProfile(d)

	log.Printf("Creating profile: %s", d.Get("name").(string))

	if err := c.session.CreateModel(profile); err != nil {
		return fmt.Errorf("error creating profile: %s", err)
	}

	d.SetId(profile.Key())

	return resourceProfileRead(d, m)
}

func resourceProfileRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Reading profile: %s", d.Get("name").(string))

	pr, err := c.session.GetModel("profiles", d.Id())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			d.SetId("")
			return flattenProfile(d, &models.Profile{})
		} else {
			return fmt.Errorf("error reading profile: %s", err)
		}
	}

	return flattenProfile(d, pr.(*models.Profile))
}

func resourceProfileUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Updating profile: %s", d.Get("name").(string))

	profile := expandProfile(d)

	if err := c.session.PutModel(profile); err != nil {
		return fmt.Errorf("error updating profile: %s", err)
	}

	return resourceProfileRead(d, m)
}

func resourceProfileDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("Deleting profile: %s", d.Id())

	if _, err := c.session.DeleteModel("profiles", d.Id()); err != nil {
		return fmt.Errorf("error deleting profile: %s", err)
	}

	d.SetId("")

	return nil
}
