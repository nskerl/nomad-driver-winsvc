data_dir  = "c:\\hashicorp\\nomad\\data"

plugin_dir = "C:\\dev\\github\\nomad-driver-winsvc"

log_file = "c:\\logs\\nomad\\nomad.log"
log_level = "DEBUG"

bind_addr = "0.0.0.0" # the default

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled       = true
}

plugin "winsvc" {
  config {
  }
}