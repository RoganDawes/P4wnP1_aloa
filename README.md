# P4wnP1 A.L.O.A.

P4wnP1 A.L.O.A. by MaMe82 is a framework which turns a Rapsberry Pi Zero W into a flexible, low-cost platform for 
pentesting, red teaming and physical engagements ... or into "A Little Offensive Appliance".

## 1. Features

### Plug&Play USB device emulation
- USB functions:
  - USB Ethernet (RNDIS and CDC ECM)
  - USB Serial
  - USB Mass Storage (Flashdrive or CD-Rom)
  - HID Keyboard
  - HID Mouse
- runtime reconfiguration of USB stack (no reboot)
- detection of connect/disconnect makes it possible to keep P4wnP1 A.L.O.A powered up (external supply) and trigger 
action if the emulated USB device is attached to a new host
- no need to deal with different internal ethernet interfaces, as CDC ECM and RNDIS are connected to a virtual bridge
- persistent store and load of configuration templates for USB settings
 
### HIDScript
- replacement for limited DuckyScript
- sophisticated scripting language to automate keyboard and **mouse**
- up to 8 HIDScript jobs could run in parallel (keep a job up to jiggle the mouse, while others are started on-demand to
do arbitrary mouse and keystroke injection seamlessly)
- HIDScript is based on JavaScript, with common libraries available, which allows more complex scripts (function calls,
using `Math` for mouse calculations etc.)
- keyboard
  - based on UTF-8, so there's no limitation to ASCII characters
  - could react on feedback from the hosts real keyboard by reading back LED state changes of NUMLOCK, CAPSLOCK and 
  SCROLLLOCK (if the target OS shares LED state across all connected keyboards, which isn't the case for OSX)
  - take branching decisions in HIDScript, based on LED feedback
- mouse
  - relative movement (fast, but not precise)
  - stepped relative movement (slower, but accurate ... moves mouse in 1 DPI steps) 
  - **absolute positioning** on Windows (pixel perfect if target's screen dimensions are known)
- Keyboard and mouse are not only controlled by the same scripting language, both could be used in the same script. This
allows combining them in order to achieve goals, which couldn't be achieved using only keyboard or mouse.
  
### Bluetooth
- full interface to Bluez stack (currently no support for remote device discovery/connect)
- allows to run a Bluetooth Network Access Point (NAP)
- customizable Pairing (PIN based legacy mode or SSP)
- High Speed support (uses 802.11 frames to achieve WiFi like transfer rates)
- Runtime reconfiguration of the Bluetooth stack
- Note: PANU is possible, too, but currently not supported (no remote device connection)
- persistent store and load of configuration templates for Bluetooth settings
  
### WiFi
- modified Firmware (build with Nexmon framework)
  - allows KARMA (spoof valid answers for Access Points probed by remote devices and allow association)
  - broadcast additional Beacons, to emulate multiple SSIDs
  - WiFi covert channel
  - Note: Nexmon legacy monitor mode is included, but not supported by P4wnP1. Monitor mode is still buggy and likely to
  crash the firmware if the configuration changes. 
- easy Access Point configuration
- easy Station mode configuration (connect to existing AP)
- failover mode (if it is not possible to connect to the target Access Point, bring up an own Access Point)
- runtime reconfiguration of WiFi stack 
- persistent store and load of configuration templates for WiFi settings

### Networking
- easy ethernet interface configuration for
  - bluetooth NAP interface
  - USB interface (if RNDIS/CDC ECM is enabled)
  - WiFi interface
- supports dedicated DHCP server per interface
- support for DHCP client mode
- manual configuration
- persistent store and load of configuration templates for each interface

### Tooling
Not much to say here, P4wnP1 A.L.O.A. is backed by KALI Linux, so everything should be right at your hands (or could be 
installed using apt)

### Configuration and Control via CLI, remotely if needed
- all features mentioned so far, could be configured using a CLI client
- the P4wnP1 core service is a single binary, running as systemd unit which preserves runtime state
- the CLI client interfaces with this service via RPC (gRPC to be specific) to change the state of the core
- as the CLI uses a RPC approach, it could be used for **remote configuration**, too
- if P4wnP1 is accessed via SSH, the CLI client is there, waiting for your commands (or your tab completion kung fu) 
- the CLI is written in Go (as most of the code) and thus **compiles for most major platforms and architectures**

So if you want to use a a batch file running on a remote Windows host to configure P4wnP1 ... no problem:
1) compile the client for windows
2) make sure you could connect to P4wnP1 somehow (Bluetooth, WiFi, USB)
3) add the `host` parameter to your client commands
4) ... and use the CLI as you would do with local access. 

### Configuration and Control via web client

Although it wasn't planned initially, P4wnP1 A.L.O.A. could be configured using a webclient.
Even though the client wasn't planned, it evolved to a nice piece of software. In fact it ended up as the main 
configuration tool for P4wnP1 A.L.O.A.
The webclient has capabilities, which couldn't be accessed from the CLI (templates storage, creation of 
"TriggerActions").

