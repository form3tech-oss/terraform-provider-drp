package drpv4

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceTask is the Terraform resource for a task
//
//	resource "drp_task" "test" {
//	  name = "test"
//	  description = "test"
//	  required_params = ["test"]
//	  optional_params = ["test"]
//	  templates {
//		 name = "test"
//		 path = "test"
//		 contents = "test"
//		 link = "test"
//		 meta = {
//			test = "test"
//		 }
//	  }
//	  extra_claims {
//		 scope = "test"
//		 action = "test"
//		 specific = "test"
//	  }
//	  extra_roles = ["test"]
//	  prerequisites = ["test"]
// }

func resourceTask() *schema.Resource {
	r := &schema.Resource{
		Create: resourceTaskCreate,
		Read:   resourceTaskRead,
		Update: resourceTaskUpdate,
		Delete: resourceTaskDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Task name",
				Required:    true,
				ForceNew:    true,
				Optional:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Task description",
				ForceNew:    false,
				Optional:    true,
			},
			"required_params": {
				Type:        schema.TypeList,
				Description: "Task required params",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"optional_params": {
				Type:        schema.TypeList,
				Description: "Task optional params",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"templates": {
				Type:        schema.TypeList,
				Description: "Task templates",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"template_id": {
							Type:        schema.TypeString,
							Description: "Template id",
							ForceNew:    false,
							Optional:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Template name",
							Required:    true,
							ForceNew:    false,
						},
						"path": {
							Type:        schema.TypeString,
							Description: "Template path",
							ForceNew:    false,
							Optional:    true,
						},
						"contents": {
							Type:        schema.TypeString,
							Description: "Template contents",
							ForceNew:    false,
							Optional:    true,
						},
						"link": {
							Type:        schema.TypeString,
							Description: "Template link",
							ForceNew:    false,
							Optional:    true,
						},
						"meta": {
							Type:        schema.TypeMap,
							Description: "Template meta",
							ForceNew:    false,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"extra_claims": {
				Type:        schema.TypeSet,
				Description: "Task extra claims",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope": {
							Type:        schema.TypeString,
							Description: "Claim scope",
							ForceNew:    false,
							Optional:    true,
						},
						"action": {
							Type:        schema.TypeString,
							Description: "Claim action",
							ForceNew:    false,
							Optional:    true,
						},
						"specific": {
							Type:        schema.TypeString,
							Description: "Claim specific",
							ForceNew:    false,
							Optional:    true,
						},
					},
				},
			},
			"extra_roles": {
				Type:        schema.TypeList,
				Description: "Task extra roles",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"prerequisites": {
				Type:        schema.TypeList,
				Description: "Task prerequisites",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
	}

	return r
}

// flattenTemplates flattens the TemplateInfo object into the resource data.
func flattenTemplates(templates []models.TemplateInfo) []map[string]interface{} {
	result := make([]map[string]interface{}, len(templates))
	for i, template := range templates {
		result[i] = map[string]interface{}{
			"template_id": template.ID,
			"name":        template.Name,
			"path":        template.Path,
			"contents":    template.Contents,
			"link":        template.Link,
			"meta":        template.Meta,
		}
	}
	return result
}

// flattenClaims flattens the Claim object into the resource data.
func flattenClaims(claims []*models.Claim) []map[string]interface{} {
	result := make([]map[string]interface{}, len(claims))
	for i, claim := range claims {
		result[i] = map[string]interface{}{
			"scope":    claim.Scope,
			"action":   claim.Action,
			"specific": claim.Specific,
		}
	}
	return result
}

// flattenTask flattens the Task object into the resource data.
func flattenTask(d *schema.ResourceData, task *models.Task) error {
	if err := d.Set("name", task.Name); err != nil {
		return err
	}
	if err := d.Set("description", task.Description); err != nil {
		return err
	}
	if err := d.Set("required_params", task.RequiredParams); err != nil {
		return err
	}
	if err := d.Set("optional_params", task.OptionalParams); err != nil {
		return err
	}
	if err := d.Set("extra_roles", task.ExtraRoles); err != nil {
		return err
	}
	if err := d.Set("prerequisites", task.Prerequisites); err != nil {
		return err
	}
	if err := d.Set("templates", flattenTemplates(task.Templates)); err != nil {
		return err
	}
	if err := d.Set("extra_claims", flattenClaims(task.ExtraClaims)); err != nil {
		return err
	}
	return nil
}

// expandTaskTemplates expands the templates into the Task object.
func expandTaskTemplates(tpls interface{}) []models.TemplateInfo {
	templates := []models.TemplateInfo{}
	for _, rawTemplate := range tpls.([]interface{}) {
		template := rawTemplate.(map[string]interface{})
		templates = append(templates, models.TemplateInfo{
			ID:       template["template_id"].(string),
			Name:     template["name"].(string),
			Path:     template["path"].(string),
			Contents: template["contents"].(string),
			Link:     template["link"].(string),
			Meta:     expandMapInterface(template["meta"]),
		})
	}

	return templates
}

// expandClaims expands the claims into the Task object.
func expandClaims(claims interface{}) []*models.Claim {
	result := []*models.Claim{}
	for _, rawClaim := range claims.(*schema.Set).List() {
		claim := rawClaim.(map[string]interface{})
		result = append(result, &models.Claim{
			Scope:    claim["scope"].(string),
			Action:   claim["action"].(string),
			Specific: claim["specific"].(string),
		})
	}
	return result
}

// expandTask expands the resource data into the Task object.
func expandTask(d *schema.ResourceData) *models.Task {
	task := models.Task{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		RequiredParams: expandStringList(d.Get("required_params")),
		OptionalParams: expandStringList(d.Get("optional_params")),
		ExtraRoles:     expandStringList(d.Get("extra_roles")),
		Prerequisites:  expandStringList(d.Get("prerequisites")),
		Templates:      expandTaskTemplates(d.Get("templates")),
		ExtraClaims:    expandClaims(d.Get("extra_claims")),
	}

	return &task
}

func resourceTaskCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	task := expandTask(d)

	log.Printf("[INFO] Creating task %s", d.Id())

	if err := c.session.CreateModel(task); err != nil {
		return fmt.Errorf("error creating task: %s", err)
	}

	d.SetId(task.Name)

	return nil
}

func resourceTaskRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[INFO] Reading task %s", d.Id())

	to, err := c.session.GetModel("tasks", d.Id())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			d.SetId("")
			return flattenTask(d, &models.Task{})
		} else {
			return fmt.Errorf("error reading task: %s", err)
		}
	}

	return flattenTask(d, to.(*models.Task))
}

func resourceTaskUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	task := expandTask(d)

	log.Printf("[INFO] Updating task %s", d.Id())

	if err := c.session.PutModel(task); err != nil {
		return fmt.Errorf("error updating task: %s", err)
	}

	return nil
}

func resourceTaskDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[INFO] Deleting task %s %v", d.Id(), c)

	if _, err := c.session.DeleteModel("tasks", d.Id()); err != nil {
		return fmt.Errorf("error deleting task: %s", err)
	}

	return nil
}
