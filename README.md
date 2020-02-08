# P4wnP1 A.L.O.A.

P4wnP1 A.L.O.A. by MaMe82 is a framework which turns a Rapsberry Pi Zero W into a flexible, low-cost platform for 
pentesting, red teaming and physical engagements ... or into "A Little Offensive Appliance".

## 0. How to install

The latest image could be found under release tab.

The easiest way to access a fresh P4wnP1 A.L.O.A. installation is to use the web client via the spawned WiFi (the PSK
is `MaMe82-P4wnP1`, the URL `http://172.24.0.1:8000`) or SSH (default password `toor`).

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
- current language layouts: br, de, es, fr, gb, it, ru and us 
  
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
- all features mentioned so far, can be configured using a CLI client
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

Although it wasn't planned initially, P4wnP1 A.L.O.A. can be configured using a webclient.
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

# Usage tutorial

## 2. Workflow part 1 - HIDScript

P4wnP1 A.L.O.A. doesn't use concepts like static configuration or payloads. In fact it has no static workflow at all.
 
P4wnP1 A.L.O.A. is meant to be as flexible as possible, to allow using it in all possible scenarios (including the ones
I couldn't think of while creating P4wnP1 A.L.O.A.).

But there are some basic concepts, I'd like to walk through in this section. As it is hard to explain everything without
creating a proper (video) documentation, I visit some common use cases and examples in order to explain what needs
to be explained.

Nevertheless, it is unlikely that I'll have the time to provide a full-fledged documentation. **So I encourage everyone
to support me with tutorials and ideas, which could be linked back into this README**

Now let's start with one of the most basic tasks:

### 2.1 Run a keystroke injection against a host, which has P4wnP1 attached via USB

The minimum configuration requirement to achieve this goal is:
- The USB sub system is configured to emulate at least a keyboard
- There is a way to access P4wnP1 (remotely), in order to initiate the keystroke injection

The default configuration of P4wnP1's (unmodified image) meets these requirements already:
- the USB settings are initialized to provide **keyboard**, mouse and ethernet over USB (both, RNDIS and CDC ECM) 
- P4wnP1 could already be accessed remotely, using one of the following methods:
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
Deploying a HTTPS connection is currently not in scope of the project. So please keep this in mind, if you handle 
sensitive data, like WiFi credentials, in the webclient. The whole project isn't built with security in mind (and it is 
unlikely that this will ever get a requirement). So please deploy appropriate measures (f.e. restricting access
to webclient with iptables, if the Access Point is configured with Open Authentication; don't keep Bluetooth 
Discoverability and Connectability enabled without PIN protection etc. etc.)*

At this point I assume:
1) You have attached P4wnP1 to some target host via USB (the innermost of the Raspberry's micro USB ports is the one to 
use)
2) The USB host runs an application, which is able to receive the keystrokes and has the current keyboard input focus 
(f.e. a text editor)
3) You are remotely connected to P4wnP1 via SSH (the best way is WiFi), preferably the SSH connection is running from
a different host, then the the one which has P4wnP1 A.L.O.A. attached over USB

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


The help screen already shows, that the CLI client uses different commands to interact with the various subsystems of 
P4wnP1 A.L.O.A. Most of these commands have own sub-commands, again. The help for each command or sub-command could be 
accessed by appending `-h` to the CLI command:

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

