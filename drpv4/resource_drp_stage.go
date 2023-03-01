package drpv4

import (
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
func flattenStage(d *schema.ResourceData, stage *models.Stage) {
	d.Set("name", stage.Name)
	d.Set("description", stage.Description)
	d.Set("documentation", stage.Documentation)
	d.Set("bootenv", stage.BootEnv)
	d.Set("optional_params", stage.OptionalParams)
	d.Set("params", stage.Params)
	d.Set("profiles", stage.Profiles)
	d.Set("reboot", stage.Reboot)
	d.Set("required_params", stage.RequiredParams)
	d.Set("runner_wait", stage.RunnerWait)
	d.Set("tasks", stage.Tasks)
	d.Set("template", flattenTemplates(stage.Templates))
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
		Templates:      expandTemplateStage(d.Get("template").([]interface{})),
	}
	return stage
}

// expandTemplateStage expands a list of Templates from a Terraform ResourceData
func expandTemplateStage(templates interface{}) []models.TemplateInfo {
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
			Meta:     map[string]string{},
		}

		for k, v := range tpl["meta"].(map[string]interface{}) {
			t[i].Meta[k] = v.(string)
		}
	}
	return t
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

	stage := res.(*models.Stage)

	flattenStage(d, stage)

	return nil
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

	res, err := c.session.DeleteModel("stages", d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Stage delete: %#v", res)

	d.SetId("")

	return nil
}
