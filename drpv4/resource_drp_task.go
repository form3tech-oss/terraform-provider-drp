package drpv4

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

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

func expandTaskConfig(d *schema.ResourceData) *models.Task {
	task := models.Task{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		RequiredParams: []string{},
		OptionalParams: []string{},
		ExtraRoles:     []string{},
		Prerequisites:  []string{},
	}

	if requiredParams, ok := d.GetOk("required_params"); ok {
		for _, param := range requiredParams.([]interface{}) {
			task.RequiredParams = append(task.RequiredParams, param.(string))
		}
	}

	if optionalParams, ok := d.GetOk("optional_params"); ok {
		for _, param := range optionalParams.([]interface{}) {
			task.OptionalParams = append(task.OptionalParams, param.(string))
		}
	}

	if extraRoles, ok := d.GetOk("extra_roles"); ok {
		for _, role := range extraRoles.([]interface{}) {
			task.ExtraRoles = append(task.ExtraRoles, role.(string))
		}
	}

	if prerequisites, ok := d.GetOk("prerequisites"); ok {
		for _, prereq := range prerequisites.([]interface{}) {
			task.Prerequisites = append(task.Prerequisites, prereq.(string))
		}
	}

	templates := d.Get("templates").([]interface{})
	task.Templates = make([]models.TemplateInfo, len(templates))
	for i, t := range templates {
		template := t.(map[string]interface{})
		templateInfo := models.TemplateInfo{
			ID:       template["template_id"].(string),
			Name:     template["name"].(string),
			Path:     template["path"].(string),
			Contents: template["contents"].(string),
			Link:     template["link"].(string),
			Meta:     map[string]string{},
		}

		if meta, ok := template["meta"]; ok {
			for k, v := range meta.(map[string]interface{}) {
				templateInfo.Meta[k] = v.(string)
			}
		}

		task.Templates[i] = templateInfo
	}

	if claims, ok := d.GetOk("extra_claims"); ok {
		for _, c := range claims.(*schema.Set).List() {
			claim := c.(map[string]interface{})
			task.ExtraClaims = append(task.ExtraClaims, &models.Claim{
				Scope:    claim["scope"].(string),
				Action:   claim["action"].(string),
				Specific: claim["specific"].(string),
			})
		}
	}
	return &task
}

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

func resourceTaskCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	task := expandTaskConfig(d)

	task.Validate()
	if task.Error() != "" {
		return fmt.Errorf("error validating task: %s", task.Error())
	}

	log.Printf("[INFO] Creating task %s", d.Id())

	req := c.session.Req().Post(task).UrlFor("tasks")
	if err := req.Do(&task); err != nil {
		return fmt.Errorf("error creating task: %s", err)
	}

	if task.Error() != "" {
		return fmt.Errorf("error creating task: %s", task.Error())
	}

	d.SetId(task.Name)

	return nil
}

func resourceTaskRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[INFO] Reading task %s", d.Id())

	to, err := c.session.GetModel("tasks", d.Id())
	if err != nil {
		return fmt.Errorf("error reading task: %s", err)
	}

	taskObject := to.(*models.Task)

	if taskObject.Error() != "" {
		return fmt.Errorf("error reading task: %s", taskObject.Error())
	}

	if err := d.Set("name", taskObject.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err := d.Set("description", taskObject.Description); err != nil {
		return fmt.Errorf("error setting description: %s", err)
	}

	if err := d.Set("required_params", taskObject.RequiredParams); err != nil {
		return fmt.Errorf("error setting required_params: %s", err)
	}

	if err := d.Set("optional_params", taskObject.OptionalParams); err != nil {
		return fmt.Errorf("error setting optional_params: %s", err)
	}

	if err := d.Set("extra_roles", taskObject.ExtraRoles); err != nil {
		return fmt.Errorf("error setting extra_roles: %s", err)
	}

	if err := d.Set("prerequisites", taskObject.Prerequisites); err != nil {
		return fmt.Errorf("error setting prerequisites: %s", err)
	}

	if err := d.Set("templates", flattenTemplates(taskObject.Templates)); err != nil {
		return fmt.Errorf("error setting templates: %s", err)
	}

	if err := d.Set("extra_claims", flattenClaims(taskObject.ExtraClaims)); err != nil {
		return fmt.Errorf("error setting extra_claims: %s", err)
	}

	return nil
}

func resourceTaskUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	task := expandTaskConfig(d)

	task.Validate()
	if task.Error() != "" {
		return fmt.Errorf("error validating task: %s", task.Error())
	}

	log.Printf("[INFO] Updating task %s", d.Id())

	err := c.session.PutModel(task)
	if err != nil {
		return fmt.Errorf("error updating task: %s", err)
	}

	return nil
}

func resourceTaskDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[INFO] Deleting task %s %v", d.Id(), c)

	_, err := c.session.DeleteModel("tasks", d.Id())
	if err != nil {
		return fmt.Errorf("error deleting task: %s", err)
	}

	return nil
}
