// Creds to "Mohamed A. Baset" @SymbianSyMoh (see https://twitter.com/SymbianSyMoh/status/987140763673706496)

// Endless looping script moves mouse on any new LED report

while(true) {
    // waitLED(ANY_OR_NONE) blocks till any LED of the P4wnP1 keyboard changes.
	// The special flag "ANY_OR_NONE" additionally triggers, if a new LED state arrives, which doesn't differ from the
	// old state, at all.
	// On Windows and some Linux OS, all attached keyboards share a global LED state. In order to show the correct state
	// on a newly attached keyboard, it has to be sent from the windows host to the external keyboard, at least once.
	// As P4wnP1 saves the LED state internally in order to detect changes, it could happen, that a newly received state
	// doesn't differ from the internal one (it would be ignored by waitLED(ANY) for example).
	// `waitLED(ANY_OR_NONE)` assures that unchanged states are reported, too. This again allows to trigger the
	// fellow commands, as soon as P4wnP1 is attached to a windows host.
	//
	// It seems that this technique couldn't be used on OSX, as users reported, that the LED state on OSX is handled
	// per keyboard, not globally.
 	waitLED(ANY_OR_NONE); // wait for new LED report, even if there's no change

	// move mouse to indicate success
	moveStepped(200,0);
	moveStepped(0,-200);
	moveStepped(-200,0);
	moveStepped(0,200);
	delay(1000);
}
