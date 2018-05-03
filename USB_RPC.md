# USB gadget RPC mechanincs

- RPC server is in `service.go`
- RPC client with CLI is in `cli_client.go`

## RPC server
- holds state of desired USB setting with name `settingsState` and type `proto.GadgetSettings`
- `settingsState` is initialized to the hard-coded default settings when the RPC server is started (init) and deployed to the kernel via config FS
- the active USB gadget has not the same setting as the state stored in `settingsState`, this is only the case on init, when the default settings are deployed
- `settingsState` is meant to store the desired settings, which are customized by the user via RPC calls, in order to deploy the settings, an additional RPC call is needed (`deploySettings`)
    - the user can change a single or multiple USB gadget settings with a single RPC call from the cli_client
    - for this purpose, the RPC call `SetGadgetSettings` is used, which has to provide a full `proto.GadgetSettings` struct
    - the settings provided with `SetGadgetSettings` overwrite the settings stored in `settingsState` (RPC doesn't keep track of changed and unchanged settings)
    - thus, in order to change only a part of the settings, the client has to retrieve current `settingsState` via the RPC call `GetGadgetSettings` modify the result to its needs and send it back with `SetGadgetSettings`
    - once the client is done with changing settings (could be done with a single or multiple calls to `SetGadgetSettings`) he issues the RPC call `deploySettings` to deploy the settings (real active seetings and `settingsState` are the same again, if no error occures)
- `GetDeployedGadgetSetting` RPC call could be used to retrieve the real settings in use from the running gadget
 
## Error handling
- the RPC server checks the settings to be valid after every call of `SetGadgetSettings`, if an error occures it is reported back and the `settingsState`isn't updated
- if `deploySettings` is called, the server stores the old settings of the gadgets, in case of an error it is reported back and the active gadget configuration, as well as the `settingsState`, are reverted to the stored configuration
- before Deploying new gadget settings, they are compared to the active ones, to avoid re-building of the composite gadget if it isn't needed
- the most likely reason to end up with invalid gadget settings is that too many USB gadgets functions are enabled simultaneously, as each function consumes one or more USB endpoints (there're only 7 EPs available)  
 
## Additional notes
- the reason for not directly applying changed gadget settings, is that the whole composite gadget is destroyed and brought up again for every change
- there're likely parameters which could be changed without destroying the gadget (like the image backing an USB Mass Storage), these parameters are handled differently
- Re-deployment of the USB gagdget with changed settings interrupts the workflow, for example USB-based network interfaces are shut down during re-deployment (even if they aren't affected by the changed settings) - this would interrupt running connections or listening server sockets