*If your SSH client runs on the USB host itself, the typed "Hello world" ends up somewhere between the resulting output 
of the CLI command (it doesn't belong to the output, but has been typed in between).*

**Goal achieved. We injected keystrokes to the target.**
 
Much reading for a simple task like keystroke injection, but again, this section is meant to explain basic concepts.

### 2.2 Moving on to more sophisticated language features of HIDScript

If you managed to run the "Hello world" keystroke injection, this is a good point to explore some additional HIDScript
features. 

We already know the `type` command, but let's try and discuss some more sophisticated HIDScript commands: 

#### Pressing special keys and combinations

The `type` command supports pressing return, by encoding a "new line" character into the input string, like this:
```
P4wnP1_cli hid run -c 'type("line 1\nline 2\nline 3 followed by pressing RETURN three times\n\n\n")'
```

But what about special keys or key combinations? 

The `press` command comes to help!

Let's use `press` to send CTRL+ALT+DELETE to the USB host:

```
P4wnP1_cli hid run -c 'press("CTRL ALT DELETE")'
```

*Note: Two of keys have been modifiers (CTRL and ALT) and only one has been an actual key (DELETE)*

Let's press the key 'A' without any modifier key:

```
P4wnP1_cli hid run -c 'press("A")'
```

The resulting output should be a lowercase 'a', because `press("A")` interprets 'A' as key. The command `type("A")`, 
on the other hand, tries to press a key combination which should result in an uppercase 'A' output character.

Let's combine a modifier and a non-modifier key, in order to produce an uppercase 'A' output character (mimic the 
behavior of `type("A"):

```
P4wnP1_cli hid run -c 'press("SHIFT A")'
```

This should have produced an uppercase A output.

It is important to understand, that `press` interprets the given its key arguments as keys, while type tries to find the
appropriate key combinations to produce the intended output characters.   

In a last example, let's combine `press` and `type`.

```
P4wnP1_cli hid run -c 'type("before caps\n"); press("CAPS"); type("after caps\n"); press("CAPS");'
```
 
The last command typed a string, toggled CAPSLOCK, typed another string and toggled CAPS lock again. 
In result, CAPSLOCK should be in its initial state (toggled two times), but one of the strings is typed uppercase, the
other lowercase although both strings have been given in lower case.

Additional notes on key presses with `press`: 

I don't want to dive into the depth of USB keyboard reports inner workings, but some things are worth mentioning to 
pinpoint the limits and possibilities of the `press` command (which itself works based on raw keyboard reports):
- a keyboard report can contain up to 8 modifier keys at once
- the modifier keys are
  - LEFT_CTRL
  - RIGHT_CTRL
  - LEFT_ALT
  - RIGHT_ALT
  - LEFT_SHIFT
  - RIGHT_SHIFT
  - LEFT_GUI
  - RIGHT_GUI
- P4wnP1 allows using aliases for common modifiers
  - CTRL == CONTROL == LEFT_CTRL
  - ALT == LEFT_ALT
  - SHIFT == LEFT_SHIFT
  - WIN == GUI == LEFT_GUI
- in addition to the modifiers, `press` consumes up to six normal or special keys
  - normal keys represent characters and special keys
  - example of special keys: BACKSPACE, ENTER (== RETURN), F1 .. F12)
  - the keys are language layout agnostic (`press("Z")` results in USB_KEY_Z fo EN_US keyboard layout, but produces
  USB_KEY_Y for a German layout. This corresponds to pressing the hardware key 'Z' on a German keyboard, which would 
  produce a USB_KEY_Y, too.)
  - `/usr/local/P4wnP1/keymaps/common.json` holds a formatted JSON keymap with all possible keys (be careful not to 
  change the file) 
- **adding multiple keys to the a single `press` command, doesn't produce a key sequence.** All given given keys are 
pressed at the same time and release at the same time.
- `press` releases keys automatically, this means a sequence like "hold ALT, press TAB, press TAB, release ALT" 
currently isn't possible 

#### Keyboard layout

The HIDScript command to change the keyboard layout is `layout(<language map name>)`.

The following example switches keyboard layout to 'US' types something and switches the layout to 'German' before it
goes on typing:

```
P4wnP1_cli hid run -c 'layout("us"); type("Typing with EN_US layout\n");layout("de"); type("Typing with German layout supporting special chars üäö\n");'
```
 
The output result of the command given above, depends on the target layout used by the USB host. 

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

Please note, that the intended output is only achieved, if P4wnP1's keyboard layout aligns with the keyboard layout
actually used by the USB host. 

The `layout` command allows to align P4wwP1's internal layout to the one of the target USB host. 

Being able to change the layout in the middle of a running HIDScript, could come in handy: Who knows, maybe you like to 
brute force the target host's keyboard layout by issuing commands with changing layouts till one of the typed commands
achieves the desired effect.

**Important:** The layout has global effect. This means if multiple HIDScripts are running concurrently and one of the 
scripts sets a new layout, all other scripts are effected immediately, too.

#### Typing speed

By default P4wnP1 injects keystrokes as fast as possible. Depending on your goal, this could be a bit too much (think of 
counter measures which prevent keystroke injection based on behavior analysis of typing speed). HIDScript supports a 
command to change this behavior.

`typingSpeed(delayMillis, jitterMillis)`

The first argument to the `typingSpeed` command represents a constant delay in milliseconds, which is applied between 
two keystrokes. The second argument is an additional jitter in milliseconds. It adds an additional random delay, which 
scales between 0 and the given jitter in milliseconds, to the static delay provided with the first argument.

Let's try to use `typingSpeed` to slow down the typing:

```
P4wnP1_cli hid run -c 'typingSpeed(100,0); type("Hello world")'
```

Next, instead of a constant delay, we try a random jitter:
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

Waiting for LED report, or to be precise LED state changes, is one of the more sophisticated keyboard features of 
HIDScript. It could be very powerful but needs a bit of explanation.

You may have noticed that (depending on the USB host's OS) the keyboard state modifiers (NUM LOCK, SCROLL LOCK, 
CAPS LOCK) are shared across multiple connected keyboards. For example, if you connect two keyboards to a Windows host, 
and toggle CAPS LOCK on one of them, the CAPS LOCK LED changes on both keyboards. 

Exactly this test could be used, to determine if the keyboard state modifiers are shared across all keyboards for a 
given OS. 

In case a USB host supports this kind of state sharing (for example Windows does), P4wnP1's HIDScript language could 
make use out of it.

Imagine the following scenario:

P4wnP1 is connected to a USB host and you want to apply keystroke injection, but you don't want the HIDScript to 
run the keystrokes immediately. Instead the HIDScript should sit and wait till you hit NUMLOCK, CAPSLOCK or SCROLLLOCK
on the host's real keyboard. Why? Maybe you're involved in an engagement, somebody walked in and you don't want that
this exact "somebody" could see how magically a huge amount of characters are typed into a console window which suddenly
popped up. So you wait till "somebody" walks out, hit NUM LOCK and ultimately a console window pops up and a huge amount
of characters are magically type ... I think you got it. 

The described behavior could be achieved like this:

```
P4wnP1_cli hid run -c 'waitLED(NUM); type("A huge amount of characters\n")'
```

If you tested the command above, typing should only start if NUM LOCK is pressed on the USB host's hardware keyboard, 
but you might encounter cases where the keystrokes immediately are issued, even if NUM LOCK wasn't pressed (and the 
keyboard LED hasn't changed).

This is intended behavior and the reason for this is another use case for the `waitLED` command:
 
Maybe you have used other keyboard scripting languages and other USB devices capable of injecting keystrokes, before. 
Most of these devices share a common problem: You don't know when to start typing! 

If you start typing immediately after the USB device is powered up, it is likely that the USB host hasn't finished 
device enumeration and thus hasn't managed to bring up the keyboard drivers. Ultimately your keystrokes are lost.

To overcome this you could add a delay before the keystroke injection starts. But how long should this delay be? Five 
seconds, 10 seconds, 30 seconds ? 

The answer is: it depends! It depends on how fast the host is able to enumerate the device and bring up the keyboard 
driver. In fact you couldn't know how long this takes, without testing against the actual target.

But as we have already learned, Operating Systems like Windows share the LED state across multiple keyboards.
This means if the NUMLOCK LED of the host keyboard is set to ON before you attach a second keyboard, the NUMLOCK LED 
on this new keyboard has to be set to ON, too, once attached. If the NUM LOCK LED would have been set to OFF, anyways, 
the newly attached keyboard receives the LED state (all LEDs off in this case). The interesting thing about this is,
that this "LED update" could only be send from the USB host to the attached keyboard, if the keyboard driver has 
finished loading (sending LED state wouldn't be possible otherwise).

Isn't that beautiful? The USB host tells us: "I'm ready to receive keystrokes". There is no need to play around with 
initial delays. 

But here is another problem: Assume we connect P4wnP1 to an USB host. We run a HIDScript starting with `waitLED` instead
of a hand crafted delay. Typing starts after the `waitLED`, but nothing happens - our keystrokes are lost, anyways! Why? 
Because, it is likely that we missed the LED state update, as it arrived before we even started our HIDScript. 

Exactly this "race condition" is the reason why P4wnP1 preserves all recognized LED state changes, unless at least one 
HIDScript consumes them by calling `waitLED` (or `waitLEDRepeat`). This could result in the behavior describe earlier,
where a `waitLED` returns immediately, even though no LED change occurred. We now know: The LED change indeed occurred, 
but it could have happened much earlier (berfore we even started the HIDScript), because the state change was preserved.
We also know, that this behavior is needed to avoid missing LED state changes, in case `waitLED` is used to test for
"USB host's keyboard driver readiness".

*Note: It is worth mentioning, that `waitLED` returns ONLY if the received LED state differs from P4wnP1's internal 
state. This means, even if we listen for a change on any LED with `waitLED(ANY)` it still could happen, that we receive 
an initial LED state from a USB host, which doesn't differ from P4wnP1's internal state. In this case `waitLED(ANY)` 
would block forever (or till a real LED change happens).
This special case could be handled by calling `waitLED(ANY_OR_NONE)`, which returns as soon as a new LED state arrive,
even if it doesn't result in a change.*

**Enough explanation, let's get practical ... before we do so, we have to change the hardware setup a bit:**

Attach an external power supply to the second USB port of the Raspberry Pi Zero (the outer one). This assures that
P4wnP1 doesn't loose power when detached from the USB host, as it doesn't rely on bus power anymore. The USB port which
should be used to connect P4wnP1 to the target USB host is the inner most of the two ports.

Now start the following HIDScript

``` 
P4wnP1_cli hid run -c 'while (true) {waitLED(ANY);type("Attached\n");}'
``` 

Detach P4wnP1 from the USB host (and make sure it is kept powered on)! Reattach it to the USB host ...
Every time you reattach P4wnP1 to the host, "Attached" should be typed out to the host.

This taught us 3 facts:
1) `waitLED` could be used as initial command in scripts, to start typing as soon as the keyboard driver is ready
2) `waitLED` isn't the perfect choice, to pause HID scripts until a LED changing key is pressed on the USB host, as 
preserved state changes could unblock the command in an unintended way
3) Providing more complex HIDScript as parameter to the CLI isn't very convenient

As we still aren't done with the `waitLED` command, we take care of the third fact, now. Let us leave the CLI.

- abort the P4wnP1 CLI with CTRL+C (in case the looping HIDScript is still running)
- open a browser on the host yor have been using for the SSH connection to P4wnP1 (not the USB host)
- the webclient could be accessed via the same IP as the SSH server, the port is 8000 (for WiFi 
`http://172.24.0.1:8000`)
- navigate to the "HIDScript" tab in the now opened webclient
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

In my case, `NUM` became true. In your case it maybe was `CAPS`. It doesn't matter which LED it was. "hat does matter is 
the fact, that the return value gives the opportunity to examine the LED change which makes the command return and thus
it could be used to take branching decisions in your HIDScript (based on LED state changes issued from the USB host's
real keyboard).

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
toggled", while pressing SCROLL LOCK results in the typed text "SCROLL has been toggled". This behavior repeats, until
CAPS LOCK is pressed and the resulting LED change aborts the loop and ends the HIDScript.

Puhhh ... a bunch of text on this command for a single HIDScript command, but there still some things left.

We provided arguments like `NUM`, `ANY` or `ANY_OR_NONE` to the `waitLED` command, without further explanation.

The `waitLED` accepts up to two arguments: 

The first argument, as you might have guessed, is a whitelist filter for the LEDs to watch. Valid arguments are:
- `ANY` (react on a change to any of the LEDs)
- `ANY_OR_NONE` (react on every new LED state, even if there's no change)
- `NUM` (ignore all LED changes, except on the NUM LED)
- `CAPS` (ignore all LED changes, except on the NUM CAPS)
- `SCROLL` (ignore all LED changes, except on the NUM SCROLL)
- multiple filters could be combined like this `CAPS | NUM`, `NUM | SCROLL`

The second argument, we haven't used so far, is a timeout duration in milliseconds. If no LED change occurred during 
this timeout duration, `waitLED` returns and has `TIMEOUT: true` set in the resulting object (additionally `ERROR` is 
set to true and `ERRORTEXT` indicates a timeout).

The following command would wait for a change on the NUM LED, but aborts waiting after 5 seconds:

```
waitLED(NUM,5000)
```

Even though `waitLED` is a very powerful command if used correctly, it hasn't helped to deal with our easy task of 
robustly pausing a HIDScript till a state modifier key is pressed on the target USB host (remember: We wanted to pause
execution to assure the unwanted "somebody" walked out before typing starts, but `waitLED` occasionally returned early, 
because of preserved LED state changes).

This is where `waitLEDRepeat` joins the game and comes to rescue.

Paste the following script into the editor and try to make the command return. Inspect the HIDScript results afterwards.
```
return waitLEDRepeat(ANY)
```

You should quickly notice, that the same LED has to be changed multiple times frequently, in order to make the 
`waitLEDRepeat` command return. The `waitLEDRepeat` command wouldn't return if differing LEDs change state or if the 
LED changes on a single LED are occurring too slow. 

The argument provided to `waitLEDRepeat` (which is `ANY` in the example) serves the exact same purpose as for `waitLED`. 
It is a whitelist filter. For example `waitLEDRepeat(NUM)` would only return for changes of the NUM LOCK LED - no matter
how fast and often you'd hammer on the CAPS LOCK key, it wouldn't return unlees NUM LOCK is pressed frequently.

By default, one of the whitelisted LEDs has to change 3 times and the delay between two successive changes mustn't be
greater than 800 milliseconds in order to make `waitLEDRepeat` return. This behavior could tuned, by providing 
additional arguments like shown in this example:

```
filter = ANY;		// same filters as for waitLED
num_changes = 5;	// how often the SAME LED has to change, in order to return from waitLEDRepeat
max_delay = 800;	// the maximum duration between two LED changes, which should be taken into acccount (milliseconds)
timeout = 10000;    // timeout in milliseconds

waitLEDRepeat(filter, num_changes, max_delay); 			//wait till a LED frequently changed 5 times, no timeout
waitLEDRepeat(filter, num_changes, max_delay, timeout); //wait till a LED frequently changed 5 times, abort after 10 seconds
```

So that's how to interact with LED reports from an USB host in HIDScript.

*Note: `waitLEDRepeat` doesn't differ from `waitLED`, when it comes to consumption of preserved LED state changes. 
Anyways, it is much harder to trigger it unintended.*
 
So `waitLEDRepeat` is the right choice, if the task is to pause HIDScripts till human interaction happens. Of course it 
could be used for branching, too, as it provides the same return object as `waitLED` does. 

Up to this point we gained a good bit of knowledge about HIDScript (of course not about everything, we haven't even 
looked into mouse control capabilities of this scripting language). Anyways, this tutorial is about P4wnP1 A.L.O.A. 
workflow and basic concepts. So we don't look into other HIDScript features, for now, and move on.

Let's summarize what we learned about P4wnP1's workflow and concepts so far:
- we could start actions like keystroke injection from the CLI client, on-demand
- we could use the webclient to achieve the same, while having additional control over HIDScript jobs
- if we connect an external power supply to P4wnP1 A.L.O.A., we attach/detach to/from different USB hosts and already 
started HIDScripts go on working seamlessly 
- we could configure the USB stack exactly to our needs (and change its configuration at runtime, without rebooting 
P4wnP1)
- we could write multi purpose HIDScripts, with complex logic based on JavaScript (with support for functions, loops, 
branching etc. etc.)

### 3. Workflow part 2 - Templating and TriggerActions

Before go on with the other major concepts of P4wnP1 A.L.O.A. let's refine our first goal, which was to "run a keystroke 
injection against a USB host":

- The new goal is to type "Hello world" into the editor of a Windows USB host (notepad.exe). 
- The editor should be opened by P4wnP1 (not manually by the user).
- The editor should automatically be closed, when any of the keyboard LEDs of the USB host is toggled.
- Everytime P4wnP1 is attached to the USB host, this behavior should repeat (with external power supply, no reboot of 
P4wnP1)
- The process *should only run once*, unless P4wnP1 is re-attached to the USB host, even if successive keyboard LED 
changes occur after the HIDScript has been started.
- Even if P4wnP1 is rebooted, the same behavior should be recoverable without recreating detail of the setup from 
scratch, again.

Starting notepad, typing "Hello world" and closing notepad after a LED change could be done with the things we learned 
so far. An according HIDScript could look something like this: 

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

The only thing new in this script is the `delay` command, which doesn't need much explanation. It delays execution for 
the given amount of milliseconds. 

The script could be pasted into the webclient HIDScript editor and started with "run" in order to test it.

It should work as intended, so we are nearly done. In order to be able reuse the script, even after a reboot, we store 
it persistently. This could be achieved by hitting the "store" button in the HIDScript tab of the webclient. After 
entering a name (we use `tutorial1` for now) and confirming the dialog, the HIDScript should have been stored. 
We could check this, by hitting the "Load & Replace" button in the webclient. The stored script should occur in the list
of stored scripts with the name `tutorial1.js` (the `.js` extension is appended automatically, if it hasn't been 
provided in the "store" dialog, already).

**Warning: If a name of an already existing file used in the store dialog, the respective file gets overwritten without 
asking for further confirmation.**

Let's try to start the stored script using the CLI client from a SSH session, like this:
```
P4wnP1_cli hid run tutorial1.js
```

This should have worked. This means, it is possible to start stored HIDScripts from all applications which support shell
commands or from a simple bash script, by using the P4wnP1 A.L.O.A. CLI client.

It would even be possible to start the script remotely from a CLI client compiled for Windows. Assuming the Windows host
is able to reach P4wnP1 A.L.O.A. via WiFi and the IP of P4wnP1 is set to `172.24.0.1` the proper command would look like
this:
```
P4wnP1_cli.exe --host 172.24.0.1 hid run tutorial1.js
```

*Note: At the time of this writing, I haven't decided yet if P4wnP1 A.L.O.A. ships a CLI binary for each and every 
possible platform and architecture. But it is likely that precompiled versions for major platforms are provided. If
not - this isn't a big problem, as cross-compilation of the CLI client's Go code takes less than a minute.*

The next step is to allow the script to run again, every time P4wnP1 is re-attached to a USB host. A approach we already
used to achieve such a behavior, was to wrap everything into a loop and prepend a `waitLED(ANY_OR_NONE)`. The
`waitLED(ANY_OR_NONE)` assured tha the loop only continues, if the target USB host signals that the keyboard driver is 
ready to receive input by sending an update of the global keyboard LED sate. An accordingly modified script could look 
like this:

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

The script given above, indeed, would run, every time P4wnP1 is attached to an USB host. But the script isn't very 
robust, because there's a second `waitLED` involved, which waits till notepad.exe should be is closed, again. 

Doing it like this involves several issues. For example if P4wnP1 is detached before the "Hello world" is typed out, the
now blocking `waitLED` would be the one before `press("ALT F4")` and execution would continue at exact this point of
the HIDScript once P4wnP1 is attached to a (maybe different) USB host, again.

A definitive kill criteria for the chosen approach is the following problem: The requirement that the script should be 
run only once after attaching P4wnP1 to an USB host couldn't be met, as hitting NUM LOCK multiple times would restart 
the script over and over.

So how do we solve this ?

#### Let's introduce TriggerActions

The solution to the problem are so called "TriggerActions". As the name implies, this P4wnP1 A.L.O.A. workflow concept 
fires actions based on predefined triggers.

To get an idea of what I'm talking about, head over to the "TRIGGER ACTIONS" tab on the webclient. Depending on the 
current setup, there may already exist TriggerActions. We don't care for existing TriggerActions, now.

Hit the "ADD ONE" button and a new TriggerActions should be added and instantly opened in edit mode. 
The new TriggerAction is disabled by default and has to be enabled in order to make it editable. So we toggle the enable
switch.

Now from the pull down menu called "Trigger" the option "USB gadget connected to host" should be selected. The action
should have a preset of "write log entry" selected. We leave it like this and hit the "Update" button. 

The newly added TriggerAction should be visible in the TriggerActions overview now (the one with the highest ID) and 
show a summary of the selected Trigger and selected Action in readable form.

To test if the newly defined TriggerAction works, navigate over to the "Event Log" tab of the webclient. 
Make sure you have the webclient opened via WiFi (not USB ethernet). Apply external power to P4wnP1, disconnect it from 
the USB host and connect it again. A log message should be pushed to the client every time P4wnP1 is attached to a USB 
host, immediately.

If you repeated this a few times, you maybe noticed that the "USB gadget connected to host" trigger fires very fast
(or in an early stage of USB enumeration phase). To be more precise: When this trigger fires, it is known that P4wnP1 
was connected to a USB host, but there is no guarantee that the USB host managed to load all needed USB device drivers. 
**In fact it is very unlikely that the USB keyboard driver is loaded when the trigger fires. We have to keep this in 
mind.**

Before we move on with our task, we do an additional test. Heading back to the "TriggerAction" tab and we press the 
little blue button looking like a pen for our newly created TriggerAction. We end up in edit mode again.
 
This time, we enable the `One shot` option. Head back to the "Event Log" afterwards, and again, detach and re-attach 
P4wnP1 from the USB host. This time the TriggerAction should fire only once. No matter how often P4wnP1 is re-attached 
to the USB host afterwards, no new log message indicating a USB connect should be created. 

It is worth mentioning that a "One shot" TriggerAction isn't deleted after the Trigger has fired. Instead the 
TriggerAction is disabled, again. Re-enabling allows reusing a TriggerAction without redefining it. Nothing gets lost 
until the red "trash" button is hit on a TriggerAction, which will delete the respective TriggerAction.

**Warning: If the delete button for a TriggerAction is clicked, the TriggerAction is deleted permanently without further 
confirmation.**

At this point let's do the obvious. We edit the created TriggerAction and select "start a HIDScript" instead of "write
log entry" for the action to execute. Additionally we disable "one-shot", again. A new input field called "script name" 
is shown. Clicking on this input field brings up a selection dialog for all stored HIDScripts, including our formerly 
created `tutorial1.js` HIDScript.

*Before we test if this works, let me make a quick note on the action "write log entry": P4wnP1 A.L.O.A. doesn't keep 
track of Triggers which have already be fired. This means the log entries created by a "write log entry" action are 
delivered to all listening client, but aren't stored by the P4wnP1 service (for various reasons). The webclient on the 
other hand stores the log entry until the the webclient itself is reloaded. The same applies to events, which are 
related to HIDScript jobs. If a HIDScript ends (with success or error), an event pushed to all currently open 
webclients. In summary, each webclient has a runtime state, which holds more information than the core service, itself. 
If the runtime state of the webclient grows to large (too much memory usage), one only needs to reload the client to 
clear "historical" sate information. If the core service would behave the same and store every historical information, 
it would run out of resources very soon. Thus this concept applies to most sub systems of P4wnP1 A.L.O.A.*

Now back to our task. We have a TriggerAction ready, which should fire our HIDScript every time P4wnP1 is attached to 
an USB host. 

Depending on the target USB host, this works more or less reliably. In my test setup it didn't work at all and there's
a reason:
 
Let's review the first few lines our HIDScript:

```
// Starting notepad
press("WIN R");         // Windows key + R, to open run dialog
delay(500);             // wait 500ms for the dialog to open
type("notepad.exe\n"); 	// type 'notepad.exe' to the run dialog, append a RETURN press
... snip ...
```

Recalling the fact, that the "USB gadget connected" Trigger fires in early USB enumeration phase and the USB host's 
keyboard driver hasn't necessarily been loaded, the problem becomes obvious. We have to prepend some kind of delay to 
the script to assure the keyboard driver is up (otherwise our keystrokes would end up in nowhere).

As we already know that it isn't possible to predict the optimal delay, we go with the `waitLED(ANY_OR_NONE)` approach, 
explained earlier. The new script looks like this:
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

Storing the modified script under the exact same name (`tutorial1`) overwrites the former HIDScript without further
confirmation, as already pointed out. Thus there is no need to adjust our TriggerAction, as the HIDScript name the 
TriggerAction refers hasn't changed.

With this little change everything should work as intended and the script should trigger everytime we attach to an USB 
host, but only run once.

Now, if P4wnP1 is rebooted or looses power, our HIDScript would survive, because we have stored it persistently, but the
TriggerAction would be gone. Needless to say, that TriggerActions could be stored persistently, too.

The "store" button in the "TriggerAction" tab works exactly like the one in the HIDScript editor. It should be noted
that *all currently active TriggerActions* will be stored if the "store" dialog is confirmed (including the disabled 
ones).
The best practice is to delete all TriggerActions which don't belong to the task in current scope before storing (they 
should have been stored earlier, if needed) and to only store the small set of TriggerActions relevant to the current 
task, using a proper name. There are two options to load back stored TriggerActions to the active ones:
 - "load & replace" clears all active trigger actions and loads only the stored ones
 - "load & add" keeps the already active TriggerActions and adds in the stored ones. Thus "load & add" could be used to 
 build a complex TriggerAction set out of smaller sets. The resulting set could then, again, be stored.
 
For now we should only store our single TriggerAction, which starts out HIDScript. The name we use to store is 
`tutorial1` again and won't conflict with the HIDScript called `tutorial1`.

Confirm successful storing, by hitting the "load&replace" button in the "TriggerAction" tab. The stored TriggerAction 
set should be in the list and named `tutorial1`.

**Warning: The TriggerAction "load" dialogs allow deleting stored TriggerActions by hitting the red "trash" button
next to each action. Hitting the button permanently deletes the respective TriggerAction set, without further 
confirmation**

At this point we could safely delete our TriggerAction from the "TriggerActions" tab (!!not with the trash button from 
one of the load dialogs!!).

With the TriggerAction deleted from the active ones, nothing happens if we detach and re-attach P4wnP1 from the USB 
host.

Anyways, the stored TriggerAction set `tutorial1` will persists reboots and could be reloaded at anytime. 

Instead of reloading the TriggerAction set from with the webclient, we try to accomplish that using the CLI client.

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

The usage screen shows, that TriggerAction Templates could be deployed with the `-t` flag. We run the following command,
to restore the stored TriggerAction set:

``` 
P4wnP1_cli template deploy -t tutorial1
``` 

The TriggerAction which fires out HIDScript on USB host connections is now loaded again and should be shown in the 
TriggerActions tab of the webclient. If P4wnP1 A.L.O.A. is attached to an USB host, the script should run again.

Storing, loading and deploying of templates is one of the two main concepts behind P4wnP1's automation workflow, 
the other one are the already known TriggerActions. It is worth mentioning, that not only TriggerAction sets could be
stored and loaded as templates themselves, but that TriggerActions could be used to deploy already stored templates, if 
that makes sense.

Revisiting our tasks, it seems all defined requirements are met now:
- we typed "Hello world" into the editor of a Windows USB host 
- the editor is opened by P4wnP1, not manually by the user
- the editor is closed automatically, when one of the keyboard LEDs toggled once
- every time P4wnP1 is attached to a USB host, this behavior repeats
- the HIDScript runs only once, unless P4wnP1 is re-attached to the USB host, even if successive keyboard LED changes
occur
- if P4wnP1 is rebooted, the same behavior could be recovered by loading the stored TriggerAction set (which again 
refers to the stored HIDScript). This could either be achieved with a single CLI command or with a simple "load&add" or
"load&replace" from the webclient's trigger action tab.

Once more let us add additional goals:
- it should be assured, that the USB configuration has the keyboard functionality enabled (the current setup doesn't do
this and the TriggerAction couldn't start the HIDScript in case the USB keyboard is disabled)
- the created setup should applied at boot of P4wnP1 A.L.O.A., without the need of manually loading of the TriggerAction 
set. The setup has to survive a reboot of P4wnP1.

To achieve the two additional goals, we have to dive into a new topic and ...

#### Introduce Master Templates and Startup Master Template

Before we look into Master Templates, we do something we haven't done, yet, because everything just worked as intended 
so far: We define a valid USB configurations, matching our task!

- device serial number: 123456789
- device product name: Auto Writer
- device manufacturer: The Creator
- Product ID: 0x9876
- Vendor ID: 0x1D6B
- enabled USB functions
  - HID keyboard
  - HID mouse
  
Let's take a look into the usage screen of the CLI command, which could bes used to deploy these settings, first:

``` 
root@kali:~# P4wnP1_cli usb set -h
set USB Gadget settings

Usage:
  P4wnP1_cli usb set [flags]

Flags:
  -e, --cdc-ecm               Use the CDC ECM gadget function
  -n, --disable               If this flag is set, the gadget stays inactive after deployment (not bound to UDC)
  -h, --help                  help for set
  -k, --hid-keyboard          Use the HID KEYBOARD gadget function
  -m, --hid-mouse             Use the HID MOUSE gadget function
  -g, --hid-raw               Use the HID RAW gadget function
  -f, --manufacturer string   Manufacturer string (default "MaMe82")
  -p, --pid string            Product ID (format '0x1347') (default "0x1347")
  -o, --product string        Product name string (default "P4wnP1 by MaMe82")
  -r, --rndis                 Use the RNDIS gadget function
  -s, --serial                Use the SERIAL gadget function
  -x, --sn string             Serial number (alpha numeric) (default "deadbeef1337")
  -u, --ums                   Use the USB Mass Storage gadget function
      --ums-cdrom             If this flag is set, UMS emulates a CD-Rom instead of a flashdrive (ignored, if UMS disabled)
      --ums-file string       Path to the image or block device backing UMS (ignored, if UMS disabled)
  -v, --vid string            Vendor ID (format '0x1d6b') (default "0x1d6b")

Global Flags:
      --host string   The host with the listening P4wnP1 RPC server (default "localhost")
      --json          Output results as JSON if applicable
      --port string   The port on which the P4wnP1 RPC server is listening (default "50051")
``` 

The command has a bunch of flags, but there exists a bunch of changeable USB settings, too. 
Deploying our defined USB setup could be done like this, using the CLI:

```
root@kali:~# P4wnP1_cli usb set \
> --sn 123456789 \
> --product "Auto Writer" \
> --manufacturer "The Creator" \
> --pid "0x9876" \
> --vid "0x1d6b" \
> --hid-keyboard \
> --hid-mouse
Successfully deployed USB gadget settings
Enabled:      true
Product:      Auto Writer
Manufacturer: The Creator
Serialnumber: 123456789
PID:          0x9876
VID:          0x1d6b

Functions:
    RNDIS:        false
    CDC ECM:      false
    Serial:       false
    HID Mouse:    true
    HID Keyboard: true
    HID Generic:  false
    Mass Storage: false
```

The result output of the (longish) command shows the resulting USB settings. Let us check the "USB settings" tab of the 
webclient to confirm that they have been applied. All changes should be reflected, if nothing went wrong.

Although it is perfectly possible to deploy a USB setup using the CLI, there are several benefits using the webclient 
in favor of the CLI. In this case:
- changing the settings from the webclient is easier and more convenient
- the webclient holds an internal settings state, this allows defining USB settings without actually deploying them (the
CLI on the other hand, could only manipulate settings by deploying them. This, again, resets the while USB stack of 
P4wnP1 and all dependent functionality. F.e. already running HIDScript would be interrupted or USB network interfaces 
are redeployed 
- the current settings of the webclient could be stored to a persistent template, without deploying them upfront
- the CLI client, (currently) isn't able to store USB settings

In our current case, it is obviously a better choice to use the webclient for the needed changes to the USB settings. 
The good thing about the CLI approach (we already used here): As the CLI forced us to deploy the USB settings, we could
confirm that they are working, before we store them into a persistent template.

Let's go on with storing the USB settings:

Again we hit the "store" button, this time in the "USB settings" tab. Once more we call the template `tutorial1` (there
is no conflict with the TriggerAction template stored under the same name, because a different namespace is used for 
USB settings).


Now we have two new and persistently stored templates::
1) a template for the TriggerAction set, named `tutorial1`
2) a template for the USB settings, also name `tutorial1`

Assuming the state (of current USB settings, TriggerActions or both) changed somehow, we could reload both of the 
stored settings at once, by issuing the following CLI command:

```
P4wnP1_cli template deploy --usb tutorial1 --trigger-actions tutorial1
```

The command `P4wnP1 template deploy` could load a template for each of the sub systems of P4wnP1 A.L.O.A. in a single
run (for the network subsystem multiple templates could be loaded, one per each adapter). Deploying templates for
various subsystems is considered a common task while working with P4wnP1 A.L.O.A., because in most cases it should be 
necessary to reconfigure several subsystems to reach a single goal. To account for this, so called *Master Templates*
have been introduced.

A Master Template could consist of:
- a already stored TriggerAction set template
- a already stored USB settings template
- a already stored WiFi settings template
- a already stored Bluetooth settings template
- multiple stored Network settings templates (one per each adapter)

A Master Templated could be defined, stored or loaded, using the "Master Template Editor" from the "Generic Settings"
tab of the webclient. Using the webclient is a convenient way to define Master Templates, as it supports you by only
allowing to select templates, which have been already stored for the respective sub systems (and currently the 
webclient is the only way to define Master Templates).

So lets define a Master Template for our current task:
1) Navigate to the "Generic Settings" tab of the webclient
2) On the "Master Template Editor" click on the small button on the right to the "TriggerActions Template" field
3) From the Dialog choose the `tutorial1` template and confirm with "OK" button
4) If you selected the wrong template, re-open the dialog and select a different one or use the "x" icon on the right 
to "TriggerAction Template" to delete the current selection
5) Repeat the steps for the "USB Template" selection, again choose `tutorial1` (which is a different template for
the USB sub system, although it shares the name with the one for the TriggerActions)
6) Check that the correct templates have been selected for bot, USB and TriggerActions, and all other Templates are left 
empty
7) Store the new Master Template, by hitting the "Store" button and providing the name `tutorial1`

