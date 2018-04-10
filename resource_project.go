package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"time"
)

// This JSON structure is used only when creating a new project. A different
// structure is used when retrieving/updating already-existing projects
type NewProject struct {
	Name        string `json:"projectName,omitempty"`
	Description string `json:"projectDescription,omitempty"`
	Type        string `json:"processTemplateTypeId,omitempty"` // Agile = 'adcc42ab-9882-485e-a3ed-7678f01f66bc'
	Source      string `json:"source,omitempty"`
	Data        string `json:"projectData,omitempty"` // '{"VersionControlOption":"Git","ProjectVisibilityOption":null}'
}

// This JSON structure is used when retrieving info about one or more existing
// projects, or when updating a single one. A different structure is used when
// creating one.
type Project struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	State      string `json:"state"`
	Revision   int    `json:"revision"`
	Visibility string `json:"visibility"`
}

type ProjectSet struct {
	Count    int       `json:"count"`
	Projects []Project `json:"value"`
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NewProjectCreation",
			},
			"data": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func newProjectFromResource(d *schema.ResourceData) *NewProject {
	proj := &NewProject{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
		Source:      d.Get("source").(string),
		Data:        d.Get("data").(string),
	}

	return proj
}

func getAllProjects(m interface{}) (*ProjectSet, error) {
	log.Printf("[DEBUG] going to read all projects")
	client := m.(*VSTSClient)

	req, err := client.Get("_apis/projects/")

	if err != nil {
		return nil, err
	}

	if req.StatusCode == 200 {
		log.Printf("[DEBUG] got success reading all projects")
		var projects ProjectSet
		body, readerr := ioutil.ReadAll(req.Body)
		if readerr != nil {
			return nil, readerr
		}
		log.Printf("[DEBUG] all projects: %s", body)

		decodeerr := json.Unmarshal(body, &projects)
		if decodeerr != nil {
			return nil, decodeerr
		}
		return &projects, nil
	}
	log.Printf("[DEBUG] error reading all projects: %s", req.StatusCode)
	return nil, errors.New(fmt.Sprintf("Got HTTP status %d when reading all projects", req.StatusCode))

}

func getProjectById(id string, m interface{}) (*Project, error) {
	client := m.(*VSTSClient)
	proj_req, err := client.Get(fmt.Sprintf("_apis/projects/%s", id))

	if err != nil {
		return nil, err
	}

	if proj_req.StatusCode == 200 {
		var proj Project
		body, readerr := ioutil.ReadAll(proj_req.Body)
		if readerr != nil {
			return nil, readerr
		}

		log.Printf("[DEBUG] Read all project info: >>%s<<", body)
		decodeerr := json.Unmarshal(body, &proj)
		if decodeerr != nil {
			return nil, decodeerr
		}
		log.Printf("[DEBUG] decoded json: %s", proj)
		return &proj, nil
	}
	return nil, errors.New(fmt.Sprintf("Project id %s Not found", id))
}

func getProjectByName(name string, m interface{}) (*Project, error) {
	allProjects, err := getAllProjects(m)

	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Going to iterate over the list of projects")
	log.Printf("[DEBUG] This list is : %s", allProjects.Projects)
	for _, p := range allProjects.Projects {
		log.Printf("[DEBUG] Checking %s", p)
		if p.Name == name {
			// At the time of writing, the API returns the same
			// attributes when reading the index as it does when
			// reading info for just one. Though it's more calls,
			// it feels more logical  to retrieve the info for the
			// needed item and then return that
			proj, err := getProjectById(p.Id, m)

			if err != nil {
				return nil, err
			}
			return proj, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Project %s Not found", name))
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	client := m.(*VSTSClient)
	proj := newProjectFromResource(d)
	bytedata, err := json.Marshal(proj)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Going to create the project (POST)")
	_, err = client.Post("_api/_project/CreateProject",
		bytes.NewBuffer(bytedata))

	log.Printf("[DEBUG] POST request made. Checking whether that was err or success")
	if err != nil {
		log.Printf("[DEBUG] Error when creating project")
		return err
	}

	log.Printf("[DEBUG] created project. sleeping")
	time.Sleep(30 * time.Second)
	log.Printf("[DEBUG] Now going to see if I can look up the ID for that")
	p, err := getProjectByName(name, m)

	if err != nil {
		log.Printf("[DEBUG] Weird. We got a success when creating but could read back what we created. Maybe sleep for more time?")
		return err
	}

	log.Printf("[DEBUG] created project. id for this is %s", p.Id)
	d.SetId(p.Id)
	return nil
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] id for this is %s", id)
	// var projectSlug string
	// 	projectSlug = d.Get("slug").(string)
	// if projectSlug == "" {
	// 	projectSlug = d.Get("name").(string)
	// }

	proj, err := getProjectById(id, m)
	if err != nil {
		return err
	}

	d.Set("name", proj.Name)
	d.Set("url", proj.URL)
	d.Set("state", proj.State)
	d.Set("revision", proj.Revision)
	d.Set("visibility", proj.Visibility)

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("name") {
		// I could update this via the interactive webinterface but ugh.
		return errors.New(fmt.Sprintf("There is no VSTS API endpoint to change the name. Please make this change via the web interface"))
	}
	// err := nil
	d.Partial(false)
	// if err != nil {
	// 	return err
	// }

	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] At the top of the ProjectDelete method")
	log.Printf("[DEBUG] Trace 0")
	client := m.(*VSTSClient)
	log.Printf("[DEBUG] Trace 1")
	name := d.Get("name").(string)
	log.Printf("[DEBUG] Trace 2")
	id := d.Id() /*Get("id").(string)*/
	log.Printf("[DEBUG] Trace 3")

	log.Printf("[DEBUG] Going to delete project %s (id: %s)", name, id)

	_, err := client.Delete(fmt.Sprintf("_apis/projects/%s", id))
	log.Printf("[DEBUG] Trace 4")

	if err != nil {
		log.Printf("[DEBUG] Could not delete project %s: %s", id, err)
	}

	return err
}
