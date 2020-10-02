//Log something to internal console
console.log("HID testscript");

layout("US"); //set US layout

//Natural typing speed (100 ms between keys + additional jitter up to 200 ms)
typingSpeed(100,200);
type("Typing in natural speed");

layout("DE"); //Switching language layout, while script still running

//Fastest typing speed (no delays)
typingSpeed(0,0);
type("Typing fast, including  unicode: üÜöÖäÄ");

//Do some relative mouse movement
for (var i = 0; i<10; i++) {
    x = Math.random() * 256 - 128; //x, scaled between -128 and 127
    y = Math.random() * 256 - 128; //y, scaled between -128 and 127
    move(x,y);
    delay(500); //wait a half a second
}

//Do some relative mouse movement, but divide it into 1 DPI substeps (pixel perfect mouse move, but slow)
for (var i = 0; i<10; i++) {
    x = Math.random() * 256 - 128; //x, scaled between -128 and 127
    y = Math.random() * 256 - 128; //y, scaled between -128 and 127
    moveStepped(x,y);
    delay(500); //wait a half a second
}

//Do some absolute Mouse positioning (not stepped, mouse moves immediately, thus delays are added)
moveTo(0.2,0.2);
delay(1000);
moveTo(0.8,0.2);
delay(1000);
moveTo(0.8,0.8);
delay(1000);
moveTo(0.2,0.8);
delay(1000);

//press button 1, move mouse stepped, release button 1
console.log("Moving mouse with button 1 pressed");
button(BT1);
moveStepped(20,0);
button(BTNONE);
delay(500);

//Click button 2
console.log("Click button 2");
click(BT2);

//Doubleclick button 1
console.log("Double click button 2");
doubleClick(BT1);