To confirm if that the template has been stored, you could use the "Load Stored" button - the template should be listed
in the selection. Cancel the "Load Store" dialog again.

Now hit the "Deploy Stored" button, select the template named `startup` and confirm with "OK".

In contrast to the "Load Stored" function, which loads a stored template to the Master Template Editor, the 
"Deploy Stored" function applies all settings of a Master Template to the corresponding sub systems of P4wnP1, 
immediately (without even loading them to the Master Template Editor).

As the `startup` Master Template overwrites the current WiFi settings, it could happen that you have lost the connection
to the webclient and need to reconnect to the P4wnP1 WiFi network.

Once you've reconnected successfully and inspect the current USB settings and current TriggerActions, the settings we
stored earlier have been overwritten by the sub settings of the `startup` Master Template.

There are two ways to deploy the `tutorial1` Master Template again:
1) Deploying it using the "Deploy Stored" dialog from the "Master Template Editor" (as done with the `startup` Master 
Template a minute, ago)   
2) Deploying it using the CLI client with `P4wnP1_cli template deploy --full tutorial1` (the `--full` flag is an alias 
for Master Template)

Beeing able to deploy the Master Template `tutorial1`, we already achieved one of our new goals:
 
It is assured, that the USB configuration has the keyboard functionality enabled when we load our keystroke injection
setup. 

