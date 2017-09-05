package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/models"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform/helper/resource"
)

type Client struct {
	APIKey      string
	APIUser     string
	APIPassword string
	APIURL      string

	netClient *http.Client
}

/*
 * Builds a client object for this config
 */
func (c *Client) Client() (interface{}, error) {
	log.Println("[DEBUG] [Config.Client] Configuring the DRP API client")

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	c.netClient = netClient

	return c, nil
}

func (c *Client) buildRequest(method, path string, data io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, c.APIURL+"/api/v3/"+path, data)
	if err != nil {
		log.Printf("[DEBUG] [buildRequest] %s request error = %v\n", method, err)
		return nil, err
	}

	if c.APIKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.APIKey)
	} else {
		hdr := base64.StdEncoding.EncodeToString([]byte(c.APIUser + ":" + c.APIPassword))
		request.Header.Set("Authorization", "Basic "+hdr)
	}
	return request, nil
}

// GREG: Remove these when they are moved to models.
type Stat struct {
	// required: true
	Name string `json:"name"`
	// required: true
	Count int `json:"count"`
}

// swagger:model
type Info struct {
	// required: true
	Arch string `json:"arch"`
	// required: true
	Os string `json:"os"`
	// required: true
	Version string `json:"version"`
	// required: true
	Id string `json:"id"`
	// required: true
	ApiPort int `json:"api_port"`
	// required: true
	FilePort int `json:"file_port"`
	// required: true
	TftpEnabled bool `json:"tftp_enabled"`
	// required: true
	DhcpEnabled bool `json:"dhcp_enabled"`
	// required: true
	ProvisionerEnabled bool `json:"prov_enabled"`
	// required: true
	Stats []*Stat `json:"stats"`
}

type UserToken struct {
	Token string
	Info  Info
}

func (c *Client) getToken(machineId string) (string, error) {
	request, err := c.buildRequest("GET", "users/"+c.APIUser+"/token", nil)
	if err != nil {
		return "", err
	}

	q := request.URL.Query()
	q.Add("ttl", "3600")
	q.Add("scope", "machines")
	q.Add("specific", machineId)

	request.URL.RawQuery = q.Encode()

	if response, err := c.netClient.Do(request); err != nil {
		log.Printf("[DEBUG] [getToken] call error = %v\n", err)
		return "", err
	} else {
		defer response.Body.Close()

		// We aren't authorized
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("getToken: Unauthorized access")
		}

		// We got an error
		if response.StatusCode > 299 || response.StatusCode < 200 {
			berr := models.Error{}
			if err := json.NewDecoder(response.Body).Decode(&berr); err != nil {
				return "", err
			} else {
				return "", &berr
			}
		}

		// Gots data
		var data UserToken
		err := json.NewDecoder(response.Body).Decode(&data)
		if err != nil {
			return "", fmt.Errorf("getToken: unmarshall error: %v", err)
		}
		return data.Token, nil
	}
}

func (c *Client) doGet(path string, params url.Values, data interface{}) error {
	request, err := c.buildRequest("GET", path, nil)
	if err != nil {
		return err
	}

	q := request.URL.Query()
	q.Add("terraform/managed", "true")
	q.Add("terraform/allocated", "false")
	for _, s := range params["filters"] {
		arr := strings.SplitN(s, "=", 2)
		q.Add(arr[0], arr[1])
	}
	request.URL.RawQuery = q.Encode()

	if response, err := c.netClient.Do(request); err != nil {
		log.Printf("[DEBUG] [doGet] call error = %v\n", err)
		return err
	} else {
		defer response.Body.Close()

		// We aren't authorized
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			return fmt.Errorf("Unauthorized access")
		}

		// We got an error
		if response.StatusCode > 299 || response.StatusCode < 200 {
			berr := models.Error{}
			if err := json.NewDecoder(response.Body).Decode(&berr); err != nil {
				return err
			} else {
				return &berr
			}
		}

		// Gots data
		return json.NewDecoder(response.Body).Decode(data)
	}
}
func (c *Client) doPatch(path string, patch jsonpatch2.Patch, data interface{}) error {
	jp, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("Failed to marshal patch: %v", err)
	}

	request, err := c.buildRequest("PATCH", path, bytes.NewBuffer(jp))
	if err != nil {
		log.Printf("[DEBUG] [doPatch] failed to build requiest error = %v\n", err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	if response, err := c.netClient.Do(request); err != nil {
		log.Printf("[DEBUG] [doPatch] call error = %v\n", err)
		return err
	} else {
		defer response.Body.Close()

		// We aren't authorized
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			log.Printf("[DEBUG] [doPatch] unauthorized\n")
			return fmt.Errorf("Unauthorized access")
		}

		// We got an error
		if response.StatusCode > 299 || response.StatusCode < 200 {
			berr := models.Error{}
			if err := json.NewDecoder(response.Body).Decode(&berr); err != nil {
				log.Printf("[DEBUG] [doPatch] responded error = %v\n", err)
				return err
			} else {
				log.Printf("[DEBUG] [doPatch] berr: responded error = %v\n", berr)
				return &berr
			}
		}

		// Gots data
		return json.NewDecoder(response.Body).Decode(data)
	}
}

