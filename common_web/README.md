# Common functionality used by "service" and "cli_client"

The folder hold consts, data and functions, which are used by the cli_client, the service AND/OR the web_service.
Using huge packages like "fmt" should not be used by functionality in this folder, in order to avoid importing it into the webclient.
  