The core features:
- should work on most major mobile and desktop browsers, with consistent look and feel (Quasar Framework)
- uses gRPC via websockets (no RESTful API, no XHR, nearly same approach as CLI)
- Thanks to this interface, the weblient does not rely on a request&reply scheme only, but receives "push events" from 
the P4wnP1 core. This means:
  - if you (or a script) changes the state of P4wnP1 A.L.O.A. these changes will be immediately reflected into the 
  webclient
  - if you have multiple webclients running, changes of the core's state will be reflected from one client to all other
  clients
- includes a HIDScript editor, with 
  - syntax highlighting
  - auto-complete (`CTRL+SPACE`)
  - persistent storage & load for HIDScripts 
  - **on-demand execution** of HIDScript directly from the browser
  - a HIDScript job manager (cancel running jobs, inspect job state and results)
- includes an overview and editor for TriggerActions
- full templating support for all features described so far
- the WebClient is a Single Page Application, once loaded everything runs client side, only gRPC request are exchanged

### Automation
The automation approach of the old P4wnP1 version (static bash scripts) couldn't be used anymore.

The automation approach of P4wnP1 A.L.O.A. had to fulfills these requirements:
- easy to use and understand
- usable from a webclient
- be generic and flexible, at the same time
- everything doable with the old "bash script" approach, should still be possible
- able to access all subsystems (USB, WiFi, Bluetooth, Ethernet Interfaces, HIDScript ... )
- modular, with reusable parts
- ability to support (simple) logical tasks without writing additional code
- **allow physical computing, by utilizing of the GPIO ports**

With introducing of the so called "TriggerActions" and by combining them with the templating system (persistent settings
storage for all sub systems) all the requirements could be satisfied. Details on TriggerActions could be find in the 
WorkFlow section.

## 2. Workflow part 1 - HIDScript

P4wnP1 A.L.O.A. has no static configuration (or payloads). In fact it has no static workflow at all.
 
P4wnP1 A.L.O.A. is meant to be as flexible as possible, to allow using it in all possible scenarios (including the ones
I couldn't think of while creating P4wnP1 A.L.O.A.).

But there are some basic concepts, I'd like to walk through in this section. As it is hard to explain everything without
creating a proper (video) documentation, I visit some some common use cases and examples in order to explain what needs
to be explained.

Nevertheless, it is unlikely that I'll have the time to provide a full-fledged documentation. **So I encourage everyone
to support me with tutorials and ideas, which could be linked back into this README**

Now let's start with one of the most basic tasks:

### 2.1 Run a keystroke injection against a host, which has P4wnP1 attached via USB

The minimum configuration requirements to achieve this goal are:
- The USB sub system is configured to emulate at least a keyboard
- There is a way to access P4wnP1 (remotely), in order to initiate the keystroke injection

The default configuration of P4wnP1's (unmodified image) meets these requirements already:
- the USB settings are initialized to provide **keyboard**, mouse and ethernet (both, RNDIS and CDC ECM) 
- P4wnP1 could already be accessed using one of the following methods:
	- WiFi
	  - the Access Point name should be obvious
	  - the password is `MaMe82-P4wnP1`
	  - the IP of P4wnP1 is `172.24.0.1`
	- USB Ethernet
	  - the IP of P4wnP1 is `172.16.0.1`
	- Bluetooth
	   - device name `P4wnP1`
	   - PIN `1337`
	   - the IP is `172.26.0.1`
       - Note: Secure Simple Pairing is OFF in order to force PIN Pairing. This again means, high speed mode is turned 
       off, too. So the bluetooth connection is very slow, which is less of a problem for SSH access, but requesting the
       webclient could take up to 10 minutes (in contrast to some seconds with high speed enabled).
- a SSH server is accessible from all the aforementioned IPs
- The SSH user for KALI Linux is `root`, the default password is `toor`
- The webclient could be reached over all three connections on port 8000 via HTTP

*Note:
Deploying a HTTPS connection is currently not in scope of the project. So please keep this in mind, if you use the 
webclient with sensitive data (like WiFi credentials). The whole project isn't built with security in mind (and it is 
unlikely that this will ever get a requirement). So keep in mind to deploy appropriate measures (f.e. restricting access
to webclient with iptables, if the Access Point is configured with Open Authentication; don't keep Bluetooth 
Discoverability and Connectability enabled without PIN protection etc. etc.)*

At this point I assume:
1) You have attached P4wnP1 to some target host via USB (the innermost of the Raspberry's micro USB ports is the one to 
use)
2) The USB host has an application running, which is able to receive the keystrokes and has keyboard input focus (f.e. 
a text editor)
3) You are remotely connected to P4wnP1 via SSH (the best way should be WiFi), preferably not from the same host which 
has the USB connection attached 

In order to run the CLI client from the SSH session, issue the following command:
```
root@kali:~# P4wnP1_cli 
The CLI client tool could be used to configure P4wnP1 A.L.O.A.
from the command line. The tool relies on RPC so it could be used 
remotely.

Version: v0.1.0-alpha1

Usage:
  P4wnP1_cli [command]

Available Commands:
  db          Database backup and restore
  evt         Receive P4wnP1 service events
  help        Help about any command
  hid         Use keyboard or mouse functionality
  led         Set or Get LED state of P4wnP1
  net         Configure Network settings of ethernet interfaces (including USB ethernet if enabled)
  system      system commands
  template    Deploy and list templates
  trigger     Fire a group send action or wait for a group receive trigger
  usb         USB gadget settings
  wifi        Configure WiFi (spawn Access Point or join WiFi networks)

Flags:
  -h, --help          help for P4wnP1_cli
      --host string   The host with the listening P4wnP1 RPC server (default "localhost")
      --port string   The port on which the P4wnP1 RPC server is listening (default "50051")

Use "P4wnP1_cli [command] --help" for more information about a command.
```


