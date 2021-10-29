# nomad-driver-winsvc

A task driver plugin to orchestrate Windows Services as [Hashicorp Nomad](https://www.nomadproject.io/) job tasks. </br>


Requirements
------------
- [Nomad](https://www.nomadproject.io/downloads.html) v0.9+
- [Go](https://golang.org/doc/install) v1.11+ (to build the provider plugin)
- Windows Server 2016+


Configuring the Plugin
-------------------------
A task is configured using the below options:


|Option          |Type      |Description                                                    |
|:---------------|:--------:|:--------------------------------------------------------------|
| **executable** | `string` | Path to executable binary                                     |
| **args**       | `string` | Command line arguments to pass to executable                  |
| **username**   | `string` | Name of the account under which the service should run        |
| **password**   | `string` | Password of the account                                       |