// Gets all managed and unallocated machines (in addition to the other params)
func (c *Client) getAllMachines(params url.Values) ([]*models.Machine, error) {
	log.Printf("[DEBUG] [getAllMachines] Getting all machines from DRP\n")
	data := []*models.Machine{}
	return data, c.doGet("machines", params, &data)
}

func (c *Client) getSingleMachine(uuid string) (*models.Machine, error) {
	log.Printf("[DEBUG] [getSingleMachine] Getting a machine (%s) from DRP\n", uuid)
	data := &models.Machine{}
	return data, c.doGet("machines/"+uuid, map[string][]string{}, data)
}

func (c *Client) AllocateMachine(params url.Values) (*models.Machine, error) {
	log.Printf("[DEBUG] [allocateMachines] Allocating a machine with following params: %+v", params)
	for {
		if machines, err := c.getAllMachines(params); err != nil {
			return nil, err
		} else {
			if len(machines) == 0 {
				return nil, fmt.Errorf("No machines available")
			}

			patch := jsonpatch2.Patch{}

			if machines[0].Profile.Params["terraform/allocated"] == nil {
				p_test := jsonpatch2.Operation{Op: "test", Path: "/Profile/Params/terraform/allocated",
					From: "", Value: nil}
				patch = append(patch, p_test)

				p_add := jsonpatch2.Operation{Op: "add", Path: "/Profile/Params/terraform/allocated",
					From: "", Value: true}
				patch = append(patch, p_add)
			} else {
				p_test := jsonpatch2.Operation{Op: "test", Path: "/Profile/Params/terraform/allocated",
					From: "", Value: false}
				patch = append(patch, p_test)

				p_repl := jsonpatch2.Operation{Op: "replace", Path: "/Profile/Params/terraform/allocated",
					From: "", Value: true}
				patch = append(patch, p_repl)
			}

			machine := &models.Machine{}
			err = c.doPatch("machines/"+machines[0].UUID(), patch, machine)
			if err != nil {
				berr, ok := err.(*models.Error)
				if ok {
					// If we get a patch error, the machine was allocated while we were
					// waiting.  Try again.
					if berr.Type == "JsonPatchError" {
						continue
					}
				}
				return nil, err
			}

			return machine, nil
		}
	}
}

func (c *Client) ReleaseMachine(uuid string) error {
	log.Printf("[DEBUG] [releaseMachine] Releasing machine: %s", uuid)
	if machine, err := c.getSingleMachine(uuid); err != nil {
		return err
	} else {
		patch := jsonpatch2.Patch{}

		p_test := jsonpatch2.Operation{Op: "test", Path: "/Profile/Params/terraform/allocated",
			From: "", Value: true}
		patch = append(patch, p_test)

		p_repl := jsonpatch2.Operation{Op: "replace", Path: "/Profile/Params/terraform/allocated",
			From: "", Value: false}
		patch = append(patch, p_repl)

		err = c.doPatch("machines/"+machine.UUID(), patch, machine)
		if err != nil {
			return err
		}

		return nil
	}
}

