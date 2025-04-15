package drpv4

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourcePool is the schema for the pool resource.
//
//	resource "drp_pool" "test" {
//	  pool_id = "test"
//	  description = "test pool"
//	  documentation = "test pool"
//	  parent_pool = "test"
//	  allocate_actions = {
//	     add_parameters = {
//	       "key0" = {}
//	       "key1" = {}
//		    }
//	     add_profiles = ["test"]
//	     remove_parameters = ["test"]
//		    remove_profiles = ["test"]
//	     workflow = "test"
//	  }
//	  release_actions = {
//	     add_parameters = {
//	       "key0" = {}
//	       "key1" = {}
//		    }
//	     add_profiles = ["test"]
//		    remove_profiles = ["test"]
//	     remove_parameters = ["test"]
//	     workflow = "test"
//	  }
//	  enter_actions = {
//	     add_parameters = {
//	       "key0" = {}
//	       "key1" = {}
//		    }
//	     add_profiles = ["test"]
//		    remove_profiles = ["test"]
//	     remove_parameters = ["test"]
//	     workflow = "test"
//	  }
//	  exit_actions = {
//	     add_parameters = {
//	       "key0" = {}
//	       "key1" = {}
//		    }
//	     add_profiles = ["test"]
//		    remove_profiles = ["test"]
//	     remove_parameters = ["test"]
//	     workflow = "test"
//	  }
//	  autofill = {
//	     acquire_pool = "test"
//	     create_parameters = {
//	       "key0" = {}
//	       "key1" = {}
//		    }
//		    max_free = 0
//	     min_free = 0
//	     return_pool = "test"
//	     use_autofill = true
//	  }
//	}
func resourcePool() *schema.Resource {
	actionResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"add_parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The pool add parameters.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"add_profiles": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool add profiles.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"remove_parameters": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool remove parameters.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"remove_profiles": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool remove profiles.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workflow": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pool workflow.",
			},
		},
	}

	r := &schema.Resource{
		Create: resourcePoolCreate,
		Read:   resourcePoolRead,
		Update: resourcePoolUpdate,
		Delete: resourcePoolDelete,

		Schema: map[string]*schema.Schema{
			"pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The pool id.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pool description.",
			},
			"documentation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pool documentation.",
			},
			"parent_pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pool parent pool.",
			},
			"allocate_actions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool allocate actions.",
				Elem:        actionResource,
				MaxItems:    1,
			},
			"release_actions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool release actions.",
				Elem:        actionResource,
				MaxItems:    1,
			},
			"enter_actions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool enter actions.",
				Elem:        actionResource,
				MaxItems:    1,
			},
			"exit_actions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool exit actions.",
				Elem:        actionResource,
				MaxItems:    1,
			},
			"autofill": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The pool autofill.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"acquire_pool": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The pool acquire pool.",
						},
						"create_parameters": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "The pool create parameters.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"max_free": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The pool max free.",
						},
						"min_free": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The pool min free.",
						},
						"return_pool": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The pool return pool.",
						},
						"use_autofill": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The pool use autofill.",
						},
					},
				},
			},
		},
	}

	return r
}

// flattenPoolActions flattens a PoolActions to a list of map.
func flattenPoolActions(actions *models.PoolTransitionActions) []map[string]interface{} {
	if actions == nil {
		return nil
	}

	m := make([]map[string]interface{}, 1)
	m[0] = map[string]interface{}{
		"add_parameters":    actions.AddParameters,
		"add_profiles":      actions.AddProfiles,
		"remove_parameters": actions.RemoveParameters,
		"remove_profiles":   actions.RemoveProfiles,
		"workflow":          actions.Workflow,
	}
	return m
}

// flattenPoolAutofill flattens a PoolAutofill to a map.
func flattenPoolAutofill(autofill *models.PoolAutoFill) []map[string]interface{} {
	if autofill == nil {
		return nil
	}

	m := make([]map[string]interface{}, 1)
	m[0] = map[string]interface{}{
		"acquire_pool": autofill.AcquirePool,
		"max_free":     autofill.MaxFree,
		"min_free":     autofill.MinFree,
		"return_pool":  autofill.ReturnPool,
		"use_autofill": autofill.UseAutoFill,
	}
	if autofill.CreateParameters != nil {
		m[0]["create_parameters"] = autofill.CreateParameters
	}

	return m
}

