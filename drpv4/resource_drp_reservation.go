package drpv4

import (
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

// resourceReservation represents a reservation in DRP system
//
//		resource "drp_reservation" "reservation" {
//		  name = "reservation"
//		  description = "reservation"
//		  documenation = "reservation"
//		  addr = "192.168.0.1"
//	   duration = 86400
//	   next_server = "192.168.2.1"
//	   scoped = false
//	   strategy = "strategy"
//	   subnet = "subnet"
//	   token = "token"
//
//	   options {
//	     code = "value1"
//	     value = "value2"
//	   }
//
//	   options {
//	     code = "value3"
//	     value = "value4"
//	   }
//	}
func resourceReservation() *schema.Resource {
	r := &schema.Resource{
		Create: resourceReservationCreate,
		Read:   resourceReservationRead,
		Update: resourceReservationUpdate,
		Delete: resourceReservationDelete,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Description: "Reservation description",
				Optional:    true,
			},
			"documentation": {
				Type:        schema.TypeString,
				Description: "Reservation documentation",
				Optional:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "Reservation address",
				Required:    true,
				ForceNew:    true,
			},
			"duration": {
				Type:        schema.TypeInt,
				Description: "Reservation duration",
				Optional:    true,
			},
			"next_server": {
				Type:        schema.TypeString,
				Description: "Reservation next server",
				Optional:    true,
			},
			"scoped": {
				Type:        schema.TypeBool,
				Description: "Reservation scoped",
				Optional:    true,
				ForceNew:    true,
			},
			"strategy": {
				Type:        schema.TypeString,
				Description: "Reservation strategy",
				Optional:    true,
				Default:     "MAC",
			},
			"token": {
				Type:        schema.TypeString,
				Description: "Reservation token",
				Required:    true,
			},
			"options": {
				Type:        schema.TypeList,
				Description: "Reservation options",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeInt,
							Description: "Option code",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Option value",
							Required:    true,
						},
					},
				},
			},
		},
	}

	return r
}

// flattenReservationOptions converts a list of options to a list of maps
func flattenReservationOptions(options []models.DhcpOption) []interface{} {
	result := make([]interface{}, len(options))

	for i, option := range options {
		result[i] = map[string]interface{}{
			"code":  int32(option.Code),
			"value": option.Value,
		}
	}

	return result
}

// flattenReservation flattens a reservation object
func flattenReservation(d *schema.ResourceData, reservation *models.Reservation) {
	d.Set("description", reservation.Description)
	d.Set("documentation", reservation.Documentation)
	d.Set("address", reservation.Addr.String())
	d.Set("duration", reservation.Duration)

	if reservation.NextServer != nil {
		d.Set("next_server", reservation.NextServer.String())
	}

	d.Set("scoped", reservation.Scoped)
	d.Set("strategy", reservation.Strategy)
	d.Set("token", reservation.Token)

	if reservation.Options != nil {
		d.Set("options", flattenReservationOptions(reservation.Options))
	}
}

// expandReservationOptions expands the options list
func expandReservationOptions(options []interface{}) []models.DhcpOption {
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

// expandReservation expands a reservation object
func expandReservation(d *schema.ResourceData) *models.Reservation {
	reservation := &models.Reservation{
		Description:   d.Get("description").(string),
		Documentation: d.Get("documentation").(string),
		Addr:          net.ParseIP(d.Get("address").(string)),
		Duration:      int32(d.Get("duration").(int)),
		Scoped:        d.Get("scoped").(bool),
		Strategy:      d.Get("strategy").(string),
		Token:         d.Get("token").(string),
		Options:       expandReservationOptions(d.Get("options").([]interface{})),
	}

	if nextServer, ok := d.GetOk("next_server"); ok {
		reservation.NextServer = net.ParseIP(nextServer.(string))
	}

	return reservation
}

// resourceReservationCreate creates a reservation in the DRP system
func resourceReservationCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Creating reservation %s", d.Get("address").(string))

	reservation := expandReservation(d)

	if err := c.session.CreateModel(reservation); err != nil {
		return err
	}

	d.SetId(reservation.Addr.String())

	return resourceReservationRead(d, m)
}

// resourceReservationRead reads a reservation from the DRP system
func resourceReservationRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Reading reservation %s", d.Id())

	res, err := c.session.GetModel("reservations", d.Id())
	if err != nil {
		return err
	}

	reservation := res.(*models.Reservation)

	flattenReservation(d, reservation)

	return nil
}

// resourceReservationUpdate updates a reservation in the DRP system
func resourceReservationUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Updating reservation %s", d.Id())

	reservation := expandReservation(d)

	if err := c.session.PutModel(reservation); err != nil {
		return err
	}

	return resourceReservationRead(d, m)
}

// resourceReservationDelete deletes a reservation from the DRP system
func resourceReservationDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[DEBUG] Deleting reservation %s", d.Id())

	if _, err := c.session.DeleteModel("reservations", d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