The resulting usage help screen shows, that the CLI uses different commands to interact with various subsystems of 
P4wnP1 A.L.O.A. Most of these commands have own sub-commands, again. The help for each command or sub-command could be 
accessed by appending `-h`:

```
root@kali:~# P4wnP1_cli hid run -h
Run script provided from standard input, commandline parameter or by path to script file on P4wnP1

Usage:
  P4wnP1_cli hid run [flags]

Flags:
  -c, --commands string      HIDScript commands to run, given as string
  -h, --help                 help for run
  -r, --server-path string   Load HIDScript from given path on P4wnP1 server
  -t, --timeout uint32       Interrupt HIDScript after this timeout (seconds)

Global Flags:
      --host string   The host with the listening P4wnP1 RPC server (default "localhost")
      --port string   The port on which the P4wnP1 RPC server is listening (default "50051")
```

Now, in order to type out "Hello world" to the USB host, the following CLI command could be used:

`P4wnP1_cli hid run -c 'type("Hello world")'`

The result output in the SSH session should look similar to this:

```
TempFile created: /tmp/HIDscript295065725
Start appending to 'HIDscript295065725' in folder 'TMP'
Result:
null
```

On the USB host "Hello World" should have been typed to the application with keyboard focus.

*If your SSH client runs on the USB host, the "Hello world" is typed somewhere into the result output of the CLI command
(it doesn't belong to the output, but has been typed in between).*

Goal achieved. We injected keystrokes to the target. Much reading for a simple task, but again, this section is meant to
explain basic concepts.

### 2.2 Moving on to more sophisticated language features of HIDScript

If you managed to run the "Hello world" keystroke injection, this is a good point to explore some additional HIDScript
features. We already know the `type` command, here are some more: 

#### Pressing special keys and combinations

The `type` command supports pressing return, by encoding a "new line" character into the input string, like this:
```
P4wnP1_cli hid run -c 'type("line 1\nline 2\nline 3 followed by pressing RETURN three times\n\n\n")'
```

But what about special keys or key combinations ? The `press` command comes to help!

Pressing CTRL+ALT+DELETE

```
P4wnP1_cli hid run -c 'press("CTRL ALT DELETE")'
```

*Note: all 3 keys have been modifiers in the last example*

Press the non-modifier key A:

```
P4wnP1_cli hid run -c 'press("A")'
```

*Note: The resulting output should be a lowercase 'a', because `press` interprets 'A' as key. In contrast, the `type`
command would try to press a key combination, which results in an uppercase 'A' output character.*

Combine a modifier and a non-modifier key, to get an uppercase 'A' as resulting character:

```
P4wnP1_cli hid run -c 'press("SHIFT A")'
```

Combine `press` and `type`.

```
P4wnP1_cli hid run -c 'type("before caps\n"); press("CAPS"); type("after caps\n"); press("CAPS");'
```
 
The last command typed a string, toggled CAPSLOCK, typed another string and toggled CAPS lock again. 
In result, CAPSLOCK should be in its initial state (toggled two times), but one of the strings is typed uppercase, the
other lowercase although both strings have been given in lower case.

Additional notes on key presses with `press`: 

I don't want to dive into the depth of USB keyboard reports inner workings, but some things are worth mentioning to 
pinpoint the limits and possibilities of the `press` command (which is based on raw keyboard reports):
- a keyboard report can contain up to 8 modifier keys at once
- the modifier keys are: LEFT_CTRL, RIGHT_CTRL, LEFT_ALT, RIGHT_ALT, LEFT_SHIFT, RIGHT_SHIFT, LEFT_GUI, RIGHT_GUI
- P4wnP1 allows aliases for common modifiers
  - CTRL == CONTROL == LEFT_CTRL
  - ALT == LEFT_ALT
  - SHIFT == LEFT_SHIFT
  - WIN == GUI == LEFT_GUI
- in addition to the modifiers, up to eight keys could be added
  - normal keys represent characters and special keys
  - example of special keys: BACKSPACE, ENTER (== RETURN), F1 .. F12)
  - `/usr/local/P4wnP1/keymaps/common.json` holds a formatted JSON keymap with all possible keys (be careful not to 
  change the file) 
- adding multiple keys to the press command, doesn't result in a key sequence, but in pressing all the given keys at the
same time
- `press` releases keys automatically, this means a sequence like "hold ALT, press TAB, press TAB, release ALT" 
currently isn't possible 

#### Keyboard layout

The HIDScript command to change the keyboard layout is `layout(language map string)`.

Here's an example on how to switch keyboard layout, multiple times in a single script:

```
P4wnP1_cli hid run -c 'layout("us"); type("Typing with EN_US layout\n");layout("de"); type("Typing with German layout supporting special chars üäö\n");'
```
 
The output result of the command given above, depends on the layout used by the USB host. 