A quick summary on how this works:
- the Master Template `tutorial1` loads USB settings, called `tutorial1` which have
  - USB keyboard and USB mouse enabled
- the Master Template `tutorial1` loads a TriggerAction set with a single TriggerAction
  - the TriggerAction starts the HID script `tutorial1.js` each time P4wnP1 gets attached to an USB host
    - the HIDScript starts typing, once the `waitLED` trigger fires (keyboard driver ready) and ends after a successive
    LED change 
 

The only remaining goal is the following: The created setup should applied at boot of P4wnP1 A.L.O.A., without the need 
of manually loading of the TriggerAction set. The setup has to survive a reboot of P4wnP1.

This goal could be achieved pretty easy now. The "Generic Settings" tab of the webclient presents a card called
*Startup Master Template*. Changing the Startup Master Template to `tutorial1` at this point would have immediate effect
and it would likely *destroy the working boot configuration of P4wnP1 A.L.O.A.". 


**Important: If a Master Template has sub templates left empty (f.e. if there is no Bluetooth template selected), the 
respective sub system isn't reconfigured when the Master Template is loaded. While this comes in handy for runtime 
reconfiguration without resetting already running sub systems like the USB stack or WiFi stack if not needed, Master 
Templates used as Startup Master Template leave sub systems without defined templates in an UNDEFINED STATE. If, for 
example, no valid WiFi template is provided, it is unlikely that P4wnP1 A.L.O.A. will be reachable via WiFi after 
reboot**

