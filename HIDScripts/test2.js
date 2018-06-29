//Endless looping script moving mouse, used to test interrupts and timeouts
while(true) {
	moveStepped(200,0);
	moveStepped(0,-200);
	moveStepped(-200,0);
	moveStepped(0,200);
	delay(1000);
}
