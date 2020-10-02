// Creds to Rogan Dawes, MouseJiggler idea was part of USaBUSe
//
// Moves the mouse slightly from time, to simulate user activity (supress screensaver/lockscreen)
// Meant to run as background job
// With extrem settings, it could be use to prank users

scale = 2.0; //if not stepped, 127 is valid max
sleeptime = 3000;
stepped = true;
while (true) {
	x = (Math.random() * 2.0 - 1.0) * scale;
	y = (Math.random() * 2.0 - 1.0) * scale;
	if (stepped) moveStepped(x,y);
	else move(x,y);
	delay(sleeptime);
}

