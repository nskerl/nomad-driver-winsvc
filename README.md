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


|Option          |Type        |Description                                                    |
|:---------------|:----------:|:--------------------------------------------------------------|
| **executable** | `string`   | Path to executable binary                                     |
| **args**       | `[]string` | Command line arguments to pass to executable                  |
| **username**   | `string`   | Name of the account under which the service should run        |
| **password**   | `string`   | Password of the account                                       |

Example Job
-----------

```console
PS> nomad job run example/example.nomad
```

When run, the above example job will result in the creation of (2) services each named to reflect their respective `JobName` & `AllocID` values:

```console
PS❯ get-service nomad.winsvc.* | select Status, Name, BinaryPathName

Status  Name                                                        BinaryPathName
------  ----                                                        --------------
Running nomad.winsvc.example.e402ad0a-7d7d-e72e-1823-64ba5ded3711   c:\hashicorp\nomad\data\alloc\e402ad0a-7d7d-e72e-1823-64ba5ded3711\task\local\example.exe
Running nomad.winsvc.example.c59a986f-6e14-f14f-d1ac-593cd86025b7   c:\hashicorp\nomad\data\alloc\c59a986f-6e14-f14f-d1ac-593cd86025b7\task\local\example.exe
```

Stopping the job will result in the graceful _stop_ and eventual removal of the services as the Nomad job completes and gc is run.

TODO
----
1. Support running as LOCALSYSTEM (perhaps just as default when username is nil)
2. Document how to pass password through environment variables rather than exposed in job file
3. Document how to support [TopShelf](https://github.com/Topshelf/Topshelf) services
