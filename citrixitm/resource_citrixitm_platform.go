package citrixitm

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/tolgaakyuz/citrix-go/itm"

	backoff "github.com/cenkalti/backoff/v4"
)

func resourceCitrixITMPlatform() *schema.Resource {
	return &schema.Resource{
		Create: resourceCitrixITMPlatformCreate,
		Read:   resourceCitrixITMPlatformRead,
		Update: resourceCitrixITMPlatformUpdate,
		Delete: resourceCitrixITMPlatformDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Delivery Networks",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"openmix_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"radar": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"probe_response_time_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"probe_availability_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"probe_throughput_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Required: false,
							Default:  10,
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCitrixITMPlatformCreate(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Println("[INFO][PLATFORM][CREATE] Start")

	// options for the API request
	options, err := prepareOptionsForPlatform(d)
	if err != nil {
		log.Printf("[DEBUG][PLATFORM][CREATE] Error in preparing options: %s", err)
		return err
	}

	log.Printf("[DEBUG][PLATFORM][CREATE] options:\n%+v", options)

	var platform *itm.Platform
	err = backoff.Retry(func() error {
		log.Printf("[INFO][PLATFORM][CREATE] Trying... time: %s\n", time.Now())
		platform, err = client.Platform.Create(options)
		return err
	}, NewExponentialBackOff())
	if err != nil {
		log.Printf("[DEBUG][PLATFORM][CREATE] API error: %s", err)
		return err
	}

	d.SetId(strconv.Itoa(platform.Id))
	log.Printf("[INFO][PLATFORM][CREATE] Success with ID %s", d.Id())

	return resourceCitrixITMPlatformRead(d, m)
}

func resourceCitrixITMPlatformRead(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Printf("[INFO][PLATFORM][READ] Start")

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][PLATFORM][READ] Converting app id (%s) to an integer: %s", d.Id(), err)
	}

	var platform *itm.Platform
	err = backoff.Retry(func() error {
		log.Printf("[INFO][PLATFORM][READ] Trying... time: %s\n", time.Now())
		platform, err = client.Platform.Get(id)
		return err
	}, NewExponentialBackOff())
	if err != nil {
		return fmt.Errorf("[ERROR][PLATFORM][READ] API Error for id: %s: %s", d.Id(), err)
	}

	log.Printf("[INFO][PLATFORM][READ] Success\n %+v", platform)

	return updateStateForPlatform(platform, d)
}

func resourceCitrixITMPlatformUpdate(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Printf("[INFO][PLATFORM][UPDATE] Start")

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][PLATFORM][UPDATE] Error converting app id (%s) to an integer: %s", d.Id(), err)
	}

	if d.HasChange("name") ||
		d.HasChange("alias") ||
		d.HasChange("description") ||
		d.HasChange("category") ||
		d.HasChange("radar") {
		options, err := prepareOptionsForPlatform(d)
		if err != nil {
			log.Printf("[DEBUG][PLATFORM][CREATE] Error in preparing options: %s", err)
			return err
		}

		log.Printf("[DEBUG][PLATFORM][UPDATE] options:\n%#v", options)

		var platform *itm.Platform
		err = backoff.Retry(func() error {
			log.Printf("[INFO][PLATFORM][UPDATE] Trying... time: %s\n", time.Now())
			platform, err = client.Platform.Update(id, options)
			return err
		}, NewExponentialBackOff())
		if err != nil {
			return fmt.Errorf("[WARN][PLATFORM][UPDATE] API error with ID %s: %s", d.Id(), err)
		}

		log.Printf("[INFO][PLATFORM][UPDATE] Success with ID %s: \n %+v", d.Id(), platform)
	} else {
		log.Printf("[DEBUG][PLATFORM][UPDATE] no change is detected for %s", d.Id())
	}

	return resourceCitrixITMPlatformRead(d, m)
}

func resourceCitrixITMPlatformDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*itm.Client)
	log.Printf("[INFO][PLATFORM][DELETE] Start with ID %s", d.Id())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][PLATFORM][DELETE] Converting app id (%s) to an integer: %s", d.Id(), err)
	}

	err = backoff.Retry(func() error {
		log.Printf("[INFO][PLATFORM][DELETE] Trying... time: %s\n", time.Now())
		return client.Platform.Delete(id)
	}, NewExponentialBackOff())
	if err != nil {
		return fmt.Errorf("[WARN][PLATFORM][DELETE] API error with ID %s: %s", d.Id(), err)
	}

	log.Printf("[INFO][PLATFORM][DELETE] Success with ID %s", id)

	return nil
}

// Helpers

func updateStateForPlatform(platform *itm.Platform, d *schema.ResourceData) error {
	d.Set("name", platform.DisplayName)
	d.Set("alias", platform.Name)
	d.Set("description", platform.Description)
	d.Set("enabled", true)
	d.Set("openmix_enabled", platform.OpenMixEnabled)
	d.Set("category", platform.Category["name"].(string))

	if err := d.Set("radar", flattenPlatformRadar(platform.RadarOpts)); err != nil {
		return err
	}

	return nil
}

func prepareOptionsForPlatform(d *schema.ResourceData) (*itm.PlatformOpts, error) {
	category_name := d.Get("category").(string)
	category_id := 0

	if category_name == "Delivery Networks" {
		category_id = 3
	}

	category := map[string]interface{}{
		"id":   category_id,
		"name": category_name,
	}

	radar, err := expandPlatformRadar(d.Get("radar").(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	name := d.Get("name").(string)
	alias := d.Get("alias").(string)

	if alias == "" {
		alias = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	}

	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	openMixEnabled := d.Get("openmix_enabled").(bool)

	return &itm.PlatformOpts{
		Name:                      alias,
		DisplayName:               name,
		Category:                  category,
		RadarOpts:                 radar,
		Description:               description,
		Enabled:                   enabled,
		OpenMixEnabled:            openMixEnabled,
		OpenmixVisible:            true,
		IsPrivate:                 true,
		PublicProviderArchetypeId: 0,
	}, nil
}

func NewExponentialBackOff() *backoff.ExponentialBackOff {
	exp := backoff.NewExponentialBackOff()
	exp.InitialInterval = 1 * time.Second
	exp.MaxElapsedTime = 20 * time.Second
	exp.Multiplier = 2
	exp.Reset()

	return exp
}

// Flatteners

func flattenPlatformRadar(radarOpt map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	m["probe_response_time_url"] = radarOpt["rttSecureUrl"]
	m["probe_availability_url"] = radarOpt["primeSecureUrl"]
	m["probe_throughput_url"] = radarOpt["xlSecureUrl"]
	m["weight"] = fmt.Sprintf("%d", int(radarOpt["weight"].(float64)))
	return m
}

// Expanders

func expandPlatformRadar(state map[string]interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 0)
	m["rttSecureUrl"] = state["probe_response_time_url"]
	m["primeSecureUrl"] = state["probe_availability_url"]
	m["xlSecureUrl"] = state["probe_throughput_url"]

	weight, err := strconv.Atoi(state["weight"].(string))
	if err != nil {
		return nil, fmt.Errorf("[ERROR][PLATFORM][EXPAND] Converting id (%s) to an integer: %s", state["weight"], err)
	}
	m["weight"] = weight
	return m, nil
}