On a host with German keyboard layout the result looks like this:
```
Tzping with EN?US lazout
Typing with German layout supporting special chars üäö
```
On a host with US keyboard layout it looks like this:
```
Typing with EN_US layout
Tzping with German lazout supporting special chars [';
```

Please note, that the intended output is only achieved, if P4wnP1's keyboard layout aligns with the one of the USB host. 
To achieve, this is the obvious aim of the `layout` command. 

Being able to change the layout in the middle of a running HIDScript, could come in handy: Who knows, maybe you like to 
brute force the target host's keyboard layout by issuing commands with changing layouts till they achieve the desired 
effect.

**Important:** The layout has global effect. This means if multiple HIDScripts are running concurrently and one of the 
scripts sets a new layout, all other scripts are effected immediately, too.

#### Typing speed

By default P4wnP1 injects keystrokes as fast as possible. Depending on your goal, this could be a bit too much (think of 
detecting keystroke injection attacks based on typing speed). HIDScript supports a command to change this behavior.

`typingSpeed(delayMillis, jitterMillis)`

The first argument is a delay in milliseconds, which is applied before each single keystroke. The second argument is an 
additional jitter in milliseconds. A random delay between 0 milliseconds and jitter is applied additionally.

Let's try it to type slower:

```
P4wnP1_cli hid run -c 'typingSpeed(100,0); type("Hello world")'
```

Instead of a constant delay, try a random jitter:
```
P4wnP1_cli hid run -c 'typingSpeed(0,500); type("Writing with random jitter up to 500 milliseconds")'
```

Finally, by combining and tuning both values, we could simulate natural typing speed:
```
P4wnP1_cli hid run -c 'typingSpeed(100,150); type("Writing with more natural speed")'
```

**Important:** The typing speed has global effect. This means if multiple HIDScripts are running concurrently and one of
the scripts sets a new typing speed, all other scripts are effected immediately, too.

#### Wait for LED report

This HIDScript feature needs a bit of explanation.

You may have noticed that (depending on the host OS) the keyboard state modifiers across multiple connected keyboards.
For example, if you connect two keyboards to a Windows host, and toggle CAPSLOCK on one of the, the CAPSLOCK LED changes
on both of them. 

If the state modifier state is shared across multiple keyboards on a given OS, could be tested exactly like this.

In case a USB host supports this kind of state sharing, P4wnP1's HIDScript language could make use out of it.

Imagine the following scenario:

P4wnP1 is connected to a USB host and you want to apply keystroke injection, but you don't want the HIDScript to 
run the keystrokes immediately. Instead the HIDScript should sit and wait till you hit NUMLOCK, CAPSLOCK or SCROLLLOCK
on the host's real keyboard. Why? Maybe you're involved in an engagement, somebody walked in and you don't want that
this somebody could see how a ton of characters magically are typed into a console window which suddenly popped up.
So you wait till somebody walks out, hit NUMLOCK and ultimately a console window pops up and a ton of character 
magically are ... yes you got it. It could be done like this:

```
P4wnP1_cli hid run -c 'waitLED(NUM); type("A ton of characters\n")'
```

If you tested the command above, you might encounter cases where the keystrokes immediately are issued, even if NUMLOCK
wasn't pressed (and the LED didn't toggle), after starting the script.

This is intended behavior and there's a good reason for it. Maybe you have used other keyboard scripting languages,
before which target USB devices capable of injecting keystrokes. Most of them have a common problem:

You don't know when to start typing! If you type immediately after the USB device is powered, it is likely that the USB
host didn't finished device enumeration and hasn't managed to bring up the keyboard driver. So your keystrokes are lost.
To overcome this you could add a delay, before keystroke injection starts. But how long should this delay be? Five 
seconds, 10 seconds, 30 seconds ? The answer is: it depends! It depends on how fast the host is able to enumerate the 
device and bring up the driver. In fact you couldn't know how long this takes, without testing against the actual 
target.

But as we have already learned, Operating Systems like Windows share the LED state across multiple keyboards.
This means if the NUMLOCK LED of the host keyboard is ON and you attach a second keyboard, the NUMLOCK LED on this 
keyboard has to be set to ON, too. If the NUMLOCK LED would have been OFF, anyways, the newly attached keyboard receives
the LED state anyways (all LEDs off in this case). The interesting thing about this, is that this "LED update" could
only happen, if the keyboard driver of the USB host has finished loading.

Isn't that beautiful, the USB host tells us: "I'm ready to receive keystrokes". There is no need to play around with 
initial delays. 

But here's the problem: We connect P4wnP1 to an USB host. We run a HIDScript starting with `waitLED` instead of a delay
and start typing afterwards, but nothing happens. Why? It is likely that we missed the LED state update, because it
arrived before we started the HIDScript at all. Exactly this is the reason, why P4wnP1 preserves all recognized LED 
state changes, unless at least one running HIDScript consumes them by calling `waitLED` (or `waitLEDRepeat`).

It is worth mentioning, that `waitLED` returns ONLY if the received LED state differs from P4wnP1's internal state.
This means, even if we listen for a change on any LED with `waitLED(ANY)` it still could happen, that we receive an 
initial LED state from a USB host after attaching P4wnP1 to it, which doesn't differ from P4wnP1's internal state.
In this case `waitLED(ANY)` would block forever (or till a real LED change happens).

