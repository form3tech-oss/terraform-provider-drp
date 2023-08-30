package drpv4

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// ResourceStage is the Schema for the stages API
func resourceStage() *schema.Resource {
	r := &schema.Resource{
		Create: resourceStageCreate,
		Read:   resourceStageRead,
		Update: resourceStageUpdate,
		Delete: resourceStageDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Stage name",
				ForceNew:    true,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Stage description",
				Optional:    true,
			},
			"documentation": {
				Type:        schema.TypeString,
				Description: "Stage documentation",
				Optional:    true,
			},
			"bootenv": {
				Type:        schema.TypeString,
				Description: "Stage bootenv",
				Optional:    true,
			},
			"optional_params": {
				Type:        schema.TypeList,
				Description: "Stage optional params",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"params": {
				Type:        schema.TypeMap,
				Description: "Stage params",
				Optional:    true,
			},
			"profiles": {
				Type:        schema.TypeList,
				Description: "Stage profiles",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"reboot": {
				Type:        schema.TypeBool,
				Description: "Stage reboot",
				Optional:    true,
			},
			"required_params": {
				Type:        schema.TypeList,
				Description: "Stage required params",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"runner_wait": {
				Type:        schema.TypeBool,
				Description: "Stage runner wait",
				Optional:    true,
				Computed:    true,
			},
			"tasks": {
				Type:        schema.TypeList,
				Description: "Stage tasks",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"template": {
				Type:        schema.TypeList,
				Description: "Stage templates",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Template name",

							Required: true,
						},
						"contents": {
							Type:        schema.TypeString,
							Description: "Template content",
							Optional:    true,
						},
						"path": {
							Type:        schema.TypeString,
							Description: "Template path",
							Optional:    true,
						},
						"template_id": {
							Type:        schema.TypeString,
							Description: "Template ID",
							Optional:    true,
						},
						"link": {
							Type:        schema.TypeString,
							Description: "Template link",
							Optional:    true,
						},
						"meta": {
							Type:        schema.TypeMap,
							Description: "Template meta",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}

	return r
}

// flattenStage flattens a Stage from a Terraform ResourceData
func flattenStage(d *schema.ResourceData, stage *models.Stage) error {
	if err := d.Set("name", stage.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}
	if err := d.Set("description", stage.Description); err != nil {
		return fmt.Errorf("error setting description: %s", err)
	}
	if err := d.Set("documentation", stage.Documentation); err != nil {
		return fmt.Errorf("error setting documentation: %s", err)
	}
	if err := d.Set("bootenv", stage.BootEnv); err != nil {
		return fmt.Errorf("error setting bootenv: %s", err)
	}
	if err := d.Set("optional_params", stage.OptionalParams); err != nil {
		return fmt.Errorf("error setting optional_params: %s", err)
	}
	if err := d.Set("params", stage.Params); err != nil {
		return fmt.Errorf("error setting params: %s", err)
	}
	if err := d.Set("profiles", stage.Profiles); err != nil {
		return fmt.Errorf("error setting profiles: %s", err)
	}
	if err := d.Set("reboot", stage.Reboot); err != nil {
		return fmt.Errorf("error setting reboot: %s", err)
	}
	if err := d.Set("required_params", stage.RequiredParams); err != nil {
		return fmt.Errorf("error setting required_params: %s", err)
	}
	if err := d.Set("runner_wait", stage.RunnerWait); err != nil {
		return fmt.Errorf("error setting runner_wait: %s", err)
	}
	if err := d.Set("tasks", stage.Tasks); err != nil {
		return fmt.Errorf("error setting tasks: %s", err)
	}
	if err := d.Set("template", flattenTemplates(stage.Templates)); err != nil {
		return fmt.Errorf("error setting template: %s", err)
	}
	return nil
}

// expandStageTemplates expands a list of Templates from a Terraform ResourceData
func expandStageTemplates(templates interface{}) []models.TemplateInfo {
	if templates == nil {
		return nil
	}

	t := make([]models.TemplateInfo, len(templates.([]interface{})))
	for i, template := range templates.([]interface{}) {
		tpl := template.(map[string]interface{})

		t[i] = models.TemplateInfo{
			Name:     tpl["name"].(string),
			Contents: tpl["contents"].(string),
			Path:     tpl["path"].(string),
			ID:       tpl["template_id"].(string),
			Link:     tpl["link"].(string),
			Meta:     expandMapInterface(tpl["meta"].(map[string]interface{})),
		}
	}

	return t
}

// expandStage expands a Stage from a Terraform ResourceData
func expandStage(d *schema.ResourceData) *models.Stage {
	stage := &models.Stage{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Documentation:  d.Get("documentation").(string),
		BootEnv:        d.Get("bootenv").(string),
		OptionalParams: expandStringList(d.Get("optional_params")),
		Params:         d.Get("params").(map[string]interface{}),
		Profiles:       expandStringList(d.Get("profiles")),
		Reboot:         d.Get("reboot").(bool),
		RequiredParams: expandStringList(d.Get("required_params")),
		RunnerWait:     d.Get("runner_wait").(bool),
		Tasks:          expandStringList(d.Get("tasks")),
		Templates:      expandStageTemplates(d.Get("template")),
	}
	return stage
}

// resourceStageCreate creates a new Stage
func resourceStageCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	stage := expandStage(d)

	log.Printf("[DEBUG] Stage create: %#v", stage)

	if err := c.session.CreateModel(stage); err != nil {
		return err
	}

	d.SetId(stage.Name)

	return resourceStageRead(d, m)
}

// resourceStageRead reads a Stage
func resourceStageRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Stage read: %s", d.Id())

	res, err := c.session.GetModel("stages", d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Stage read: %#v", res)

	return flattenStage(d, res.(*models.Stage))
}

// resourceStageUpdate updates a Stage
func resourceStageUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	stage := expandStage(d)

	log.Printf("[DEBUG] Stage update: %#v", stage)

	err := c.session.PutModel(stage)
	if err != nil {
		return err
	}

	return resourceStageRead(d, m)
}

// resourceStageDelete deletes a Stage
func resourceStageDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Stage delete: %s", d.Id())

	_, err := c.session.DeleteModel("stages", d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
