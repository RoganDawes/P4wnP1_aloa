step = Math.PI / 90.0;
scale = 6.0
alpha = 0.0
horizontal = true;
while(true) {
	v = Math.sin(alpha) * scale;
	alpha += step;
	if (horizontal) moveStepped(v,0);
	else moveStepped(0,v);
	delay(10);
}
