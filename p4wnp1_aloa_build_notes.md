# P4wnP1 ALOA build components

## Note on possible issues for ownership transfer

Majority of code is written in code and there are a whole lot of dependencies.
Some of the dependencies are covered using Go modules. More complicated ones rely
on custom code, written by myself and residing in various repos. Go imports code
dependencies based on the path to the source (essentially github repo URLs).
With transfer of ownership these import path's change. Although the [Github instructions
for transferring repositories](https://help.github.com/en/github/administering-a-repository/transferring-a-repository) explain that old URLs end up in redirects to the transferred
repos, this could get an issue. This is especially true for code which can't utilize
the dependency tracking of Go modules (GopherJS based WebClient).

In summary: With ownership transfer there's a risk to break everything, but pre-built
release version will survive in any case.

## ALOA Components / IPC / inner design

P4wnP1 consists of a backend service which is meant to be ran as systemd service
and frontend client(s).
The idea was to keep the backend service running at all time, to keep track
of all runtime states, so that the frontend client work stateless.

The simplest real-world example leading to this requirement, is controlling of the LED.

The user should be able to change the "blink count" of the LED at runtime on demand.
Blinking the LED by a fixed count with a longer break between each blink sequence
has to be done by an **endless running loop** in software (the driver has no native
support for such a blink mode). The user - on the other hand - should be able to change
this blink count on-demand with a **non-blocking command** and shouldn't be required to
know the current blink count. This leads to the following requirements:

1. A callable non-blocking user client interface to alter runtime state of P4wnP1 (blink
   count in this case)
2. A constantly running background service which is able to constantly run multi-threaded
   code (LED blink loop in this case) and keeping track of current runtime state (blink count
   value used by the blink loop in this case)
3. The client could (and should) be stateless, as all runtime states could be fetched from
   the background service if required.
4. Communication between the user client and background service requires an IPC mechanism
   which is fast, able to cross privilege boundaries (privileged background service, unprivileged
   client).

Several IPC mechanisms have been evaluated when starting the project. Surprisingly RPC
implementations didn't have much performance impact (compared to pure system local IPC methods)
for P4wnP1 ALOA's use case, while giving the advantage to allow the user client not only to
work stateless, but also remotely. `Go` (version 1.10 when the project was started) comes with
a nice RPC implementation built in, but Google also provides GRPC, which has some additional advantages.

Advantages of GRPC:

- As this is RPC, clients could be used remotely free of costs (f.e. the CLI client could easily
  be compiled for x86/AMD64 Windows and reach out to the backend service from a different host)
- Could implement an event driven approach with "server streams", where the client keeps
  open a RPC connection, while the server sends messages "as they occur" and the client could
  process them "as they arrive" (in chunks of the stream)
- A GRPC service definition (based on protobuf) is well defined and compiles into interface code
  for all possible languages with respective compilers. This to drive the API design from a single
  service definition file, which only needs to be recompiled to implement new API functionality
  (or leave the respective calls unimplemented).
- Very "thin" wire format (binary encoding without metadata, default values are transmitted at all).
  The programming language specific GRPC interfaces (compiled from the `proto/grpc.proto` service
  definition) take care of Object (de-)serialization and provide interface methods for the API calls, which only require implementation on server (== service) end.

Note: The GRPC approach isn't meant to cross boundaries for kernel level code. The backend service
needs to implement additional IPC mechanisms to achieve this (DBUS, netlink etc).

The benefits of GRPC made it possible to implement some additional ideas:

- a GRPC based webclient (receives push notifications based on aforementioned server streams to
  allow event driven UI updates in realtime)
- event driven behavior of the CLI client

The WebUI API interface is handled in a dedicated section, but an example should be given on how
server streaming (push events) is used by the CLI client:

_P4wnP1 ALOA is using another event driven system which allow the user to do some automation. This system
is called `TriggerActions`. One possible Trigger could arise from user configurable GPIO input. The
respective Action (an Action is similar to a configurable event handler) could send a value to a so-called "group channel" (a GroupChannel is very simple implementation of an event bus, which transmits arbitrary
integer values on named channels), where the value send to the channel could represent the GPIO which
triggered the Action (requires one TriggerAction per GPIO, sending to the same channel name). Now let's say
the CLI shall be used in a bash script and block execution till the value `1` arrives on a group channel
named `GPIO_TRIGGER`. The respective command `P4wnP1_cli trigger wait --group-name "GPIO_TRIGGER" --group-value 1`. For most RPC implementations, the CLI client would have to poll the server for the runtime
state og GPIO 1, using a RPC call like `get_gpio()`. This type of busy waiting wastes CPU and network resources. With GRPC server streaming on the other hand, the CLI client starts a call into the RPC server
which receives an endless stream of events, where each fragment transmitted within the stream can be
immediately be processed by the GRPC client. In this special case, each arriving fragment represents an event. The CLI client essentially "wakes up" whenever an event arrives. In case the event matches the filter
rule (group channel value 1 on channel "GPIO_TRIGGER") the client aborts the RPC call and returns execution._

The "event push mechanism" described above is also used by the WebClient, to do event based live updates
for the reactive WebUI.

## Implementation notes on WebClient

For P4wnP1 there was a high user demand for a WebClient. With GRPC in place, I thought it could be easy
to do a neat implementation. To teach myself something new and keep the code base simple, i decided to
write the entire WebUI in `Go` (the CLI client and backend service are already Go based). Which seems to be
a good idea in the beginning turned out to be a real challenge. At the time I started with ALOA
it seems to be a good choice to master this challenge. Looking back it wasn't a good idea, especially if
others want to modify and adopt the code there are several hurdles.

Here is why:

- At the time of the project start, Go supported cross-compilation for a ton of platforms and architectures.
  Unfortunately this didn't include Browsers (with JavaScript). Anyways, a project called `GopherJS` came to
  rescue, which allowed to compile Go code into JavaScript blobs, which could be called from inside a simple
  HTML page. Problem solved!
- It was pretty to compile a Go program for the Browser, let it do some calculations and output the results
  to the internal JavaScript console. Rendering output to the DOM of the WebPage required additional efforts
  and utilized non-standard go bindings, which are only available for GopherJS. This means the WebClient code
  had to be completely separated from the CLI client code, although they share the same low level logic (code
  redundancy)
- Things got even worse when it comes to GRPC, which was supposed to be the "magic tool" which allows an easy
  Browser-to-backend interface. This didn't hold true at the time ALOA development started. There was no GRPC
  support for GopherJS. In fact there was no GRPC support for browser-JavaScript from Google, at all. This gap
  was closed by a company called "Improbable". Improbable came up with a [grpc-web](https://github.com/improbable-eng/grpc-web/) implementation, which brought GRPC support to browsers.
  This was achieved using a JavaScript client library and a wrapper proxy around the GRPC server, which
  allows usage of HTTP and/or WebSockets for GRPC transport (real GRPC uses HTTP 2.0 for transport, which
  ironically wasn't supported by browsers). Still there was a lack of GopherJS support. In practice this means
  binding for the Impropable grpc-web implementation are required (otherwise the JavaScript objects and
  methods couldn't be used from GopherJS Go code). Also, the grpc-web implementation didn't support server
  side streaming, which is utilized by P4wnP1's event system to push data to the client. Luckily, at exact the
  time I was looking into this, Johan Brandhorst came up with [GopherJS bindings](https://github.com/johanbrandhorst/protobuf/tree/master/grpcweb)
  for Improbable's GRPC-web. In addition Johan accepted a commit, which allows the user to enforce using
  WebSockets instead of HTTP as transport, which ultimately brought back server streaming.
- I already used terms like "challenge" and "worse" but this isn't even half of the story. At this point I had
  external libraries which allowed me to use a "single source of truth" for API definition and automated code
  generation for the server RPC method stubs, CLI client API calls and browser-JavaScript API calls - and all of this in a single programming language: Go. Unfortunately the compiled Go client code for the CLI client (written in pure Go) and the WebClient (also written in Go, but compiled with GopherJS) isn't the same. This
  is mainly because the GopherJS interface needs to be aware of the GRPC-web implementation (the generated
  GRPC code for pure Go resides in `proto/grpc.pb.go` while the GopherJS version could be found in `proto/gopherjs/grpc.pb.gopherjs.go`).
  Otherwise it was possible to call P4wnP1's RPC API method from the browser and easily add and modify the
  API by changing the `proto/grpc.proto` definition and recompiling the respective GRPC libraries. Also objects
  are transparently (de-)serialized from/to strongly typed Go objects for the code parts running in browser,
  with all the benefits like well defined structs. The real issue started, when I tried to render this data
  into a website. As mentioned, GopherJS provides interfaces to manipulate the DOM, but building a fancy WebUI on top of that is close to impossible. Typically some JavaScript framework is used to render nice looking and
  **reactive** WebPages and for P4wnP1 ALOA I opted to use `VueJS`. To make a long story short: every library
  which could be used to build a fancy WebPage is based on JavaScript. This again means, there have to be
  bindings for GopherJS to call into this libraries **and the nice strictly typed Go objects have to be converted
  back to JavaScript in order to hand them over to such a framework.** That's not all, if the - now JavaScript objects -
  are manipulated by the WebPage (think of network settings for example), they have to be traversed
  back to Go objects after manipulation, in order to use them as arguments for the GRPC-web calls in GopherJS
  code. As this would still have been to easy, I decided to write no single line of JavaScript, which means
  object transformation from Go to JavaScript and back are all handled in Go(pherJS) code. This complicates
  things even more, because:

  - The browser isn't a stateless client anymore, as it needs to hold state copies (for example while the
    user is changing an IP address in network setting, he doesn't want to trigger an RPC call on every typed
    character, but needs to trigger this call when he's done typing. This call eventually updates a complete
    network settings data structure via RPC and has to handle errors, including fallback to previous state
    in that case)
  - The JavaScript objects and the GopherJS internal representation can't share the same data structures.
    One of many reasons for this: The JS objects used by frameworks like VueJS use getters and setters on
    each attribute (to keep track of data changes, which result in DOM changes and thus require re-rendering
    of parts of the WebPage).

    The Go(pherJS) structs, on the other hand, don't use such functionality. Also, converting more complex
    (nested) objects between JavaScript and GopherJS doesn't work without issues. That's why each and every
    relevant object of the auto-generated GopherJS-GRPC-interface needs manually written methods to transform
    it to JavaScript and back (most of this code is in `/web_client/jsDataHandling.go`).

- VueJS is a template based reactive rendering framework. A template defines a **reusable** component, which
  implies that each instance of such a component keeps its own state (for example a toggle switch could be
  designed in a template, but the switch state - on vs. off - would have to exist per each instance). Following
  that model, would have meant that each API relevant Go data structure not only needs to be transformed to
  a JavaScript version and back, it would also have to be passed through to the correct component instance.
  This could be avoided by an additional JS library, which is called `Vuex` or Vue Store. Vuex essentially allows
  to keep the whole logical state of a WebApplication at a single source of truth (one state object so to say)
  which then could be accessed by all components as they have to. This also means, if one component changes
  a state (foor example an IP of the network configuration) all other components using this state are updated
  and re-rendered automatically. In addition, Vuex is intended to be used with Vue. In JavaScript world it requires
  about two line of code to add Vuex functionality to Vue ... in GopherJS world, it requires a whole set of additional
  bindings, to allow calling the Vuex methods from GopherJS. As I hardly tried to stay away from JavaScript (at least
  code wise, as the aforementioned constraints require a deep understanding of all involved JS libraries), I implemented
  Go bindings for all required JS libraries or modified existing ones:

  1. `Vue`: The bindings for Vue are a modified version of `hvue` from HuckRidgesSW and can be found
     [here](github.com/mame82/hvue)
  2. `Vuex`: The bindings for the Vuex Store are written from scratch, but follow the concept of hvue.
     They could be found [here](https://github.com/mame82/mvuex) (the Vuex code of P4wnP1 provides an interface
     to most of the GRPC API calls for JavaScript world and could basically be used by other JS libraries `web_client/mvuexGlobalState.go`)
  3. Simplified version of JavaScript Promises: `web_client/promise.go`
  4. `Quasar`: Quasar is a template library working on top of Vue. It assures proper and consistent look&feel
     across different desktop and mobile browsers and provides a huge library of pre-defined Vue components.
     Only an absolute minimum of binding code has been written in `web_client/quasarHelper.go`. The framework is
     heavily utilized by P4wnP1 ALOA WebUI, but this is mainly done by template definitions in the respective Go
     files for the components.
  5. `Vue Router`: in `web_client/vueRouter.go`

This section introduced some heavy and complicated parts of P4wnP1 ALOA's webclient. Still, everything isn't tied together.

How to re-compile the GRPC interface code is described in the next section. Anyways, this is only required when
the GRPC service description changes. The pre-compiled interfaces are already part of the repo (`proto` directory).

To build the monolithic JavaScript blob from the GopherJS code the following command has to be issued from within
the `./web_client` subdirectory:

```
gopherjs build -o ../build/webapp.js
```

Before beeing able to build, the build toolchain has to be present (GopherJS and Johan Brandhorsts gopherjs-GRPC
plugin). This could be done with the following commands:

```
# Install gopherjs-gRPC plugin by Johan Brandhorst

# go and protoc have to be installed already
# $GOPATH/bin has to be in path

go get -u github.com/gopherjs/gopherjs
go get -u github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs
```

The resulting JavaScript blob alone wouldn't make up a WebPage. Beside a wrapping HTML file,
there are dependencies to external JavaScript libraries (Vue, Vuex, Quasar, CodeMirror, FontAwesome etc).
As the WebPage - or so called Single Page Application - should be self contained, loading JavaScript
libraries from external sources isn't an option. Thus all relevant parts are "vendored" in the `./dist/www/`
of the project. The GopherJS compiled `./build/webapp.js` and `./build/webapp.js.map` files
and the `./dist/www` directory all end up in the same directory on the Raspberry Pi 0, which is `usr/local/P4wnP1/www`.
This is handled by the `installkali` make target in the `Makefile` of the P4wnP1 ALOA repo.

Finally the WebUI needs to be hosted, which is handled by P4wnP1's systemd backend service itself.
The service listens on port 80 for HTTP request. It distinguishes GRPC calls from the WebClient versus
normal HTTP requests to content based on the `Content-Type` of the HTTP request header. In result,
either a component of the WebPage (static file) will be delivered or the request is unwrapped and moved
on to the HTTP server. This is implemented in the `func (srv *server) StartRpcServerAndWeb` function
in `./service/rpc_server.go`.

It should be noted, that the server has no kind of protection. At least, end-to-end encryptions utilizing TLS/SSL
could easily be added, but I never had time to do so.

## GRPC protobuf source components (rebuild only if `proto/grpc.proto`was changed)

Re-build GRPC interfaces for Go core code.

```
protoc -I proto/ proto/grpc.proto --go_out=plugins=grpc:proto
```

Re-build GRPC interphases for GopherJS WebClient code

```
protoc -I proto/ proto/grpc.proto --gopherjs_out=plugins=grpc:proto/gopherjs
```

## Backend service (dependencies / modified kernel modules / communication with Kernel parts)

t.b.d.

## Building everything

For the Makefile of this repo, only the `installkali` target it valid. This target is intended to
be called from a proper Raspberry Pi 0 Kali image (or within a chroot for such an image) and **ONLY
installs the pre-compiled WebClient, CLIClient and backend service.** It also creates a systemd service
running at boot.

With this alone the following dependencies of the backend service wouldn't be met:

- fixed `pi-bluetooth` package (proper BLUEZ stack DBUS integration fo Pi0 chipset)
- Broadcom WiFi kernel module and driver, modified for P4wnP1 (`{kernel source}/drivers/net/wireless/broadcomm/brcm80211/brcmfmac`)
- Pi0 DWC2 USB Controller Kernel module, modified for P4wnP1 (`{kernel source}/drivers/usb/dwc2`)
- various Kali packages

The modified Kernel modules alone are complex to maintain (and of course they depend on the kernel version used by Kali).

It gets even harder if a whole Kali image has to be prepared to met all dependencies.

Now for the modified Kernel modules Re4son for was kind enough to integrate my patches into a dedicated branch of his
[re4son-raspberrypi-linux](https://github.com/Re4son/re4son-raspberrypi-linux/tree/rpi-4.14.80-re4son-p4wnp1) kernel, which is used by the Kali ARM images.

While my own release image used `Kernel 4.14.80`, Re4son also ported forward the modifications up to Kernel `4.14.93`.

This already would allow to compile a Kernel containing drivers and firmware compatible with P4wnP1 ALOA.

To deal with the challenge of preparing a whole Kali image for P4wnP1 integration and ultimately build images with
P4wnP1 pre-installed, I grabbed and modified the [kali-arm-build scripts](https://github.com/mame82/kali-arm-build-scripts/blob/master/rpi0w-nexmon-p4wnp1-aloa.sh) to do the job.

The OffSec guys kindly hosted a Kali image with P4wnP1 A.L.O.A. integration based on those build scripts and ported it
forward to newer Kali releases (with newer Kernels).

Unfortunately the build system changed meanwhile, also stuff was transferred from github to gitlab. Due to time constraints,
it was impossible to me to rework the build scripts, as requested and thus official Kali support disappeared.

Also Re4son stopped porting the modification to newer Kernels.

As all resources are still available online, it should still be possible to compile the release image (old Kernel and Kali version). Porting everything forward to bleeding edge Kali, requires way more effort and testing.

The steps to compile P4wnP1 A.L.O.A. (with modifications):

1. Do required adjustments in code of P4wnP1 service, CLI and WebClient
2. Assure the (update) WebClient, CLI and Backend service reside in the `./build` directory of the ALOA repo
3. Push the changes to the P4wnP1 ALOA github repo **and assign a tag (semantic versioning)**
4. Modify the line of the build script which [clones the P4wnP1 repo into the image](https://github.com/mame82/kali-arm-build-scripts/blob/master/rpi0w-nexmon-p4wnp1-aloa.sh#L223) to use **the new tag**
5. Run the build script

Note: I haven't tested this myself for over a year, now.
