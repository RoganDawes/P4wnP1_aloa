# Common functionality used by "service" and "cli_client"

The folder hold consts, data and functions, which are used by both, the cli_client and the service.
It should not hold functionality used by the webclient, in order to avoid using huge packages like "fmt".
Functionality shared with the web_client has to be stored in the folder "common_web".
  