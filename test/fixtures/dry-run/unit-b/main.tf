# Simple terraform configuration for dry-run testing
# This file should NOT be executed when dry-run is enabled

resource "local_file" "test" {
  content  = "This file should not be created during dry-run"
  filename = "${path.module}/test-output.txt"
}

output "test_output" {
  value = "This output should not appear during dry-run"
}