This special case could be handled by calling `waitLED(ANY_OR_NONE)`, which returns as soon as a new LED state arrive,
even if it doesn't result in a change.

Enough explanation, let's get practical ... before we do so, we have to change the hardware setup a bit:

Attach an external power supply to the second USB port of the Raspberry Pi Zero (the outer one). This assures that
P4wnP1 doesn't loose power when detached from the USB host, as it doesn't rely on bus power anymore. The USB port which
should be used to connect P4wnP1 to the target USB host is the inner most of the two ports.

Now start the following HIDScript

``` 
P4wnP1_cli hid run -c 'while (true) {waitLED(ANY);type("Attached\n");}'
``` 

Detach P4wnP1 from the USB host (and make sure it is kept powered on)! Reattach it to the USB host ...
Everytime you reattach P4wnP1 to the host, "Attached" should be typed out.

This teaches us 3 things:
1) `waitLED` could be used as initial command in scripts, to start typing as soon as the keyboard driver is ready
2) `waitLED` isn't the perfect choice, to pause HID scripts until a LED changing key is pressed on the USB host, as 
preserved state changes could unblock the command in an unintended way
3) Providing more complex HIDScript as parameter to the CLI isn't very convenient

We aren't done with the `waitLED` command, yet. But before going on lets leave the CLI.

- abort the P4wnP1 CLI with CTRL+C (in case the looping HIDScript is still running)
- open a browser on the host yor have been using for the SSH connection to P4wnP1
- the webclient could be accessed via the same IP as the SSH server, the port is 8000 (for WiFi `http://172.24.0.1:8000`)
- navigate to the "HIDScript" tab
- from there you could load and store HIDScripts (we don't do this for now, although `ms_snake.js` is a very good 
example for the power of LED based triggers)

Replace the script in the editor Window with the following one:

``` 
return waitLED(ANY);
``` 

After hitting a run button, the right side of the window should show a new running HID job. If you press the little 
"info" button to the right of the HIDScript job, you could see details, like its state (should be running), the job ID
and the VM ID (this is the number of the JavaScript VM running this job. There are 8 of these VMs, so 8 HIDScripts could
run in parallel).

Now, if any LED change is emitted from the USB host (by toggling NUM, CAPS or SCROLL) the HIDScript job should end. 
It still could be found under "Succeeded" jobs.

If you press the little "info" button again, there should be an information about the result value (encoded as JSON),
which looks something like this:

```
{"ERROR":false,"ERRORTEXT":"","TIMEOUT":false,"NUM":true,"CAPS":false,"SCROLL":false,"COMPOSE":false,"KANA":false}
```
  
So the `waitLED` command returns a JavaScript object looking like this:

```
{
	ERROR:		false,	// gets true if an error occurred (f.e. HIDScript was aborted, before waitLED could return)  
	ERRORTEXT: 	"",		// corresponding error string
	TIMEOUT:	false,	// gets true if waitLED timed out (more on this in a minute)
	NUM:		true,   // gets true if NUM LED had changed before waitLED returned
	CAPS:		false,  // gets true if CAPS LED had changed before waitLED returned
	SCROLL:		false,  // gets true if SCROLL LED had changed before waitLED returned
	COMPOSE:	false,  // gets true if COMPOSE LED had changed before waitLED returned (uncommon)
	KANA:		false   // gets true if KANA LED had changed before waitLED returned (uncommon)
}
```

In my case, `NUM` became true. In your case it maybe was `CAPS`. It doesn't matter, what does matter is the fact, that
this return value gives the opportunity to take branching decisions in a HIDScript, based on LED state changes issued 
from the target USB host.

Let's try an example:

```
while (true) {
 result = waitLED(ANY);
 if (result.NUM) {
   type("NUM has been toggled\n");
 }
 if (result.SCROLL) {
   type("SCROLL has been toggled\n");
 }
 if (result.CAPS) {
   break; //exit loop
 }
}
``` 

Assuming the given script is already running, pressing NUM on the USB host should result in typing out "NUM has been 
toggled", while pressing SCROLL LOCK results in the type text "SCROLL has been toggled". This behavior repeats, until
CAPS LOCK is pressed and the resulting LED changes ends the loop.

Puhhh ... a bunch of text on this command, there still some things left.

We provided arguments like `NUM`, `ANY` or `ANY_OR_NONE` commands to `waitLED` without much explanation.
In fact `waitLED` accepts up to two arguments: 

