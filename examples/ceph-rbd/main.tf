terraform {
  required_version = ">=0.12"

  required_providers {
    ceph = {
      source  = "localhost/chrisamti/ceph"
      version = "~> 0.0.1"
    }
  }
}

provider "ceph" {
  ceph_user     = "test-user"
  ceph_password = "XJEGy5yWrYxu758"
  ceph_server   = ["192.168.21.30", "192.168.21.31"]
  ceph_port     = 8443
}

resource "ceph_rbd" "ceph_rbd_test_1" {
  pool_name = "test-pool-1"
  # name_space = ""
  img_name  = "terraform-created-1"
  size      = 1073741824
}