// flattenPool flattens a Pool to a Terraform ResourceData.
func flattenPool(d *schema.ResourceData, pool *models.Pool) error {
	if err := d.Set("pool_id", pool.Id); err != nil {
		return err
	}
	if err := d.Set("description", pool.Description); err != nil {
		return err
	}
	if err := d.Set("documentation", pool.Documentation); err != nil {
		return err
	}
	if err := d.Set("parent_pool", pool.ParentPool); err != nil {
		return err
	}
	if err := d.Set("allocate_actions", flattenPoolActions(pool.AllocateActions)); err != nil {
		return err
	}
	if err := d.Set("release_actions", flattenPoolActions(pool.ReleaseActions)); err != nil {
		return err
	}
	if err := d.Set("enter_actions", flattenPoolActions(pool.EnterActions)); err != nil {
		return err
	}
	if err := d.Set("exit_actions", flattenPoolActions(pool.ExitActions)); err != nil {
		return err
	}
	if err := d.Set("autofill", flattenPoolAutofill(pool.AutoFill)); err != nil {
		return err
	}

	return nil
}

// expandPoolActions expands a PoolTransictionActions from a map.
func expandPoolActions(actions []interface{}) *models.PoolTransitionActions {
	if len(actions) == 0 {
		return nil
	}

	a := actions[0].(map[string]interface{})

	action := &models.PoolTransitionActions{
		AddParameters:    a["add_parameters"].(map[string]interface{}),
		AddProfiles:      expandStringList(a["add_profiles"]),
		RemoveParameters: expandStringList(a["remove_parameters"]),
		RemoveProfiles:   expandStringList(a["remove_profiles"]),
		Workflow:         a["workflow"].(string),
	}

	return action
}

// expandPoolAutofill expands a PoolAutoFill from a map.
func expandPoolAutofill(autofill []interface{}) *models.PoolAutoFill {
	if len(autofill) == 0 {
		return nil
	}

	a := autofill[0].(map[string]interface{})

	af := &models.PoolAutoFill{
		AcquirePool:      a["acquire_pool"].(string),
		MaxFree:          int32(a["max_free"].(int)),
		MinFree:          int32(a["min_free"].(int)),
		ReturnPool:       a["return_pool"].(string),
		UseAutoFill:      a["use_autofill"].(bool),
		CreateParameters: a["create_parameters"].(map[string]interface{}),
	}

	return af
}

// expandPool expands the pool object.
func expandPool(d *schema.ResourceData) *models.Pool {
	pool := &models.Pool{
		Id:              d.Get("pool_id").(string),
		Description:     d.Get("description").(string),
		Documentation:   d.Get("documentation").(string),
		ParentPool:      d.Get("parent_pool").(string),
		AllocateActions: expandPoolActions(d.Get("allocate_actions").([]interface{})),
		ReleaseActions:  expandPoolActions(d.Get("release_actions").([]interface{})),
		EnterActions:    expandPoolActions(d.Get("enter_actions").([]interface{})),
		ExitActions:     expandPoolActions(d.Get("exit_actions").([]interface{})),
		AutoFill:        expandPoolAutofill(d.Get("autofill").([]interface{})),
	}

	return pool
}

// resourcePoolCreate creates a Pool resource.
func resourcePoolCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Creating Pool: %s", d.Get("pool_id").(string))

	pool := expandPool(d)

	if err := c.session.CreateModel(pool); err != nil {
		return err
	}

	d.SetId(pool.Id)

	return resourcePoolRead(d, m)
}

// resourcePoolRead reads a Pool resource.
func resourcePoolRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Reading Pool: %s", d.Id())

	pool, err := c.session.GetModel("pools", d.Id())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			d.SetId("")
			return flattenPool(d, &models.Pool{})
		} else {
			return fmt.Errorf("error reading pool: %s", err)
		}
	}

	return flattenPool(d, pool.(*models.Pool))
}

// resourcePoolUpdate updates a Pool resource.
func resourcePoolUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Updating Pool: %s", d.Id())

	pool := expandPool(d)

	if err := c.session.PutModel(pool); err != nil {
		return err
	}

	return resourcePoolRead(d, m)
}

// resourcePoolDelete deletes a Pool resource.
func resourcePoolDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Deleting Pool: %s", d.Id())

	if _, err := c.session.DeleteModel("pools", d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
