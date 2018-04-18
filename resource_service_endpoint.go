package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
)

type ServiceEndpointSet struct {
	Count            int               `json:"count"`
	ServiceEndpoints []ServiceEndpoint `json:"value"`
}

type ServiceEndpoint struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func resourceServiceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceEndpointCreate,
		Read:   resourceServiceEndpointRead,
		Update: resourceServiceEndpointUpdate,
		Delete: resourceServiceEndpointDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"project": &schema.Schema{
				// can be name or id; is used when constructing API URLs
				Type:     schema.TypeString,
				Required: true,
			},
			// README and .gitignore also exist as options but let's ignore them for now
		},
	}
}

// GET https://{accountName}.visualstudio.com/{project}/_apis/serviceendpoint/endpoints?type={type}&authSchemes={authSchemes}&endpointIds={endpointIds}&includeFailed={includeFailed}&api-version=4.1-preview.1
func newServiceEndpointFromResource(d *schema.ResourceData) *ServiceEndpoint {
	endpoint := &ServiceEndpoint{
		Name: d.Get("name").(string),
	}

	return endpoint
}

func getServiceEndpoint(projectId string, endpointId string, m interface{}) (*ServiceEndpoint, error) {
	client := m.(*VSTSClient)

	endpoint_req, err := client.Get(fmt.Sprintf("/%s/_apis/serviceendpoint/endpoints/%s", projectId, endpointId))

	if err != nil {
		return nil, err
	}

	if endpoint_req.StatusCode == 200 {
		var endpoint ServiceEndpoint
		body, readerr := ioutil.ReadAll(endpoint_req.Body)
		if readerr != nil {
			return nil, readerr
		}

		log.Printf("[DEBUG] Read all endpoint info: >>%s<<", body)
		decodeerr := json.Unmarshal(body, &endpoint)
		if decodeerr != nil {
			return nil, decodeerr
		}
		log.Printf("[DEBUG] decoded json: %s", endpoint)
		return &endpoint, nil
	}
	return nil, errors.New(fmt.Sprintf("Endpoint %s not found in project", nameOrId, projectId))
}

func resourceServiceEndpointCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	projId := d.Get("project").(string)
	client := m.(*VSTSClient)
	endpoint := newServiceEndpointFromResource(d)
	bytedata, err := json.Marshal(endpoint)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Going to create the endpoint (POST)")
	_, err = client.Post(fmt.Sprintf("%s/_apis/git/repositories", projId),
		bytes.NewBuffer(bytedata))

	log.Printf("[DEBUG] POST request made. Checking whether that was err or success")
	if err != nil {
		log.Printf("[DEBUG] Error when creating endpoint")
		return err
	}

	r, err := getProjectRepo(projId, name, m)

	if err != nil {
		log.Printf("[DEBUG] Weird. We got a success when creating but could read back what we created. Maybe sleep for more time?")
		return err
	}

	log.Printf("[DEBUG] created endpoint. id for this is %s", r.Id)
	d.SetId(r.Id)
	return nil
}

func resourceServiceEndpointRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] id for this is %s", id)

	endpoint, err := getProjectRepo(d.Get("project").(string), d.Get("name").(string), m)
	if err != nil {
		return err
	}

	d.Set("name", endpoint.Name)

	return nil
}

func resourceServiceEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*VSTSClient)
	projId := d.Get("project").(string)
	id := d.Id()
	endpoint := newServiceEndpointFromResource(d)
	if d.HasChange("name") {
		d.Partial(true)
		bytedata, err := json.Marshal(endpoint)
		if err != nil {
			return err
		}
		_, err = client.Patch(
			fmt.Sprintf("%s/_apis/git/repositories/%s", projId, id),
			bytes.NewBuffer(bytedata))
		if err != nil {
			return err
		}
	}
	d.Partial(false)

	return nil
}

func resourceServiceEndpointDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*VSTSClient)
	name := d.Get("name").(string)
	projId := d.Get("project").(string)
	id := d.Id()

	log.Printf("[DEBUG] Going to delete endpoint %s (id: %s)", name, id)

	_, err := client.Delete(fmt.Sprintf("%s/_apis/git/repositories/%s", projId, id))

	if err != nil {
		log.Printf("[DEBUG] Could not delete endpoint %s: %s", id, err)
	}

	return err
}