So before deploying our new `tutorial1` Master Template as Startup Master Template, we assure proper settings are loaded 
for the other subsystem. We do that like this:

1) From the "Master Template Editor" hit the "Load Stored" button and reload the `tutorial1` template to the editor.
2) The template should have `tutorial1` set for "TriggerActions Template" and for "USB template"
3) For "WiFi Template" select the template named `startup` 
4) For "Bluetooth Template" select the template named `startup` 
5) For "Network Templates" select the templates named:
    1) `bteth_startup` 
    2) `usbeth_startup`
    3) `wlan0_startup_dhcp_server`
6) Overwrite the `tutorial1` Master Template with the new settings (hit "Store", enter `tutorial1` and confirm with 
"OK")
7) Double check that the changes have applied, by hitting "Load Stored", again, and selecting `tutorial1`. All 
subsections of the loaded Master Template should look like described here.

Now we are ready to deploy our new Master Template as Startup Master Template. After doing so we hit the "reboot" 
button.

Once rebooted, P4wnP1 A.L.O.A. should trigger the HIDScript automatically (and should still be reachable via WiFi, 
to allow reconfiguration)

**Congratulations, all goals achieved**

You have learned about the very basic workflow concepts of P4wnP1 A.L.O.A.

## 3. Where to go from here

