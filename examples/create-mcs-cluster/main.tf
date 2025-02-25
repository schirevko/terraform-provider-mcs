terraform {
  required_providers {
    mcs = {
      source = "MailRuCloudSolutions/mcs"
      version = "~> 0.5.8"
    }
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

resource "openstack_networking_network_v2" "k8s" {
  name           = "k8s-net"
  admin_state_up = true
}

resource "openstack_networking_subnet_v2" "k8s-subnetwork" {
  name            = "k8s-subnet"
  network_id      = openstack_networking_network_v2.k8s.id
  cidr            = "10.100.0.0/16"
  ip_version      = 4
  dns_nameservers = ["8.8.8.8", "8.8.4.4"]
}

data "openstack_networking_network_v2" "extnet" {
  name = "ext-net"
}

resource "openstack_networking_router_v2" "k8s" {
  name                = "k8s-router"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.extnet.id
}

resource "openstack_networking_router_interface_v2" "k8s" {
  router_id = openstack_networking_router_v2.k8s.id
  subnet_id = openstack_networking_subnet_v2.k8s-subnetwork.id
}

resource "openstack_compute_keypair_v2" "keypair" {
  name       = "default"
  public_key = file(var.public-key-file)
}

data "openstack_compute_flavor_v2" "k8s" {
  name = var.k8s-flavor
}

data "mcs_kubernetes_clustertemplate" "ct" {
  version = "1.21.4"
}

resource "mcs_kubernetes_cluster" "k8s-cluster" {
  depends_on = [
    openstack_networking_router_interface_v2.k8s,
  ]

  name = "k8s-cluster"
  cluster_template_id = data.mcs_kubernetes_clustertemplate.ct.id
  master_flavor       = data.openstack_compute_flavor_v2.k8s.id
  master_count        = 1

  keypair = openstack_compute_keypair_v2.keypair.id
  network_id = openstack_networking_network_v2.k8s.id
  subnet_id = openstack_networking_subnet_v2.k8s-subnetwork.id
  floating_ip_enabled = true
  availability_zone = "MS1"
}