The first argument, as you might have guessed, is a whitelist filter for the LEDs to watch. Valid arguments are:
- `ANY` (react on a change to any of the LEDs)
- `ANY_OR_NONE` (react on every new LED state, even if there's no change)
- `NUM` (ignore all LED changes, except on the NUM LED)
- `CAPS` (ignore all LED changes, except on the NUM CAPS)
- `SCROLL` (ignore all LED changes, except on the NUM SCROLL)
- multiple filters could be combined like this `CAPS | NUM`, `NUM | SCROLL`

The second argument, we haven't used so far, is a timeout duration in milliseconds. If no LED change occurred during 
this timeout, `waitLED` returns and has `TIMEOUT: true` set in the result object (additionally `ERROR` is set to true 
and `ERRORTEXT` indcates a timeout).

The following command would wait for a change on the NUM LED for up to 5 seconds:

```
waitLED(NUM,5000)
```

Even though `waitLED` is a very powerful command if used correctly, it doesn't solve the easy task of robustly pausing
a script till a state modifier key is pressed on the target USB host (rember: unintended return, caused by preserved LED
state changes).

This is where `waitLEDRepeat` joins the game.

Paste the following script into the editor and try to make the command return. Inspect the HIDScript results afterwards.
```
return waitLEDRepeat(ANY)
```

You should quickly notice, that the same LED has to be changed frequently times, in order to make the command return.
It doesn't return if different LEDs change or if the changes happen to slow. The provided argument (`ANY` in the 
example) serves the same purpose as with `waitLED`. It is a whitelist filter. `waitLEDRepeat(NUM)` would only return
on changes for the NUM LOCK LED - no matter how fast and often you'd press CAPS, it won't return.

By default, one of the whitelisted LEDs has to change 3 times and the delay between two successive changes mustn't be
greater than 800 milliseconds. This behavior could be influenced, by providing additional arguments like in this 
example:

```
filter = ANY;		// same filters as for waitLED
num_changes = 5;	// how often the SAME LED has to change, in order to return from waitLEDRepeat
max_delay = 800;	// the maximum duration between two LED changes, which should be taken into acccount (milliseconds)
timeout = 10000;    // timeout in milliseconds

waitLEDRepeat(filter, num_changes, max_delay); 			//wait till a LED frequently changed 5 times, no timeout
waitLEDRepeat(filter, num_changes, max_delay, timeout); //wait till a LED frequently changed 5 times, abort after 10 seconds
```

So that's how to interact with LED reports from an USB host in HIDScript.

There's one thing left, which should be known: `waitLEDRepeat` doesn't differ from `waitLED`, when it comes to 
consumption of preserved LED state changes. Anyways, it is much harder to trigger it unintended. So `waitLEDRepeat` is
the right choice, if the task is to pause HIDScripts till human interaction happens. Of course it could be used for
branching, too. 


Up to this point we gained a good bit of knowledge on HIDScript (of course not everything, we haven't even looked into 
mouse control). Anyways, this text is about P4wnP1 A.L.O.A. workflow and concepts. So we don't look into mouse support
right now.

Let's summarize what information we gathered about P4wnP1's workflow and concepts so far:
- we could start actions like keystroke injection from the CLI client, on-demand
- we could use the webclient to achieve the same goal
- with an external powers supply connected, we could change USB hosts and go on working seamlessly 
- we could configure the USB stack exactly to our needs ()and change the configuration at runtime)
- we could write multi purpose HIDScripts, with complex logic based on JavaScript (support for functions, loops, 
branching etc. etc.)

### 3. Workflow part 2 - Templating and TriggerActions

Before moving on to the other major concepts of P4wnP1 A.L.O.A. let's refine our first goal, which was to "run a 
keystroke injection against a USB host":

- The new goal is to type "Hello world" into the editor of a Windows USB host (notepad.exe). 
- The editor should be opened by P4wnP1 (not manually by the user).
- The editor should automatically be closed, when any of the keyboard LEDs of the USB host is toggled once.
- Everytime P4wnP1 is attached to the USB host, this behavior should repeat (with external power supply, no reboot of 
P4wnP1)
- The script should only run ones, unless P4wnP1 is re-attached to the USB host, even if successive keyboard LED changes
occur. 
- Even if P4wnP1 is rebooted, the same behavior should be recoverable without recreating everything from scratch.

Starting notepad, typing "Hellow world" and closing notepad after a LED change could be done with this HIDScript: 

```
// Starting notepad
press("WIN R");         // Windows key + R, to open run dialog
delay(500);             // wait 500ms for the dialog to open
type("notepad.exe\n"); 	// type 'notepad.exe' to the run dialog, append a RETURN press
delay(2000);            // wait 2 seconds for notepad to come up

// Type the message
type("Hello world")     // Type "Hello world" to notepad

// close notepad after LED change
waitLED(ANY);           // wait for a single LED change
press("ALT F4");        // ALT+F4 shortcut to close notepad

//as we changed content, there will be a confirmation dialog before notepad exits
delay(500);             // wait for the confirmation dialog
press("RIGHT");         // move focus to next button (don't save) with RIGHT ARROW
press("SPACEBAR");      // confirm dialog with space
```

The only thing new here is the `delay` command, which doesn't need much explanation. It delays execution for the given
amount of milliseconds. The script could be pasted into the webclient HIDScript editor and started with "run" in order 
to test it.

It should work as intended, so we are nearly done. In order to be able reuse the script, even after a reboot, we store 
it persistently. This could be done by hitting the "store" button in the HIDScript tab of the webclient. After entering
the name `tutorial1` and confirming the dialog the script should be stored. We could confirm this, by hitting the 
"Load & Replace" button in the webclient, the stored script should be in the list and named `tutorial1.js` (the `.js`
extension is appended automatically, if not already done in the "store" dialog).

**Warning: If a name of an existing file is written to the store dialog, the file gets overwritten without further
confirmation.**

