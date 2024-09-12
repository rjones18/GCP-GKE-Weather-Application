resource "google_service_account" "default" {
  account_id   = "gke-service-account"
  display_name = "gke-service-account"
}

resource "google_project_iam_binding" "gcr_bucket_access" {
  project = "alert-flames-286515"  # Replace with your actual project ID
  role    = "roles/storage.admin"  # Grants full control over GCR; adjust if needed
  members = [
    "serviceAccount:${google_service_account.default.email}",
  ]
}

# Define the network (VPC) and subnetwork (subnet)
resource "google_compute_network" "vpc_network" {
  name = "project-vpc"
}

resource "google_compute_subnetwork" "vpc_subnetwork" {
  name          = "app-subnet-1"
  region        = "us-central1"
  network       = google_compute_network.vpc_network.id
  ip_cidr_range = "10.0.10.0/24"
}

resource "google_container_cluster" "primary" {
  name     = "my-gke-cluster"
  location = "us-central1"

  # Specify the VPC and Subnet
  network    = google_compute_network.vpc_network.self_link
  subnetwork = google_compute_subnetwork.vpc_subnetwork.self_link

  remove_default_node_pool = true
  initial_node_count       = 1
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  name       = "my-node-pool"
  location   = "us-central1"
  cluster    = google_container_cluster.primary.name
  node_count = 1

  # Specify the network and subnetwork for the node pool
  node_config {
    preemptible  = true
    machine_type = "e2-medium"

    service_account = google_service_account.default.email
    oauth_scopes    = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}
