// P4wnP1 HIDScript feature demo (Keyboard, JavaScript objects, Mouse and LED trigger logic)

typingSpeed(0,0); //type as fast as possible
layout("US"); //language layout

//Start paint
waitLEDRepeat(NUM);
press("GUI R");
delay(500);
type("mspaint\n"); //no need to press ENTER, encoded in '\n'
delay(1000);
//Maximize paint
press("GUI UP"); 
//Set canvas size to 1920x1080
press("CTRL E"); //open properties
delay(500);
type("1920");
press("TAB");
type("1080");
press("ENTER");

//create JavasCript player object
var player = {
	speed: 4.0/1920.0,
	dir: 0.0,
	pos: {"x": 0.5, "y":0.5},
	border: {"lx": 0.2, "ly":0.2, "hx": 0.8, "hy": 0.8},
	pressed: false,
	step : function() {
        v = this.getVec();
		this.pos.x += v.x;
		this.pos.y += v.y;
		if (this.pos.x < this.border.lx || 
			this.pos.x > this.border.hx ||
			this.pos.y < this.border.ly ||
			this.pos.y > this.border.hy) 
			return false;
		this.moveMouseToPos();
		return true;
    }
	getVec : function() {
        x = Math.sin(this.dir) * this.speed;
		y = Math.cos(this.dir) * this.speed;
		return {"x": x, "y":y};
    }
	moveMouseToPos : function() {
        moveTo(this.pos.x, this.pos.y)
    }
	toggleButton : function() {
        if (this.pressed) button(BT1);
		else button(BTNONE);
		
		this.pressed = !this.pressed;
    }
};

//draw border (delays only for visual fx)
delay(500);
moveTo(player.border.lx, player.border.ly);
button(BT1);
moveTo(player.border.hx, player.border.ly); delay(500);
moveTo(player.border.hx, player.border.hy); delay(500);
moveTo(player.border.lx, player.border.hy); delay(500);
moveTo(player.border.lx, player.border.ly); 
button(BTNONE);

//game, exits when player position is outside border (absolute mouse positioning)
var turn = Math.PI / 4.0;
while (true) {
	res = waitLED(ANY, 50) //wait 50ms for keyboard LED change
	if (res.TIMEOUT) { //no LED change --> reposition mouse according to player.dir
		if (!player.step()) { button(BTNONE); break; } //abort if player.pos outside player.border
	}
	if (res.CAPS) player.dir -= turn; //CAPSLOCK LED change --> turn player left
	if (res.NUM) player.dir += turn; //NUMLOCK LED change --> turn player right
	if (res.SCROLL) player.toggleButton(); //SCROLLLOCK LED change --> toggle mouse button1 on/off (draw)
}

//draw 100 random lines from center
moveTo(0.5,0.5);
button(BT1);
for (var i = 0; i<100; i++) {
    x = Math.random(); //x, scaled between -128 and 127
    y = Math.random(); //y, scaled between -128 and 127
    moveTo(x,y);
	moveTo(0.5,0.5);
    delay(20); //wait a half a second
}
button(BTNONE);

//End paint (discard changes)
press("ALT F4"); press("TAB"); press("ENTER");