package provider

import (
	"context"
	"fmt"
	"github.com/chrisamti/ceph-rest-client/ceph"
	"github.com/chrisamti/terraform-provider-ceph-rest/cephrest/configuration"
	"github.com/chrisamti/terraform-provider-ceph-rest/cephrest/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

//type Configuration struct {
//	Client *ceph.Client
//}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ceph_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_USER", nil),
				Description: "Username needed to login ceph rest api.",
			},
			"ceph_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_PASSWORD", nil),
				Description: "Password needed to login ceph rest api.",
			},
			"ceph_server": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_SERVER", nil),
				Description: "List of Ceph Server fqdn or IP address needed for login to ceph rest api.",
				Set:         schema.HashString,
			},
			"ceph_api_path": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_API_PATH", "api"),
				Description: "Ceph Server api path (default api).",
			},
			"ceph_port": {
				Type:        schema.TypeInt,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_PORT", 8443),
				Description: "Ceph Server TCP Port needed for login to ceph rest api.",
			},
			"ceph_http_protocol": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_HTTP_PROTOCOL", "https"),
				Description: "Ceph Server http protocol (http/https) to be used for login to ceph rest api.",
			},
			"ceph_insecure_skip_verify": {
				Type:        schema.TypeBool,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CEPH_INSECURE_SKIP_VERIFY", true),
				Description: "skip verify unknown certs.",
			},
		},
		ResourcesMap:         map[string]*schema.Resource{"ceph_rbd": service.ResourceRBD()},
		ConfigureContextFunc: providerConfigure,
	}
}

// type ConfigureContextFunc func(context.Context, *ResourceData) (interface{}, diag.Diagnostics)
func providerConfigure(_ context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var port = rd.Get("ceph_port").(int)
	var serverList []string

	if cephServerList, ok := rd.GetOk("ceph_server"); ok {
		for _, srv := range cephServerList.(*schema.Set).List() {
			serverList = append(serverList, srv.(string))
		}
	}

	// login
	user := rd.Get("ceph_user").(string)
	pass := rd.Get("ceph_password").(string)

	for _, cs := range serverList {

		server := ceph.Server{
			Address:            cs,
			Port:               uint(port),
			Protocol:           rd.Get("ceph_http_protocol").(string),
			APIPath:            rd.Get("ceph_api_path").(string),
			InsecureSkipVerify: rd.Get("ceph_insecure_skip_verify").(bool),
		}

		client, err := ceph.New(server)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "could not create ceph rest api client",
				Detail:        err.Error(),
				AttributePath: nil,
			})
		}

		statusLogin, errLogin := client.Session.Login(user, pass)

		if errLogin != nil {
			diags = append(diags, diag.FromErr(err)...)
		}

		switch statusLogin {
		case http.StatusCreated:
			// discard all previous errors and return configuration
			return &configuration.Ceph{Client: client}, diag.Diagnostics{}
		default:
			// append error to diags
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("could not login to rest api server '%s' with user '%s'", server.Address, user),
				Detail:        fmt.Sprintf("expected http status is 201 - got http status %d", statusLogin),
				AttributePath: nil,
			})
		}
	}

	// no success... return all errors
	return &configuration.Ceph{}, diags
}