It is currently not possible to provide a full documentations. So here are some comments on topics which haven't been 
touched, yet, but are worth investigating.

### BashScripts

P4wnP1 allows running BashScripts from TriggerActions. The scripts which are usable from TriggerActions are homed at
`/usr/local/P4wnP1/scripts`. If a script is called from a TriggerAction, several arguments (like the actual trigger) are
handed in via bash variables. The file `/usr/local/P4wnP1/scripts/trigger-aware.sh` provides a nice example of a bash
script which acts differently, depending on the calling trigger. It is worth having a look onto this script, as it uses
all "TriggerAction variables" currently available.

### GPIO

The community of the old P4wnP1 version occasionally came up with hardware mods or extensions of the Rapsberry PI and
the question how to integrate them. It isn't possible for me to provide a generic solution to this problem. Neither
is it a good idea, to provide support for a very specific hardware extension, which is only used by a few people. With
the introduction of TriggerAction the idea came up, to support GPIO as both, Triggers via GPIO input and Actions issuing
GPIO output. Although not planned for the first release, this feature has already been implemented. I still haven't
had the time to document it and it could easily happen that some things change. The functionality uses the "periph.io"
library with some minor extension (customized edge detection with custom de-bounce for GPIO, thanks to @marcaruel for 
exchange on this) 

### nexmon KARMA 

The WiFi firmware included with P4wnP1 A.L.O.A. has been modified (utilizing nexmon framework) to support KARMA.
This feature hasn't made it into the core so far (needs some rework firmware-wise) and thus isn't available from 
webclient or CLI. If you want to play around with the karma features, there is a legacy python CLI, which allows setting
the KARMA options on the fly. The python script could be found here:
`/usr/local/P4wnP1/legacy/karmatool.py`

Tip: To get most out of the KARMA functionality, you should setup P4wnP1 A.L.O.A. to provide a WiFi Access Point without
authentication, otherwise it wouldn't make to much sense. For poor beacon flooding this isn't needed, but (static) 
custom SSIDs for beaconing are limited in their number (saving resources on the WiFi chip)

Help screen of karmatool.py:

```
root@kali:/usr/local/P4wnP1/legacy# ./karmatool.py 
Firmware in use seems to be KARMA capable
Firmware configuration tool for KARMA modified nexmon WiFi firmware on Pi0W/Pi3 by MaMe82
=========================================================================================

RePo:       https://github.com/mame82/P4wnP1_nexmon_additions
Creds to:   seemoo-lab for "NEXMON" project

A hostapd based Access Point should be up and running, when using this tool
(see the README for details).
            
Usage:      python karmatool.py [Arguments]

Arguments:
   -h                   Print this help screen
   -i                   Interactive mode
   -d                   Load default configuration (KARMA on, KARMA beaconing off, 
                        beaconing for 13 common SSIDs on, custom SSIDs never expire)
   -c                   Print current KARMA firmware configuration
   -p 0/1               Disable/Enable KARMA probe responses
   -a 0/1               Disable/Enable KARMA association responses
   -k 0/1               Disable/Enable KARMA association responses and probe responses
                        (overrides -p and -a)
   -b 0/1               Disable/Enable KARMA beaconing (broadcasts up to 20 SSIDs
                        spotted in probe requests as beacon)
   -s 0/1               Disable/Enable custom SSID beaconing (broadcasts up to 20 SSIDs
                        which have been added by the user with '--addssid=' when enabled)
   --addssid="test"     Add SSID "test" to custom SSID list (max 20 SSIDs)
   --remssid="test"     Remove SSID "test" from custom SSID list
   --clearssids         Clear list of custom SSIDs
   --clearkarma         Clear list of karma SSIDs (only influences beaconing, not probes)
   --autoremkarma=600   Auto remove KARMA SSIDs from beaconing list after sending 600 beacons
                        without receiving an association (about 60 seconds, 0 = beacon forever)
   --autoremcustom=3000    Auto remove custom SSIDs from beaconing list after sending 3000
                        beacons without receiving an association (about 5 minutes, 0 = beacon
                        forever)
   
Example:
   python karmatool.py -k 1 -b 0    Enables KARMA (probe and association responses)
                                    But sends no beacons for SSIDs from received probes
   python karmatool.py -k 1 -b 0    Enables KARMA (probe and association responses)
                                    and sends beacons for SSIDs from received probes
                                    (max 20 SSIDs, if autoremove isn't enabled)
   
   python karmatool.py --addssid="test 1" --addssid="test 2" -s 1
                                    Add SSID "test 1" and "test 2" and enable beaconing for
                                    custom SSIDs
```
 

### WiFi covert channel

