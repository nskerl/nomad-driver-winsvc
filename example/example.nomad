job "example" {
  datacenters = ["dc1"]
  type        = "service"

  group "group" {
    count = 2
    task "task" {

      artifact {
        source      = "https://path.to.my.artifact.zip"
        destination = "local/example"
        options {
          archive = "zip"
        }
        headers {
          X-My-ApiKey = "API-1234"
        }
      }

      driver = "winsvc"

      config {
        executable = "local/example/example.exe"
        username = "domain\\example"
        password = "password"
        // services using Topshelf need to pass Nomad's instance name back in as an arg
        args = [
           "-servicename", "nomad.winsvc.${NOMAD_JOB_NAME}.${NOMAD_ALLOC_ID}",
           "-displayname", "nomad.winsvc.${NOMAD_JOB_NAME}.${NOMAD_ALLOC_ID}",
        ]
      }
    }
  }
}