package drpv4

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceWorkflow is the Terraform resource for DRP Workflows
func resourceWorkflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkflowCreate,
		Read:   resourceWorkflowRead,
		Update: resourceWorkflowUpdate,
		Delete: resourceWorkflowDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"documentation": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"stages": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// expandWorkflow expands a Terraform resource into a DRP Workflow
func expandWorkflow(d *schema.ResourceData) *models.Workflow {
	workflow := &models.Workflow{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Documentation: d.Get("documentation").(string),
		Stages:        expandStringList(d.Get("stages")),
	}

	return workflow
}

// flattenWorkflow flattens a DRP Workflow into a Terraform resource
func flattenWorkflow(d *schema.ResourceData, workflow *models.Workflow) error {
	d.Set("name", workflow.Name)
	d.Set("description", workflow.Description)
	d.Set("documentation", workflow.Documentation)
	d.Set("stages", workflow.Stages)
	return nil
}

// resourceWorkflowCreate creates a new DRP Workflow
func resourceWorkflowCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	workflow := expandWorkflow(d)

	log.Printf("[DEBUG] Creating DRP Workflow: %s", workflow.Name)

	if err := c.session.CreateModel(workflow); err != nil {
		return err
	}

	d.SetId(workflow.Name)

	return resourceWorkflowRead(d, m)
}

// resourceWorkflowRead reads a DRP Workflow
func resourceWorkflowRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	res, err := c.session.GetModel("workflows", d.Id())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			d.SetId("")
			return flattenWorkflow(d, &models.Workflow{})
		} else {
			return fmt.Errorf("error reading param: %s", err)
		}
	}

	workflow := res.(*models.Workflow)

	flattenWorkflow(d, workflow)

	return nil
}

// resourceWorkflowUpdate updates a DRP Workflow
func resourceWorkflowUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	workflow := expandWorkflow(d)

	log.Printf("[DEBUG] Updating DRP Workflow: %s", workflow.Name)

	if err := c.session.PutModel(workflow); err != nil {
		return err
	}

	return resourceWorkflowRead(d, m)
}

// resourceWorkflowDelete deletes a DRP Workflow
func resourceWorkflowDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Deleting DRP Workflow: %s", d.Id())

	_, err := c.session.DeleteModel("workflows", d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
