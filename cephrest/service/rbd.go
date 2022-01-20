package service

import (
	"context"
	"github.com/chrisamti/ceph-rest-client/ceph"
	"github.com/chrisamti/terraform-provider-ceph-rest/cephrest/configuration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func ResourceRBD() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRBDCreate,
		ReadContext:   resourceRBDRead,
		DeleteContext: resourceRBDelete,
		UpdateContext: resourceRBDUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pool_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"img_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_space": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "ceph name space",
			},
			"size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ceph rbd image size in bytes",
			},
		},
		// TODO: define TimeOuts

	}
}

func resourceRBDRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// warnings and errors
	var diags diag.Diagnostics

	cephConf := meta.(*configuration.Ceph)
	client := cephConf.Client

	imageSpec := ceph.PathJoin(
		d.Get("pool_name").(string),
		d.Get("name_space").(string),
		d.Get("img_name").(string),
	)

	_, rbd, err := client.GetBlockImage(imageSpec)

	if err != nil {
		return diag.FromErr(err)
	}

	// rbdMap := make([]map[string]interface{}, 0)

	if err = d.Set("pool_name", rbd.PoolName); err != nil {
		return diag.FromErr(err)
	}

	// set unique ID created by ceph
	d.SetId(rbd.UniqueID)

	if err = d.Set("pool_name", rbd.PoolName); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name_space", rbd.Namespace); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("img_name", rbd.Name); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("size", rbd.Size); err != nil {
		return diag.FromErr(err)
	}

	return diags

}

// type CreateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics
func resourceRBDCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// warnings and errors
	var (
		poolName  string
		imgName   string
		nameSpace *string
	)

	//// create timeout
	//ctxWithTimeOut, cancel := context.WithTimeout(ctx, time.Duration(300)*time.Second)

	//defer func() {
	//	log.Println("[DEBUG] ")
	//	cancel()
	//}()

	// ConfigureContextFunc
	cephConf := meta.(*configuration.Ceph)

	client := cephConf.Client

	poolName = d.Get("pool_name").(string)
	imgName = d.Get("img_name").(string)

	if d.Get("name_space").(string) != "" {
		*nameSpace = d.Get("name_space").(string)
	}

	rbd := ceph.RBDCreate{
		Features:      nil,
		PoolName:      poolName,
		Namespace:     nameSpace,
		Name:          imgName,
		Size:          d.Get("size").(int),
		ObjSize:       0,
		StripeUnit:    nil,
		StripeCount:   nil,
		DataPool:      nil,
		Configuration: struct{}{},
	}

	log.Printf("[DEBUG] creating rbd image %s %v %s %d", poolName, nameSpace, imgName, d.Get("size").(int))

	_, err := client.CreateBlockImage(rbd, 0)

	if err != nil {
		return diag.FromErr(err)
	}

	// try to read just created rbd image
	return resourceRBDRead(ctx, d, meta)

}

func resourceRBDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// ConfigureContextFunc
	cephConf := meta.(*configuration.Ceph)

	client := cephConf.Client

	var nameSpace *string

	if d.Get("name_space").(string) != "" {
		*nameSpace = d.Get("name_space").(string)
	}

	_, err := client.DeleteBlockImage(d.Get("pool_name").(string), nameSpace, d.Get("img_name").(string), 0)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags

}

func resourceRBDUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		nameSpace *string
		poolName  string
		imgName   string
	)

	// ConfigureContextFunc
	cephConf := meta.(*configuration.Ceph)

	client := cephConf.Client

	if d.Get("name_space").(string) != "" {
		*nameSpace = d.Get("name_space").(string)
	}

	poolName = d.Get("pool_name").(string)
	imgName = d.Get("img_name").(string)

	rbdUpdate := ceph.RBDUpdate{
		Features:      nil,
		Name:          imgName,
		Size:          d.Get("img_size").(int64),
		Configuration: struct{}{},
	}

	_, err := client.UpdateBlockImage(poolName, nameSpace, imgName, rbdUpdate, 0)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRBDRead(ctx, d, meta)
}
