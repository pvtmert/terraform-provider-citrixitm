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

func resourceCitrixITMDnsApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceCitrixITMDnsAppCreate,
		Read:   resourceCitrixITMDnsAppRead,
		Update: resourceCitrixITMDnsAppUpdate,
		Delete: resourceCitrixITMDnsAppDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "V1_JS", // ['V2_PHP' or ' V1_JS' or ' STATIC_ROUTING' or ' RT_HTTP_PERFORMANCE' or ' RR_PURE_WEIGHTED' or ' STATIC_FAILOVER']
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "http", // ['dns' or ' http']
			},
			"app_data": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: resourceCitrixITMDnsAppDiffSuppress,
			},
			"fallback_cname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fallback_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"cname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCitrixITMDnsAppCreate(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Println("[INFO][DNS_APP][CREATE] Start")

	options, err := prepareOptionsForDNSApp(d)
	if err != nil {
		log.Printf("[DEBUG][DNS_APP][CREATE] Error in preparing options: %s", err)
		return err
	}

	log.Printf("[DEBUG][DNS_APP][CREATE] options:\n%+v", options)

	var app *itm.DNSApp
	err = backoff.Retry(func() error {
		log.Printf("[INFO][DNS_APP][CREATE] Trying... time: %s\n", time.Now())
		app, err = client.DNSApps.Create(options, true)
		return err
	}, NewExponentialBackOff())
	if err != nil {
		log.Printf("[DEBUG][DNS_APP][CREATE] API error: %s", err)
		return err
	}

	d.SetId(strconv.Itoa(app.Id))
	log.Printf("[INFO][DNS_APP][CREATE] Success with ID %s", d.Id())

	return resourceCitrixITMDnsAppRead(d, m)
}

func resourceCitrixITMDnsAppRead(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Printf("[INFO][DNS_APP][READ] Start")

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][DNS_APP][READ] Converting app id (%s) to an integer: %s", d.Id(), err)
	}

	var app *itm.DNSApp
	err = backoff.Retry(func() error {
		log.Printf("[INFO][DNS_APP][READ] Trying... time: %s\n", time.Now())
		app, err = client.DNSApps.Get(id)
		return err
	}, NewExponentialBackOff())
	if err != nil {
		return fmt.Errorf("[ERROR][DNS_APP][READ] API Error for id: %s: %s", d.Id(), err)
	}

	log.Printf("[INFO][DNS_APP][READ] Success\n %+v", app)

	return updateStateForDNSAPP(app, d)
}

func resourceCitrixITMDnsAppUpdate(d *schema.ResourceData, m interface{}) error {
	var err error
	client := m.(*itm.Client)
	log.Printf("[INFO][DNS_APP][UPDATE] Start")

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][DNS_APP][UPDATE] Error converting app id (%s) to an integer: %s", d.Id(), err)
	}

	if d.HasChange("name") ||
		d.HasChange("description") ||
		d.HasChange("fallback_cname") ||
		d.HasChange("fallback_ttl") ||
		d.HasChange("app_data") {

		options, err := prepareOptionsForDNSApp(d)
		if err != nil {
			log.Printf("[DEBUG][DNS_APP][CREATE] Error in preparing options: %s", err)
			return err
		}

		log.Printf("[DEBUG][DNS_APP][UPDATE] options:\n%#v", options)

		var app *itm.DNSApp
		err = backoff.Retry(func() error {
			log.Printf("[INFO][DNS_APP][UPDATE] Trying... time: %s\n", time.Now())
			app, err = client.DNSApps.Update(id, options, true)
			return err
		}, NewExponentialBackOff())
		if err != nil {
			return fmt.Errorf("[WARN][DNS_APP][UPDATE] API error with ID %s: %s", d.Id(), err)
		}

		log.Printf("[INFO][DNS_APP][UPDATE] Success with ID %s: \n %+v", d.Id(), app)
	} else {
		log.Printf("[DEBUG][DNS_APP][UPDATE] no change is detected for %s", d.Id())
	}

	return resourceCitrixITMDnsAppRead(d, m)
}

func resourceCitrixITMDnsAppDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*itm.Client)
	log.Printf("[INFO][DNS_APP][DELETE] Start with ID %s", d.Id())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("[ERROR][DNS_APP][DELETE] Converting app id (%s) to an integer: %s", d.Id(), err)
	}

	err = backoff.Retry(func() error {
		log.Printf("[INFO][DNS_APP][DELETE] Trying... time: %s\n", time.Now())
		return client.DNSApps.Delete(id)
	}, NewExponentialBackOff())
	if err != nil {
		return fmt.Errorf("[WARN][DNS_APP][DELETE] API error with ID %s: %s", d.Id(), err)
	}

	log.Printf("[INFO][DNS_APP][DELETE] Success with ID %s", id)

	return nil
}

// Helpers

func resourceCitrixITMDnsAppDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

func updateStateForDNSAPP(app *itm.DNSApp, d *schema.ResourceData) error {
	d.Set("name", app.Name)
	d.Set("description", app.Description)
	d.Set("type", app.Type)
	d.Set("protocol", app.Protocol)
	d.Set("app_data", app.AppData)
	d.Set("fallback_cname", app.FallbackCname)
	d.Set("fallback_ttl", app.FallbackTtl)
	d.Set("cname", app.AppCname)
	d.Set("version", app.Version)

	return nil
}

func prepareOptionsForDNSApp(d *schema.ResourceData) (*itm.DNSAppOpts, error) {
	return &itm.DNSAppOpts{
		Name:          d.Get("name").(string),
		AppData:       d.Get("app_data").(string),
		Description:   d.Get("description").(string),
		FallbackCname: d.Get("fallback_cname").(string),
		Platforms:     []map[string]interface{}{},
		Type:          d.Get("type").(string),
		Protocol:      d.Get("protocol").(string),
		AvlThreshold:  10,
	}, nil
}
