data "archive_file" "main" {
  type        = "zip"
  output_path = "artifacts/deploy_bundle.zip"
  source_dir  = "src"
}

resource "google_storage_bucket" "bucket" {
  name          = var.bucket_name
  force_destroy = var.bucket_force_destroy
  location      = var.region
  project       = var.project_id
  storage_class = "REGIONAL"
}

resource "google_storage_bucket_object" "archive" {
  name                = "${data.archive_file.main.output_md5}-${basename(data.archive_file.main.output_path)}"
  bucket              = google_storage_bucket.bucket.name
  source              = data.archive_file.main.output_path
  content_disposition = "attachment"
  content_encoding    = "gzip"
  content_type        = "application/zip"
}

resource "google_cloudfunctions_function" "function" {
  name        = "products"
  description = "Aggregates product price and name"
  runtime     = "go113"

  available_memory_mb   = 128
  trigger_http          = true
  entry_point           = "StartCloudFunction"
  project               = var.project_id
  region                = var.region
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.archive.name

  environment_variables = {
    PROJECT_ID = var.project_id
    DATASTORE_ID = var.datastore_id
  }
}

# IAM entry for all users to invoke the function
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

# module "datastore" {
#   source      = "terraform-google-modules/cloud-datastore/google"
#   credentials = "sa-key.json"
#   project     = var.project_id
#   indexes     = "${file("index.yaml")}"
# }