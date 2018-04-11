Terraform Provider for Visual Studio Team Services
==================================================

Provider configuration
----------------------

```
provider "vsts" {
    account  = "abcdefg"
    username = "abcdefg@example.com"
    token    = "XXXX"
}
```

`account` is turned into `https://abcdefg.visualstudio.com` and is the URL you use when using the web interface.

The token, account and username can either be set in the provider configuration (as shown above) or be set with environment variables:

```
export VSTS_TOKEN=XXXX
export VSTS_ACCOUNT=abcdefg
export VSTS_USER=abcdefg@example.com.com
```

Using environment variables will mean that you don't need to commit plain-text credential to your config repo.

You can generate the token via the webinterface at <https://abcedfg.visualstudio.com/_details/security/tokens>.


Resource definitions
--------------------

### Projects
```
resource "vsts_project" "testproj" {
     name    = "projectname"
     data    = <<JSON
        {
           "VersionControlOption": "Git",
           "ProjectVisibilityOption": null
        }
JSON
     source  = "NewProjectCreation"
     type    = "adcc42ab-9882-485e-a3ed-7678f01f66bc"
}
```

`name` is the name of your project.
`data` and `source` should probably be copied verbatim, unless you wish to override any of those.
`type` should be `ADCC42AB-9882-485E-A3ED-7678F01F66BC` for Agile, or `27450541-8E31-4150-9947-DC59F998FC01` for
CMMI, or `6B724908-EF14-45CF-84F8-768B5384DA45` for Scrum process templates.

Limitations
-----------

Resource importing is not available yet.