Let's try to start the exact same script via CLI from a SSH session now:
```
P4wnP1_cli hid run tutorial1.js
```

So it is possible to start the HIDScript from other applications, supporting shell command or from a bash script.

It would even be possible to start the script remotely from a CLI client compiled for Windows. Assuming the Windows host
is able to reach P4wnP1 A.L.O.A. via WiFi and the IP of P4wnP1 is set to `172.24.0.1` the command would look like this:
```
P4wnP1_cli.exe --host 172.24.0.1 hid run tutorial1.js
```

*Note: At the time of this writing, I haven't decided yet if P4wnP1 A.L.O.A. ships a CLI binary for each and every 
possible platform and architecture. But it is likely that precompiled versions for major platforms are provided. If
not this shouldn't be a problem. As cross compilation of the Go code takes less than a minute.*

The next step is to allow the script to start again, when P4wnP1 is re-attached to the USB host. A approach we already
used, is to wrap everything into a loop and prepend a `waitLED(ANY_OR_NONE)` to continue when the USB host signals that
the keyboard driver is ready to receive input. A modified script could look like this:

```
while (true) {
  waitLED(ANY_OR_NONE);     // wait till keyboard driver sends the initial LED state
  
  // Starting notepad
  press("WIN R");           // Windows key + R, to open run dialog
  delay(500);               // wait 500ms for the dialog to open
  type("notepad.exe\n");    // type 'notepad.exe' to the run dialog, append a RETURN press
  delay(2000);              // wait 2 seconds for notepad to come up

  // Type the message
  type("Hello world")       // Type "Hello world" to notepad

  // close notepad after LED change
  waitLED(ANY);       // wait for a single LED change
  press("ALT F4");          // ALT+F4 shortcut to close notepad

  //as we changed content, there will be a confirmation dialog before notepad exits
  delay(500);               // wait for the confirmation dialog
  press("RIGHT");           // move focus to next button (don't save) with RIGHT ARROW
  press("SPACEBAR");        // confirm dialog with space 
}
```

Indeed the script would run, every time we attach P4wnP1 to the host, but it isn't very robust, because there's a second
`waitLED` involved before notepad is closed. There are several error cases, for example if P4wnP1 is detached before
the "Hello world" is written. The blocking `waitLED` would be the one before `press("ALT F4")` now and execution would 
continue from this point on reattach.

A definitive kill criteria for this approach: The requirement that the script should be run only once isn't met anymore.
Hitting NUM LOCK repeatedly would restart the script, again and again.

So how to solve this ?

#### Let's introduce TriggerActions

The solution to the problem are TriggerActions. As the name implies, the concept is to fire actions, based on predefined
triggers.

To get an idea of what I'm talking about, head over to the "TRIGGER ACTIONS" tab on the webclient. Depending on the 
current setup, there may already exist TriggerActions, which doesn't matter for now.

Hitting the "ADD ONE" button add a new TriggerActions and opens it in edit mode. The TriggerAction is turned off by 
default and has to be enabled in order to be editable. So we toggle the enable switch.

Now from the pull down menu called "Trigger" the option "USB gadget connected to host" should be selected. The action
should have a preset of "write log entry". We leave it like this and hit the "Update" button. The newly added 
TriggerAction should be visible in the overview now (the one with the highest ID) and show a summary of the selected
Trigger and selected Action.

To test how the trigger works, navigate to the "Event Log" tab of the webclient. Make sure you have the webclient open
via WiFi (not USB ethernet). Apply external power to P4wnP1, disconnect it from the USB host and connect it again.
A log message should be pushed to the client immediately every time P4wnP1 is attached to the USB host.

Repeating this a few times, it becomes pretty obvious that this trigger fires in an early USB enumeration phase. To be 
precise: When this trigger fires, it is clear that P4wnP1 has been connected to the USB host, but there's no guarantee
that the host managed to load all needed drivers. In fact it is very unlikely that the USB keyboard driver is loaded 
when the trigger fires. We have to keep this in mind.

Before we move on with our task, we do an additional test. Heading back to the "TriggerAction" tab and pressing the 
little blue button looking like a pen, on our newly created TriggerAction we end up in edit mode again. Enable the 
`One shot` switch now. Head back to the "Event Log" and detach and re-attach P4wnP1 from the USB host again. This time
the Trigger fires only once, no matter how often P4wnP1 is re-attached to the USB host afterwards. It is worth 
mentioning that a "One shot" TriggerAction isn't deleted after the TriggerFired, instead the TriggerAction is disabled.
Re-enabling allows to reuse the TriggerAction without redefining it. Nothing gets lost until the red "trash" button is
hit on a TriggerAction.

**Warning: If the delete button for a TriggerAction is clicked, the TriggerAction is deleted without further 
confirmation.**

At this point let's do the obvious. We edit the created TriggerAction and select "start a HIDScript" instead of "write
log entry" for the action to execute. A new input field called "script name" is shown. Clicking on this input field 
brings up a selection dialog for all stored HIDScripts, including our newly created `tutorial1.js`.

