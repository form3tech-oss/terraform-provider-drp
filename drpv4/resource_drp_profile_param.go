package drpv4

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceProfileParam represents a profile param in DRP system
//
//		resource "drp_profile_param" "profile_param" {
//		 profile = "test"
//		 name = "test"
//		 value = "test"
//	  	 secure_value = "test"
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
			"value": {
				Type:         schema.TypeString,
				Description:  "Param value",
				Optional:     true,
				Sensitive:    false,
				ExactlyOneOf: []string{"value", "secure_value"},
			},
			"secure_value": {
				Type:        schema.TypeString,
				Description: "Param secure value",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}

	return r
}

// getParam return the param
func getParam(c *Config, name string) (*models.Param, error) {
	var p *models.Param

	if err := c.session.Req().UrlFor("params", name).Do(&p); err != nil {
		return nil, err
	}

	return p, nil
}

// getParamType returns the type of the parameter
func getParamSchemaType(c *Config, name string) string {
	param, err := getParam(c, name)
	if err != nil {
		return ""
	}

	if param.Schema == nil || param.Schema.(map[string]interface{})["type"] == nil {
		return ""
	}

	s := param.Schema.(map[string]interface{})["type"].(string)

	return s
}

// convertParamToType returns the value in the correct type
func convertParamToType(c *Config, name string, value string) (interface{}, error) {
	paramType := getParamSchemaType(c, name)

	switch paramType {
	case "string":
		return value, nil
	default:
		var out interface{}
		err := json.Unmarshal([]byte(value), &out)
		if err != nil {
			return value, nil
		}
		return out, nil
	}
}

// convertParamToString returns the value in the correct type
func convertParamToString(value interface{}) (string, error) {
	switch value := value.(type) {
	case string:
		return value, nil
	default:
		out, err := json.Marshal(value)
		return string(out), err
	}
}

// isParamSecure returns true if the param is secure
func isParamSecure(c *Config, name string) bool {
	res, err := c.session.GetModel("params", name)
	if err != nil {
		return false
	}

	param := res.(*models.Param)

	return param.Secure
}

// getPublickey returns the public key from DRP to use for encrypting secure params
func getPublickey(c *Config, profile string) ([]byte, error) {
	var pubkey []byte
	if err := c.session.Req().UrlFor("profiles", profile, "pubkey").Do(&pubkey); err != nil {
		return nil, err
	}

	return pubkey, nil
}

// resourceProfileParamCreate creates a profile param in the DRP system
func resourceProfileParamCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	profile := d.Get("profile").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)
	secureValue := d.Get("secure_value").(string)

	log.Printf("Creating profile param %s", name)

	if value != "" && isParamSecure(c, name) {
		return fmt.Errorf("param %s is secure, use secure_value instead", name)
	}

	req := c.session.Req().UrlFor("profiles", profile, "params", name)

	if secureValue != "" {
		sv := models.SecureData{}
		pubkey, err := getPublickey(c, profile)
		if err != nil {
			return err
		}

		err = sv.Marshal(pubkey, value)
		if err != nil {
			return err
		}

		if err := req.Post(sv).Do(nil); err != nil {
			return err
		}
	} else {
		convertedValue, err := convertParamToType(c, name, value)
		if err != nil {
			return err
		}

		if err := req.Post(convertedValue).Do(nil); err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", profile, name))

	return resourceProfileParamRead(d, m)
}

// resourceProfileParamRead reads a profile param from the DRP system
func resourceProfileParamRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	profile := d.Get("profile").(string)
	name := d.Get("name").(string)
	secureValue := d.Get("secure_value").(string)

	log.Printf("Reading profile param %s", name)

	var p interface{}
	if err := c.session.Req().UrlFor("profiles", profile, "params", name).Do(&p); err != nil {
		return err
	}

	d.Set("name", name)
	d.Set("profile", profile)

	if secureValue == "" {
		value, err := convertParamToString(p)
		if err != nil {
			return err
		}

		d.Set("value", value)
	}

	return nil
}

// resourceProfileParamUpdate updates a profile param in the DRP system
func resourceProfileParamUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceProfileParamCreate(d, m)
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