The WiFi covert channel hasn't been ported to Go and isn't part of the P4wnP1 core. Anyways, the legacy functionality
is provided. In order to make the covert channel run, several conditions have to be met:
- a keystroke injection has to be applied to the target client to inject stage1
- stage1 loads stage2 over a (simplified version) of the HID covert channel, thus a special USB HID device has to be 
provided and a special HID covert channel server has to be started on P4wnP1, in order to provide stage2
- a second server has to be started and interface the modified WiFi firmware, in order to manage client connecting
in via the WiFi covert channel and provide interactive shell access to those clients (the server is a console 
application which is meant to run in a terminal multiplexer, like `screen`)

All of the aforementioned conditions could be fulfilled using P4wnP1 A.L.O.A.'s feature set, if the needed components 
(HID stager, WiFi cover channel server, client agent to deliver) are provided.

To carry out such a task with P4wnP1 A.L.O.A. is a great example for its capabilities. Additionally it helps to
distinguish what P4wnP1 A.L.O.A. is meant to be and what is not meant to be.

P4wnP1 A.L.O.A. is not meant to:
- be a "weaponized" tool
- provide RTR payloads, which could be carried out by everybody, without understanding what's going on or which risks
are involved

P4wnP1 A.L.O.A. is meant to:
- be a flexible, low-cost, pocket sized platform
- serve as enabler for tasks like the one described here
- support prototyping, testing and carrying out all kinds of USB related tasks, commonly used during pentest or
redteam engagements, without providing a finalized static solution

In set sense, the `/usr/local/P4wnP1/legacy` folder homes the needed external tools to run the WiFi covert channel
(namely the WiFi server, the HID covert channel stager server and the WiFi covert channel client agent). Those 
components could be considered as external parts (don't belong to P4wnP1 A.L.O.A. core).

Additionally P4wnP1 A.L.O.A. provides a configuration, which utilizes the given components to do the following things:
- drive-by against Windows hosts in order to deliver in-memory client code to download stage2 via HID covert channel, 
based on keystroke injection (HIDScript)
- starting the keystroke injection, as soon as P4wnP1 is connected to a USB host (TriggerAction issuing HIDScript)
- bring up the stager, which delivers the WiFi covert channel client agent via HID covert channel, as soon as the
keystroke injection starts (TriggerAction running a bash script, which again starts the external server)
- bring up the WiFi covert channel server, when needed (same TriggerAction and BashScript)
- deploy a USB setup which provides a USB keyboard (to allow keystroke injection) and an additional raw HID device 
(serves as covert channel for stage2 delivery) - the USB settings are stored in a settings template
- deploy a WiFi setup, which allows remote access to P4wnP1, in order to allow interaction with the CLI frontend of
the WiFi covert channel server - the WiFi settings are stored in a settings template
- provide a single point of entry, to deploy all the needed configurations at once (done by a Master Template, which
consists of proper WiFi settings, proper USB settings and the TriggerActions needed to start the HIDScript)

The Master Template is called "wifi covert channel". By deploying it from the "generic settings" tap of the webclient
("DEPLOY STORED" from the Master Template Editor) P4wnP1 A.L.O.A. is ready configured to execute all the describe steps.

As soon as it is re-attached to a USB host, it should start typing out stage1 and the according servers are started 
internally.
From a SSH session (for example over WiFi) the WiFi covert channel server could be accessed using `screen -d -r wifi_c2`
to interact with clients, which connected back over the WiFi covert channel.
As the keystroke injection depends on the USB hosts language layout, the according HIDScript called 
`wifi_covert_channel.js` has a variable `language` which could be used to adjust the keyboard layout in use. 
Additionally there is a variable called `hide` (false by default). If `hide` is set to true, the console Window on
the client gets hidden while stage1 is typed. This, again, pinpoints how complex tasks could be reduced to a simple bool
variable, thanks to HIDScript and the backing JavaScript engine. 

The "wifi covert channel" demo provided with P4wnP1's Master Templates could be used as Startup Master Template, too, as
WiFi access is still possible and thus the setup could be changed again, remotely at any time. 

The involved BashScript, which is called from a TriggerAction is a good example how flexible the CLI client could get.
As the HID stager needs to know on which device file to listen (the one which represents the generic HID device), but
this information is only available at runtime (depends on enabled USB gadget functions), the script requests the CLI
to report the correct HID device by running `hidraw=$(P4wnP1_cli usb get device raw)`.

The full BashScript is hosted in the folder `/usr/local/P4wnP1/scripts`, as all bash scripts which should be accessible
from TriggerActions. 

### Bluetooth NAP

P4wnP1 provides Bluetooth based network functionality over the Bluetooth Network Encapsulation Protocol (BNEP).
The currently most interesting feature is the Bluetooth Network Access Point (NAP), which allows IP based Bluetooth 
remote access to P4wnP1, for example from mobiles.

In order to use this feature some things should be known:
- The bluetooth network interface, called `bteth` could be configured and templated like the other network interfaces 
(webclient or CLI)
- In order to allow NAP access from an Android mobile (iPhone untested), the mobile not only needs to connect, but 
additionally P4wnP1 has to hand out a proper IP for the default gateway on the `bteth` interface via DHCP. This is, 
because the mobile wants to use the NAP as gateway to the Internet (which would be the intended use). If the NAP itself
wouldn't provide a gateway, the Android Mobile would do no further requests after the DHCP D.O.R.A. The easiest way to
overcome this, is to instruct the DHCP server to provide the IP of the `bteth` interface itself as default gateway 
(DHCP option 3). Even if there is no real upstream connection, this worked during my tests - as the mobile has to access
the gateway with layer3 communication in order to "phone home". Even if successive connectivity tests fail, the working
layer 3 connection persists. This allows, for example, SSH access via Bluetooth. With "High Speed" enabled, the 
webclient works pretty nice, too.
- In order to allow PIN based pairing Simple Secure Pairing (SSP) has to be disabled. If SSP is enabled, the running
Pairing agent confirms every passkey (which means even less security than with legacy PIN pairing, as every device is 
able to connect). Maybe a confirmation dialog for SSP based passkey pairing will be implemented for the CLI/webclient
in future, but currently this is out of scope. I highly suggest to disable "discoverable" and "bondable" if SSP is in
use, as soon as the intended device has paired.
- Another shortcoming of having SSP disabled, is that "High Speed" wouldn't be usable for bluetooth conections (or
to enable High Speed, pairing has to be done with SSP). Without "High Speed" enabled (uses 802.11 frames for 
communication) it would take about 10 minutes to request the webclient, with high speed enabled it takes some seconds.
Using SSH and the CLI client over a NAP without "High Speed" should be fine, though.
- The default Bluetooth network interface settings (`bteth_startup`) and the default Bluetooth settings (`startup`) 
should allow "Low Speed" access over SSH with legacy PIN pairing. The PIN is `1337` and could be changed from the 
webclient. 

### TriggerAction Groups

The TriggerActions come with a nice rooting capability called "Groups". I didn't managed to come up with a feature demo
in time, but I'm planning to include an example for an LED based 4 bit binary counter (using GPIOs, a toggle switch
and 4 LEDs).

The idea of groups is the following:

Consider you want to have 4 TriggerActions (TAs) firing on the exact same Trigger (for example "on attached to USB host").
You could achieve this by creating 4 TAs, each with the Trigger "on attached to USB host".

Alternatively, you could bring up a TriggerAction which sends the value `1` to a group named `"connected"` when the 
"on attached to USB host" occurs. Now you define your other 4 TriggerActions to fire when the value `1` is received on
a group named `"connected"`. The result would be the same and not make too much sense for now (in fact it needs one 
more TriggerAction). The only positive effect, for now, is that the TriggerActions are slightly more readable, thanks
to the group name, which could be chosen freely.

Now the first advanced thing you could do, is to run the following CLI command:

```
P4wnP1_cli trigger send --group-name=connected --group-value=1
```

