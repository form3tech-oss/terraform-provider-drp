package drpv4

import (
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceSubnet represents a subnet in the DRP system
//
//	resource "drp_subnet" "subnet" {
//	  name = "subnet"
//	  description = "subnet"
//	  documenation = "subnet"
//	  enabled = true
//	  subnet = "255.255.255.0"
//	  active_start = "192.168.1.1"
//	  active_end = "192.168.1.255"
//	  active_lease_time = 86400
//	  next_server = "192.168.2.1"
//	  only_reservations = false
//
//	  options {
//	    code = "value1"
//	    value = "value2"
//	  }
//
//	  options {
//	    code = "value3"
//	    value = "value4"
//	  }
//
//	  pickers = ["picker1", "picker2"]
//	  proxy = false
//	  reserved_lease_time = 86400
//	  strategy = "strategy"
//	  unmanaged = false
//	}
func resourceSubnet() *schema.Resource {
	r := &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Subnet name",
				ForceNew:    true,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Subnet description",
				Optional:    true,
			},
			"documentation": {
				Type:        schema.TypeString,
				Description: "Subnet documentation",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Subnet enabled",
				Optional:    true,
			},
			"subnet": {
				Type:        schema.TypeString,
				Description: "Subnet subnet",
				Required:    true,
			},
			"active_start": {
				Type:        schema.TypeString,
				Description: "Subnet active start",
				Required:    true,
			},
			"active_end": {
				Type:        schema.TypeString,
				Description: "Subnet active end",
				Required:    true,
			},
			"active_lease_time": {
				Type:        schema.TypeInt,
				Description: "Subnet active lease time",
				Optional:    true,
				Default:     60,
			},
			"next_server": {
				Type:        schema.TypeString,
				Description: "Subnet next server",
				Optional:    true,
			},
			"only_reservations": {
				Type:        schema.TypeBool,
				Description: "Subnet only reservations",
				Optional:    true,
			},
			"options": {
				Type:        schema.TypeList,
				Description: "Subnet options",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeInt,
							Description: "Subnet option code",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Subnet option value",
							Required:    true,
						},
					},
				},
			},
			"pickers": {
				Type:        schema.TypeList,
				Description: "Subnet pickers",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"proxy": {
				Type:        schema.TypeBool,
				Description: "Subnet proxy",
				Optional:    true,
			},
			"reserved_lease_time": {
				Type:        schema.TypeInt,
				Description: "Subnet reserved lease time",
				Optional:    true,
				Default:     7200,
			},
			"strategy": {
				Type:        schema.TypeString,
				Description: "Subnet strategy",
				Optional:    true,
				Default:     "MAC",
			},
			"unmanaged": {
				Type:        schema.TypeBool,
				Description: "Subnet unmanaged",
				Optional:    true,
			},
		},
	}

	return r
}

// flattenSubnetOptions flattens the options list
func flattenSubnetOptions(options []models.DhcpOption) []interface{} {
	result := make([]interface{}, len(options))

	for i, option := range options {
		result[i] = map[string]interface{}{
			"code":  int32(option.Code),
			"value": option.Value,
		}
	}

	return result
}

// flattenSubnet flattens the subnet object
func flattenSubnet(d *schema.ResourceData, subnet *models.Subnet) {
	d.Set("name", subnet.Name)
	d.Set("description", subnet.Description)
	d.Set("documentation", subnet.Documentation)
	d.Set("enabled", subnet.Enabled)
	d.Set("subnet", subnet.Subnet)
	d.Set("active_start", subnet.ActiveStart.String())
	d.Set("active_end", subnet.ActiveEnd.String())
	d.Set("active_lease_time", subnet.ActiveLeaseTime)

	if subnet.NextServer != nil {
		d.Set("next_server", subnet.NextServer.String())
	}

	d.Set("only_reservations", subnet.OnlyReservations)

	if subnet.Options != nil {
		d.Set("options", flattenSubnetOptions(subnet.Options))
	}

	if subnet.Pickers != nil {
		d.Set("pickers", subnet.Pickers)
	}

	d.Set("proxy", subnet.Proxy)
	d.Set("reserved_lease_time", subnet.ReservedLeaseTime)
	d.Set("strategy", subnet.Strategy)
	d.Set("unmanaged", subnet.Unmanaged)
}

// expandSubnetOptions expands the options list
func expandSubnetOptions(options []interface{}) []models.DhcpOption {
	result := make([]models.DhcpOption, len(options))

	for i, option := range options {
		data := option.(map[string]interface{})

		result[i] = models.DhcpOption{
			Code:  byte(data["code"].(int)),
			Value: data["value"].(string),
		}
	}

	return result
}

// expandSubnet expands the subnet object
func expandSubnet(d *schema.ResourceData) *models.Subnet {
	subnet := &models.Subnet{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		Documentation:     d.Get("documentation").(string),
		Enabled:           d.Get("enabled").(bool),
		Subnet:            d.Get("subnet").(string),
		ActiveStart:       net.ParseIP(d.Get("active_start").(string)),
		ActiveEnd:         net.ParseIP(d.Get("active_end").(string)),
		ActiveLeaseTime:   int32(d.Get("active_lease_time").(int)),
		NextServer:        net.ParseIP(d.Get("next_server").(string)),
		OnlyReservations:  d.Get("only_reservations").(bool),
		Options:           expandSubnetOptions(d.Get("options").([]interface{})),
		Pickers:           expandStringList(d.Get("pickers")),
		Proxy:             d.Get("proxy").(bool),
		ReservedLeaseTime: int32(d.Get("reserved_lease_time").(int)),
		Strategy:          d.Get("strategy").(string),
		Unmanaged:         d.Get("unmanaged").(bool),
	}

	return subnet
}

// resourceSubnetCreate creates the subnet resource
func resourceSubnetCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Creating subnet: %s", d.Get("name").(string))

	subnet := expandSubnet(d)

	err := c.session.CreateModel(subnet)
	if err != nil {
		return err
	}

	d.SetId(subnet.Name)

	return resourceSubnetRead(d, m)
}

// resourceSubnetRead reads the subnet resource
func resourceSubnetRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Reading subnet: %s", d.Id())

	res, err := c.session.GetModel("subnets", d.Id())
	if err != nil {
		return err
	}

	subnet := res.(*models.Subnet)

	flattenSubnet(d, subnet)

	return nil
}

// resourceSubnetUpdate updates the subnet resource
func resourceSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	subnet := expandSubnet(d)

	log.Printf("[DEBUG] Updating subnet: %#v", subnet)

	err := c.session.PutModel(subnet)
	if err != nil {
		return err
	}

	return resourceSubnetRead(d, m)
}

// resourceSubnetDelete deletes the subnet resource
func resourceSubnetDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Deleting subnet: %s", d.Id())

	res, err := c.session.DeleteModel("subnets", d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleted subnet: %#v", res)

	d.SetId("")

	return nil
}
