package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"fmt"
	"log"
	"time"

	"github.com/VictorLowther/jsonpatch2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

func resourceMachinePool() *schema.Resource {
	r := &schema.Resource{
		Create: resourceMachineSetPool,
		Read:   resourceMachineGetPool,
		Update: resourceMachineSetPoolUpdate,
		Delete: resourceMachineSetPoolDelete,

		Schema: map[string]*schema.Schema{

			"pool": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "default",
				Description: "Which pool to add machine to",
				Optional:    true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Machine Name",
				Required:    true,
				ForceNew:    true,
			},

			// Machine.Address
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Returns Digital Rebar Machine.Address",
				Computed:    true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}

	return r
}

func resourceMachineSetPool(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineAllocate] Allocating new drp_machine")
	cc := m.(*Config)

	pool := d.Get("pool").(string)
	if pool == "" {
		pool = "default"
	}
	d.Set("pool", pool)
	name := d.Get("name").(string)

	requuid := cc.session.Req().Get().UrlFor("machines", "Name=", name)
	mruuid := []*models.Machine{}
	if err := requuid.Do(&mruuid); err != nil {
		log.Printf("[DEBUG] Get error %+v | %+v", err, requuid)
		return fmt.Errorf("error getting machine UUID for address %s: %s", name, err)
	}
	patch := jsonpatch2.Patch{{Op: "replace", Path: "/Pool", Value: pool}}
	reqm := cc.session.Req().Patch(patch).UrlFor("machines", string(mruuid[0].Uuid))
	mr := []*models.Machine{}
	if err := reqm.Do(&mr); err != nil {
		log.Printf("[DEBUG] POST error %+v | %+v", err, reqm)
		return fmt.Errorf("error set pool %s: %s", pool, err)
	}

	d.Set("address", mr[0].Address)
	d.SetId(string(mr[0].Uuid))
	return resourceMachineGetPool(d, m)
}

func resourceMachineGetPool(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineGetPool] Reading drp_machine")
	cc := m.(*Config)
	uuid := d.Id()
	log.Printf("[DEBUG] Reading machine %s", uuid)
	mo, err := cc.session.GetModel("machines", uuid)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineRead] Unable to get machine: %s", uuid)
		return fmt.Errorf("Unable to get machine %s", uuid)
	}
	machineObject := mo.(*models.Machine)
	d.Set("address", machineObject.Address.String())

	return nil
}

func resourceMachineSetPoolUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineUpdate] Updating drp_machine")
	cc := m.(*Config)

	// at this time there are no updates
	log.Printf("[DEBUG] Config %v", cc)

	return resourceMachineGetPool(d, m)
}

func resourceMachineSetPoolDelete(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineRelease] Releasing machine_set_pool")
	cc := m.(*Config)

	uuid := d.Id()
	if uuid == "" {
		return fmt.Errorf("Requires Uuid from id")
	}
	log.Printf("[DEBUG] Releasing %s ", uuid)

	patch := jsonpatch2.Patch{{Op: "replace", Path: "/Pool", Value: "default"}}
	reqm := cc.session.Req().Patch(patch).UrlFor("machines", uuid)
	mr := []*models.Machine{}
	if err := reqm.Do(&mr); err != nil {
		log.Printf("[DEBUG] POST error %+v | %+v", err, reqm)
		return fmt.Errorf("error set pool %s: %s", "default", err)
	}

	mc := mr[0]
	if mc.Pool == "default" {
		d.SetId("")
		return nil
	} else {
		return fmt.Errorf("Could not set default pool for  %s", uuid)
	}
}
