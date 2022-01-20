# Terraform provider plugin for ceph rest API
## Supported resources

### Debugging the provider

Debugging is available for this provider through the Terraform Plugin SDK versions 2.0.0. Therefore, the plugin can be
started with the debugging flag `--debug`.

For example (using [delve](https://github.com/go-delve/delve) as Debugger):

```bash
dlv --listen=:62630 --headless=true --api-version=2 --accept-multiclient exec .terraform/providers/localhost/chrisamti/ceph/0.0.1/darwin_amd64/terraform-provider-ceph_v0.0.1 -- --debug
API server listening at: [::]:62630
2022-01-20T08:41:45+01:00 warning layer=rpc Listening for remote connections (connections are not authenticated nor encrypted)
debugserver-@(#)PROGRAM:LLDB  PROJECT:lldb-1300.0.42.3
 for arm64.
Got a connection, launched process .terraform/providers/localhost/chrisamti/ceph/0.0.1/darwin_amd64/terraform-provider-ceph_v0.0.1 (pid = 60888).
{"@level":"debug","@message":"plugin address","@timestamp":"2022-01-20T08:42:00.771778+01:00","address":"/var/folders/k3/lydswhmd5jdbqlg9mfs0_97c0000gn/T/plugin748456223","network":"unix"}
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

	TF_REATTACH_PROVIDERS='{"localhost/chrisamti/ceph":{"Protocol":"grpc","ProtocolVersion":5,"Pid":60888,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/k3/lydswhmd5jdbqlg9mfs0_97c0000gn/T/plugin748456223"}}}'
```
Take the env var TF_REATTACH_PROVIDERS in front of your terraform in a second window:

```
TF_REATTACH_PROVIDERS='{"localhost/chrisamti/ceph":{"Protocol":"grpc","ProtocolVersion":5,"Pid":60888,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/k3/lydswhmd5jdbqlg9mfs0_97c0000gn/T/plugin748456223"}}}' terraform plan

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # ceph_rbd.ceph_rbd_test_1 will be created
  + resource "ceph_rbd" "ceph_rbd_test_1" {
      + id        = (known after apply)
      + img_name  = "terraform-created-1"
      + pool_name = "test-pool-1"
      + size      = 1073741824
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```
Happy Debugging 🐛

For more information about debugging a provider please
see: [Debugger-Based Debugging](https://www.terraform.io/docs/extend/debugging.html#debugger-based-debugging)

## Useful links

* [CEPH Documentation](https://docs.ceph.com/en/latest/)
* [CEPH RESTFUL Module](https://docs.ceph.com/en/latest/mgr/restful/)
* [CEPH RESTFUL API ](https://docs.ceph.com/en/latest/mgr/ceph_api/#)
* [Terraform documentation](https://www.terraform.io/docs/index.html)
* [Terraform example provider](https://github.com/hashicorp/terraform-provider-hashicups)