This command would have the exatc same effect as the "on attached to USB host" host TriggerAction and all 4 other
TAs, which are waiting for the value `1` to arrive on the group `connected` would fire. As you maybe remember, the
CLI client could run remotely (from different platforms), so it could be used to Trigger command remotely.

The Trigger which reacts on "Group channels" is called "value on group channel". The more interesting trigger is called
"multiple values on group channel". This "multiple values" trigger allows to listen for ordered sequences of values, or 
one of multiple values or all values in an unordered sequence, before it fires.

Let's say you qant to fire a BashScript, when these conditions are met:
- The WiFi AP is up
- P4wnP1 has been connected to a USB host

You could create TAs for both events like this:
1) On "WiFi AP up" --> send value 1 to group "conditions"
2) On "attached to USB host" --> send value 2 to group "conditions"

Now you could deploy a third TriggerAction like this
- On "multiple values on group channel"; values (1,2); type "All (logical AND)" --> start bash script

In this configuration, the bash script would only start if both "condition" Trigger have fired.

If "exact ordered sequence" would have been used, instead of "All (logical AND)" for the type, the bash script would
only start if the WiFi AP come up before the USB connected Trigger (not the otherway around). In combination with GPIO
triggers, this could for example be used, to trigger actions based on the input of a simple PIN pad.

I'm sure you have some nice usage ideas for "group" channels.

Worth mentioning:

The CLI client is able to do a blocking wait, till a dedicated value arrives on a "group channel", using a command like
this:

```
P4wnP1_cli trigger wait --group-name=waitgroup --group-value=1
```

This could be used, to drive scripts from TriggerActions, utilizing the CLI (with their full power like GPIO).


Work in progress, missing sections:
- HIDScript Trigger variables (variables handed in to HIDScripts fired from TriggerActions)
- HIDScript helpers (powershell functions)
- HIDScript demo snake (mouse)
- USB Mass storage (genimg helper)

## 4. Rescue: Help, I can't reach P4wnP1 A.L.O.A. as i messed up the configuration
 
P4wnP1 A.L.O.A. doesn't protect you from misconfiguration, which render it unusable (as a root console wouldn't protect
you from running `rm -rf /`).

In case you messed everything up, here some ideas how to fix things:

### Database backup

Before you take critical changes to a still working P4wnP1 configuration, create a database backup. This could either be 
done from the "Generic Settings" tab of the webclient or the the CLI, with the `P4wnP1_cli db backup` command.
The backup will be stored in the folder `/usr/local/P4wnP1/db` under the chosen name.
The "restore" function or `P4wnP1_cli db restore` command could be used to restore a given backup.
A backup contains all stored templates (USB, WiFi, Network, Bluetooth, TriggerActions, MasterTemplates) and the Startup
Master Template which has been set. The backup doesn't include HIDScripts or BashScripts, as both are stored as files
to allow easy editing.

### I have no backup and messed everything

When P4wnP1 A.L.O.A. starts, it checks if a database exists. If the database doesn't exist it fills a new database based
on an initial backup which ships with P4wnP1 A.L.O.A.

The initial backup is stored at `/usr/local/P4wnP1/db/init.db` and **should never be deleted or overwritten**.

In order to force P4wnP1 to re-create the actual database has to be deleted. This could be achieved by mounting the 
P4wnP1 A.L.O.A. SD card on a system which is capable of writing EXT partitions.

Once done, delete the folder `/usr/local/P4wnP1/store` from the SD card's root partition. This deletes the database and
thus forces re-creation once P4wnP1 is booted, again 

### I have a backup, but can't access P4wnP1 to restore it

If you can't restore an existing database, because you have no access, you could still follow the steps from "I have no 
backup and messed everything". In addition to deletion of the `/usr/local/P4wnP1/store` replace the 
`/usr/local/P4wnP1/db/init.db` file with the one from your backup (be sure to have a backup copy of init.db).

This should re-create your custom DB once P4wnP1 is rebooted.

### I messed the Startup Master Template of my backup

If you have a backup for which the Startup Master Template doesn't work. You have to do some additional steps, as it
isn't possible to change the Startup Template directly in a backup.

First follow the steps from "I have no backup and messed everything" which re-create the initial P4wnP1 database.
After a reboot of P4wnP1, you should be able to access P4wnP1's webclient remotely, again.

Move on to the "Generic Settings" and restore your own backup (the one with the wrong Startup Master Template).

The "Startup Master Template" should show your "broken" Master Template as selected. If this isn't the case, reload
the browser tab hosting the webclient application.

Again, navigate to the "Generic Settings" tab and select a Startup Master Template which is known to work.

At this point you should be ready to reboot. 

### none of the above helped

Sorry, seems you have to recreate your P4wnP1 A.L.O.A. SD crad from a clean image.

## 5. Credits

Under construction, random order

- @JohanBrandhorst (close exchange on gRPC-web via gopherjs, ridiculous fast implementation of 
"websocket for server streaming", feature request)
- @steevdave, @_binkybear (kali build scripts, discussion ongoing exchange)
- @Re4sonKernel (Support on moving P4wnP1 kernel changes to a well maintained and popular repo, collaboration on Bluez
fixes)
- @SymbianSyMoh (Inspiration for HID attack re-trigger without reboot)
- @quasarframework (could list this under 3rd party libs, but the work done here is insane; the look&feel of the 
P4wnP1 webclient is more or lees based on default components of this beautiful library)
- @CyberArms (one of the most early P4wnP1 supporters, writer of the best tutorial and even books on such topics)
- @LucaBongiorni (not only one of the earliest supporters, he does in hardware, what I'm only to do in software; he
gives talks on the USB topic and honors Open Source solutions, all in all a great guy and an inspiration)
- @evilsocket (his block pushed me towards Go, a great OSS developer, read his code and you know what I mean)
- @RoganDawes and @Singe from @SensePost (inspiring guys)
- @Swiftb0y (Early supporter, creator of the "old" P4wnP1 WiKi, early tester for ideas on P4wnP1 A.L.O.A.)
- @marcaruel (discussion on GPIO edge detection using periph.io) 

## 6. ToDos and support

This isn't a full fledged ToDo list, but some milestones are left and I'd be happy to receive som community support on
this
- Porting the full HID covert channel functionality to Go core (I'm on my own with that)
- **add Bluetooth configuration command for CLI**
- Create additional keyboard layouts (currently br, de, es, fr, gb, it, ru and us are supported) 
- extend Bluetooth functionality to allow connection to other discoverable devices (authentication and trust)
- move WiFi KARMA functionality from dedicated python tool to P4wnP1 core (with webclient support))
- Create full documentation for HIDScript (basically only the mouse part is missing)
- Create full documentation for P4wnP1 (hoping for community)
- Get rid of a remaining dependency on the docker netlink (see README of the `netlink` folder)

Note on Bluetooth: 

P4wnP1 works with custom bindings to the Bluez API. Although the Bluez API supports Low Energy (GATT,
emulating peripherals etc.) it isn't planned to integrate this functionality into P4wnP1 A.L.O.A. 

Note on Nexmon: 

P4wnP1 utilizes nexmon. Most people know nexmon as a firmware modification which allows enabling monitor
mode and package injection for broadcom WiFi chips (including the BCM43430a1, which is used by the Raspberry Pi Zero W).
But nexmon is more, it is a framework which allows modifying ARM firmware blobs (after a bit of reversing), with patches
written in high level C code. P4wnP1 uses this framework, to apply custom patches to the WiFi firmware, which enable
hardware based KARMA support and firmware (as well as driver) support for the WiFi covert channel. It isn't the aim
of this modifications to provide proper monitor mode or injection support for the built-in WiFi interface. Although, 
the legacy nexmon monitor mode functionality is included in the current WiFi firmware, it is considered "erroneous", as
it interferes with standard WiFi functionality used by P4wnP1 (crashes if the interface is used in station mode etc.). 

## 7. Copyright

    P4wnP1 A.L.O.A.
    Copyright (C) 2018 Marcus Mengs

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
