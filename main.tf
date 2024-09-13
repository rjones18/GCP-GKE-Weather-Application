resource "google_service_account" "default" {
  account_id   = "gke-service-account2"
  display_name = "gke-service-account2"
}

resource "google_project_iam_binding" "gcr_bucket_access" {
  project = "alert-flames-286515"  # Replace with your actual project ID
  role    = "roles/storage.admin"  # Grants full control over GCR; adjust if needed
  members = [
    "serviceAccount:${google_service_account.default.email}",
  ]
}

resource "google_container_cluster" "primary" {
  name     = "my-gke-cluster"
  location = "us-central1"

  # Reference existing VPC and Subnet
  network    = "projects/alert-flames-286515/global/networks/project-vpc"  # Replace with your existing VPC's self-link or name
  subnetwork = "projects/alert-flames-286515/regions/us-central1/subnetworks/app-subnet-1"  # Replace with your existing subnet's self-link or name

  remove_default_node_pool = true
  initial_node_count       = 1
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  name       = "my-node-pool"
  location   = "us-central1"
  cluster    = google_container_cluster.primary.name
  node_count = 1

  node_config {
    preemptible  = true
    machine_type = "e2-medium"

    service_account = google_service_account.default.email
    oauth_scopes    = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}