Before we test if this works, a quick note on the action "write log entry": P4wnP1 A.L.O.A. doesn't keep track of 
TriggerActions which have already be fired. This means the log entries created by a "write log entry" action are 
delivered to all listening client, but not stored (for various reasons). The webclient on the other hand stores the log
entry until the single page application is reloaded. The same applies to events related to HIDScript jobs. If a 
HIDScript ends (with success or error), a vent is fired and delivered to the webclient. In fact the webclient has a 
runtime state, which holds more information than the core service. If the runtime state of the webclient grows to large,
one only needs to reload the client to clear "historical" sate information. If the core service would store historical
information, it would get out of resources very soon. Thus this concept applies to most sub systems of P4wnP1 A.L.O.A.

Now back to our task. We have a TriggerAction ready, which fires our HIDScript every time P4wnP1 is attached to an USB
host. Depending on the USB host, this works more or less reliably. In my test setup it didn't work at all and there's
a reason. Let's review the first few lines our HIDScript:

```
// Starting notepad
press("WIN R");         // Windows key + R, to open run dialog
delay(500);             // wait 500ms for the dialog to open
type("notepad.exe\n"); 	// type 'notepad.exe' to the run dialog, append a RETURN press
... snip ...
```

Recalling the fact, that the "USB gadget connected" Trigger fires in early USB enumeration phase, when the USB host's 
keyboard driver hasn't necessarily loaded, the problem becomes obvious. We have to prepend some kind of delay to the
script to assure the keyboard driver is up. As we already know that it isn't possible to predict the optimal delay, we
go with the `waitLED(ANY_OR_NONE)` approach, instead. The new script should look like this:
```
waitLED(ANY_OR_NONE);   //assure keyboard driver is ready

// Starting notepad
press("WIN R");	        // Windows key + R, to open run dialog
delay(500);             // wait 500ms for the dialog to open
type("notepad.exe\n"); 	// type 'notepad.exe' to the run dialog, append a RETURN press
delay(2000);            // wait 2 seconds for notepad to come up

// Type the message
type("Hello world")     // Type "Hello world" to notepad

// close notepad after LED change
waitLEDRepeat(ANY);     // wait for a single LED change
press("ALT F4");        // ALT+F4 shortcut to close notepad

//as we changed content, there will be a confirmation dialog before notepad exits
delay(500);             // wait for the confirmation dialog
press("RIGHT");         // move focus to next button (don't save) with RIGHT ARROW
press("SPACEBAR");      // confirm dialog with space
```

Now store the modified script under the exact same name `tutorial1`, as already pointed out, the old script gets 
overwritten without further confirmation. There's no need to adjust the TriggerAction, because the script name hasn't 
changed.

At this point everything should work as intended. Anyways, if P4wnP1 is rebooted or looses power, the HIDScript persists
but the TriggerAction is gone. To cope with that, TriggerActions could be stored, too.

The "store" button in the "TriggerAction" tab works exactly like the one in the HIDScript editor. It is worth mentioning
that all currently active TriggerActions get stored if the "store" dialog is confirmed (including the disabled ones).
Best practice is to delete all TriggerActions which don't belong to the task in current scope before storing (they 
already should have been stored earlier) and only store a small set of TriggerActions relevant to the current task, 
using a proper name. There're two options to load back stored TriggerActions:
 - "load & replace" clears all active trigger actions and loads only the stored ones
 - "load & add" keeps the already active TriggerActions and adds in the stored ones. Thus "load & add" could be used to 
 build a complex TriggerAction set out of smaller sets. The resulting set could then again be stored.
 
So for now, we should only store our single TriggerAction, which runs the HID script. The name could be `tutorial1` 
again and won't conflict with the HIDScript called `tutorial1`.

Confirm successful storing, by hitting the "load&replace" button in the "TriggerAction" tab. The stored TriggerAction 
set should be in the list and named `tutorial1`.

**Warning: The TriggerAction "load" dialogs allow deleting stored TriggerActions by hitting the red "trash" button
next to each action. Hitting the button permanently deletes the respective TriggerAction set, without further 
confirmation**

At this point we could safely delete our TriggerAction from the "TriggerActions" tab (!!not from the load dialog!!).

With the TriggerAction deleted, nothing happens if we detach and re-attach P4wnP1 from the USB host.

The stored TriggerAction set persists reboots. Instead of reloading the TriggerAction set from with the webclient, we
test the CLI.

Lets take a quick look into the help screen of the `template deploy` sub-command:
 
```
root@kali:~# P4wnP1_cli template deploy -h
Deploy given gadget settings

Usage:
  P4wnP1_cli template deploy [flags]

Flags:
  -b, --bluetooth string         Deploy Bluetooth template
  -f, --full string              Deploy full settings template
  -h, --help                     help for deploy
  -n, --network string           Deploy network settings template
  -t, --trigger-actions string   Deploy trigger action template
  -u, --usb string               Deploy USB settings template
  -w, --wifi string              Deploy WiFi settings templates

Global Flags:
      --host string   The host with the listening P4wnP1 RPC server (default "localhost")
      --port string   The port on which the P4wnP1 RPC server is listening (default "50051")

``` 

The usage screen shows, that TriggerAction Templates could be deployed with the `-t` flag. By running the following 
command our TriggerAction should get loaded again and thus the HIDScript should trigger like in our tests, if P4wnP1 is
detached/re-attached from the USB host:

``` 
P4wnP1_cli template deploy -t tutorial1
``` 

