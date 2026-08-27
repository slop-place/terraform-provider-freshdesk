resource "freshdesk_solution_category" "getting_started" {
  name        = "Getting started"
  description = "First steps for new customers"
}

resource "freshdesk_solution_folder" "setup" {
  category_id = freshdesk_solution_category.getting_started.id
  name        = "Setup"
  visibility  = 1 # everyone
}

resource "freshdesk_solution_article" "install" {
  folder_id   = freshdesk_solution_folder.setup.id
  title       = "Installing the desktop client"
  description = "<h2>Download</h2><p>Grab the installer from your account page.</p>"
  status      = 2 # published
  tags        = ["setup", "desktop"]

  meta_title       = "Install the desktop client"
  meta_description = "How to download and install the desktop client."
}