// Update the machine to request position
func (c *Client) UpdateMachine(machineObj *models.Machine, constraints url.Values) error {
	oj, err := json.Marshal(machineObj)
	if err != nil {
		return err
	}

	// Apply the changes
	usingStages := false
	if machineObj.Profile.Params == nil {
		machineObj.Profile.Params = map[string]interface{}{}
	}
	if val, set := constraints["bootenv"]; set {
		machineObj.BootEnv = val[0]
	}
	if val, set := constraints["stage"]; set {
		usingStages = true
		machineObj.Stage = val[0]
	}
	if val, set := constraints["description"]; set {
		machineObj.Description = val[0]
	}
	if val, set := constraints["name"]; set {
		machineObj.Name = val[0]
	}
	if val, set := constraints["owner"]; set {
		machineObj.Profile.Params["terraform/owner"] = val[0]
	}

	userdata := map[string]interface{}{}
	if val, set := constraints["userdata"]; set {
		log.Printf("[DEBUG] [UpdateMachine] userdata: %v", val[0])
		err := yaml.Unmarshal([]byte(val[0]), &userdata)
		if err != nil {
			return fmt.Errorf("Error unmarshalling user-data: %v", err)
		}

		log.Printf("[DEBUG] [UpdateMachine] unmarshal: %+v", userdata)
	}

	token, err := c.getToken(machineObj.UUID())
	if err != nil {
		return fmt.Errorf("Failed to generate token: %v", err)
	}

	drpPhoneHome := fmt.Sprintf("/usr/local/bin/drpcli -E %s -T %s machines bootenv %s local",
		c.APIURL, strings.TrimSpace(token), machineObj.UUID())
	if usingStages {
		drpPhoneHome = fmt.Sprintf("/usr/local/bin/drpcli -E %s -T %s machines stage %s complete",
			c.APIURL, strings.TrimSpace(token), machineObj.UUID())
	}
	drpPhoneHolder := "JJJJJJJJJJJJDDDDDDDDDDD"

	obj, ok := userdata["runcmd"]
	if !ok {
		userdata["runcmd"] = append([]interface{}{}, drpPhoneHolder)
	} else {
		data, _ := obj.([]interface{})
		userdata["runcmd"] = append(data, drpPhoneHolder)
	}

	ud, err := yaml.Marshal(userdata)
	if err != nil {
		return fmt.Errorf("Error marshalling user-data: %v", err)
	}

	machineObj.Profile.Params["cloud-init/user-data"] = "#cloud-config\n" + strings.Replace(string(ud), drpPhoneHolder, drpPhoneHome, -1)

	if val, set := constraints["profiles"]; set {
		for _, p := range val {
			found := false
			for _, pp := range machineObj.Profiles {
				if pp == p {
					found = true
					break
				}
			}
			if !found {
				machineObj.Profiles = append(machineObj.Profiles, p)
			}
		}
	}

	if val, set := constraints["parameters"]; set {
		for _, parm := range val {
			arr := strings.SplitN(parm, "=", 2)

			// GREG: convert types from string to whatever
			machineObj.Profile.Params[arr[0]] = arr[1]
		}
	}

	nj, err := json.Marshal(machineObj)
	if err != nil {
		return err
	}

	patch, err := jsonpatch2.Generate(oj, nj, true)
	if err != nil {
		return fmt.Errorf("Error generating patch: %v", err)
	}

	return c.doPatch("machines/"+machineObj.UUID(), patch, machineObj)
}

func (c *Client) GetMachineStatus(uuid string) resource.StateRefreshFunc {
	log.Printf("[DEBUG] [getMachineStatus] Getting stat of machine: %s", uuid)
	return func() (interface{}, string, error) {
		machineObject, err := c.getSingleMachine(uuid)
		if err != nil {
			log.Printf("[ERROR] [getMachineStatus] Unable to get machine: %s\n", uuid)
			return nil, "", err
		}

		ta := machineObject.BootEnv
		machineStatus := "6"
		if ta != "local" {
			machineStatus = "9"
		}

		var statusRetVal bytes.Buffer
		statusRetVal.WriteString(machineStatus)
		statusRetVal.WriteString(":")

		return machineObject, statusRetVal.String(), nil
	}
}

func (c *Client) MachineDo(uuid, action string, params url.Values) error {
	log.Printf("[DEBUG] [machineDo] uuid: %s, action: %s, params: %+v", uuid, action, params)

	td := map[string]interface{}{}
	jp, err := json.Marshal(td)
	if err != nil {
		return fmt.Errorf("Failed to marshal empty map: %v", err)
	}
	request, err := c.buildRequest("POST", "machines/"+uuid+"/actions/"+action, bytes.NewBuffer(jp))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	if response, err := c.netClient.Do(request); err != nil {
		log.Printf("[DEBUG] [machineDo] call %s:%s error = %v\n", uuid, action, err)
		return err
	} else {
		defer response.Body.Close()

		// We aren't authorized
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			log.Printf("[DEBUG] [machineDo] unauthorized %s:%s\n", uuid, action)
			return fmt.Errorf("getToken: Unauthorized access")
		}

		// We got an error
		if response.StatusCode > 299 || response.StatusCode < 200 {
			berr := models.Error{}
			if err := json.NewDecoder(response.Body).Decode(&berr); err != nil {
				log.Printf("[DEBUG] [machineDo] returned %s:%s error = %v\n", uuid, action, err)
				return err
			} else {
				log.Printf("[DEBUG] [machineDo] returned %s:%s error = %v\n", uuid, action, &berr)
				return &berr
			}
		}
	}
	return nil
}