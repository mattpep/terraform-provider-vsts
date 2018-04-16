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

type RepoSet struct {
	Count int          `json:"count"`
	Repos []Repository `json:"value"`
}

type Repository struct {
	Id            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	URL           string `json:"url,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	RemoteURL     string `json:"remoteUrl,omitempty"`
	SSHURL        string `json:"sshUrl,omitempty"`
	// Project       Project `json:"project,omitempty"`
}

func resourceRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryCreate,
		Read:   resourceRepositoryRead,
		Update: resourceRepositoryUpdate,
		Delete: resourceRepositoryDelete,

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

func newRepositoryFromResource(d *schema.ResourceData) *Repository {
	repo := &Repository{
		Name: d.Get("name").(string),
	}

	return repo
}

func getProjectRepo(projectId string, nameOrId string, m interface{}) (*Repository, error) {
	client := m.(*VSTSClient)
	repo_req, err := client.Get(fmt.Sprintf("%s/_apis/git/repositories/%s", projectId, nameOrId))

	if err != nil {
		return nil, err
	}

	if repo_req.StatusCode == 200 {
		var repo Repository
		body, readerr := ioutil.ReadAll(repo_req.Body)
		if readerr != nil {
			return nil, readerr
		}

		log.Printf("[DEBUG] Read all repo info: >>%s<<", body)
		decodeerr := json.Unmarshal(body, &repo)
		if decodeerr != nil {
			return nil, decodeerr
		}
		log.Printf("[DEBUG] decoded json: %s", repo)
		return &repo, nil
	}
	return nil, errors.New(fmt.Sprintf("Repo %s not found in project", nameOrId, projectId))
}

func resourceRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	projId := d.Get("project").(string)
	client := m.(*VSTSClient)
	repo := newRepositoryFromResource(d)
	bytedata, err := json.Marshal(repo)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Going to create the repo (POST)")
	_, err = client.Post(fmt.Sprintf("%s/_apis/git/repositories", projId),
		bytes.NewBuffer(bytedata))

	log.Printf("[DEBUG] POST request made. Checking whether that was err or success")
	if err != nil {
		log.Printf("[DEBUG] Error when creating repo")
		return err
	}

	r, err := getProjectRepo(projId, name, m)

	if err != nil {
		log.Printf("[DEBUG] Weird. We got a success when creating but could read back what we created. Maybe sleep for more time?")
		return err
	}

	log.Printf("[DEBUG] created repo. id for this is %s", r.Id)
	d.SetId(r.Id)
	return nil
}

func resourceRepositoryRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] id for this is %s", id)

	repo, err := getProjectRepo(d.Get("project").(string), d.Get("name").(string), m)
	if err != nil {
		return err
	}

	d.Set("name", repo.Name)

	return nil
}

func resourceRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*VSTSClient)
	projId := d.Get("project").(string)
	id := d.Id()
	repo := newRepositoryFromResource(d)
	if d.HasChange("name") {
		d.Partial(true)
		bytedata, err := json.Marshal(repo)
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

func resourceRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*VSTSClient)
	name := d.Get("name").(string)
	projId := d.Get("project").(string)
	id := d.Id()

	log.Printf("[DEBUG] Going to delete repo %s (id: %s)", name, id)

	_, err := client.Delete(fmt.Sprintf("%s/_apis/git/repositories/%s", projId, id))

	if err != nil {
		log.Printf("[DEBUG] Could not delete repository %s: %s", id, err)
	}

	return err
}
