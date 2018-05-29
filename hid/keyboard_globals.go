package hid

/*
Keyboard descriptor used


USAGE_PAGE (Generic Desktop)	05 01
USAGE (Keyboard)	09 06
COLLECTION (Application)	A1 01
  USAGE_PAGE (Keyboard)	05 07
  USAGE_MINIMUM (Keyboard LeftControl)	19 E0
  USAGE_MAXIMUM (Keyboard Right GUI)	29 E7
  LOGICAL_MINIMUM (0)	15 00
  LOGICAL_MAXIMUM (1)	25 01
  REPORT_SIZE (1)	75 01
  REPORT_COUNT (8)	95 08
  INPUT (Data,Var,Abs)	81 02
  REPORT_COUNT (1)	95 01
  REPORT_SIZE (8)	75 08
  INPUT (Cnst,Var,Abs)	81 03
  REPORT_COUNT (5)	95 05
  REPORT_SIZE (1)	75 01
  USAGE_PAGE (LEDs)	05 08
  USAGE_MINIMUM (Num Lock)	19 01
  USAGE_MAXIMUM (Kana)	29 05
  OUTPUT (Data,Var,Abs)	91 02
  REPORT_COUNT (1)	95 01
  REPORT_SIZE (3)	75 03
  OUTPUT (Cnst,Var,Abs)	91 03
  REPORT_COUNT (6)	95 06
  REPORT_SIZE (8)	75 08
  LOGICAL_MINIMUM (0)	15 00
  LOGICAL_MAXIMUM (101)	25 65
  USAGE_PAGE (Keyboard)	05 07
  USAGE_MINIMUM (Reserved (no event indicated))	19 00
  USAGE_MAXIMUM (Keyboard Application)	29 65
  INPUT (Data,Ary,Abs)	81 00
END_COLLECTION	C0

--> Report format INPUT
Byte 0: INPUT - Modifier BitMask
	Bit 0:	Keyboard LeftControl
	Bit 7:	Keyboard Right GUI
Byte 1: INPUT - Constant field
Byte 2: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))
Byte 3: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))
Byte 4: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))
Byte 5: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))
Byte 6: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))
Byte 7: INPUT Keyboard (Values between "reserved (no event)" (0x00) and "Keyboard Application" (0x65))

see: http://www.usb.org/developers/hidpage/Hut1_12v2.pdf

--> Report format OUTPUT
Byte 0: OUTPUT - LED BitMask //Only present in output reports
	Bit 0:	Num Lock
	Bit 4:	Kana
	Bit 5 .. 7:	Constant
 */

const (
	HID_MOD_KEY_LEFT_CONTROL = 0x01
	HID_MOD_KEY_LEFT_SHIFT = 0x02
	HID_MOD_KEY_LEFT_ALT = 0x04
	HID_MOD_KEY_LEFT_GUI = 0x08
	HID_MOD_KEY_RIGHT_CONTROL = 0x10
	HID_MOD_KEY_RIGHT_SHIFT = 0x20
	HID_MOD_KEY_RIGHT_ALT = 0x40
	HID_MOD_KEY_RIGHT_GUI = 0x80

	HID_KEY_RESERVED = 0x00
	HID_KEY_ERROR_ROLLOVER = 0x01
	HID_KEY_POST_FAIL = 0x02
	HID_KEY_ERROR_UNDEFINED = 0x03
	HID_KEY_A= 0x04
	HID_KEY_B= 0x05
	HID_KEY_C= 0x06
	HID_KEY_D = 0x07 // Keyboard d and D
	HID_KEY_E = 0x08 // Keyboard e and E
	HID_KEY_F = 0x09 // Keyboard f and F
	HID_KEY_G = 0x0a // Keyboard g and G
	HID_KEY_H = 0x0b // Keyboard h and H
	HID_KEY_I = 0x0c // Keyboard i and I
	HID_KEY_J = 0x0d // Keyboard j and J
	HID_KEY_K = 0x0e // Keyboard k and K
	HID_KEY_L = 0x0f // Keyboard l and L
	HID_KEY_M = 0x10 // Keyboard m and M
	HID_KEY_N = 0x11 // Keyboard n and N
	HID_KEY_O = 0x12 // Keyboard o and O
	HID_KEY_P = 0x13 // Keyboard p and P
	HID_KEY_Q = 0x14 // Keyboard q and Q
	HID_KEY_R = 0x15 // Keyboard r and R
	HID_KEY_S = 0x16 // Keyboard s and S
	HID_KEY_T = 0x17 // Keyboard t and T
	HID_KEY_U = 0x18 // Keyboard u and U
	HID_KEY_V = 0x19 // Keyboard v and V
	HID_KEY_W = 0x1a // Keyboard w and W
	HID_KEY_X = 0x1b // Keyboard x and X
	HID_KEY_Y = 0x1c // Keyboard y and Y
	HID_KEY_Z = 0x1d // Keyboard z and Z

	HID_KEY_1 = 0x1e // Keyboard 1 and !
	HID_KEY_2 = 0x1f // Keyboard 2 and @
	HID_KEY_3 = 0x20 // Keyboard 3 and #
	HID_KEY_4 = 0x21 // Keyboard 4 and $
	HID_KEY_5 = 0x22 // Keyboard 5 and %
	HID_KEY_6 = 0x23 // Keyboard 6 and ^
	HID_KEY_7 = 0x24 // Keyboard 7 and &
	HID_KEY_8 = 0x25 // Keyboard 8 and *
	HID_KEY_9 = 0x26 // Keyboard 9 and (
	HID_KEY_0 = 0x27 // Keyboard 0 and )

	HID_KEY_ENTER = 0x28 // Keyboard Return (ENTER)
	HID_KEY_ESC = 0x29 // Keyboard ESCAPE
	HID_KEY_BACKSPACE = 0x2a // Keyboard DELETE (Backspace)
	HID_KEY_TAB = 0x2b // Keyboard Tab
	HID_KEY_SPACE = 0x2c // Keyboard Spacebar
	HID_KEY_MINUS = 0x2d // Keyboard - and _
	HID_KEY_EQUAL = 0x2e // Keyboard = and +
	HID_KEY_LEFTBRACE = 0x2f // Keyboard [ and {
	HID_KEY_RIGHTBRACE = 0x30 // Keyboard ] and }
	HID_KEY_BACKSLASH = 0x31 // Keyboard \ and |
	HID_KEY_HASHTILDE = 0x32 // Keyboard Non-US # and ~
	HID_KEY_SEMICOLON = 0x33 // Keyboard ; and :
	HID_KEY_APOSTROPHE = 0x34 // Keyboard ' and "
	HID_KEY_GRAVE = 0x35 // Keyboard ` and ~
	HID_KEY_COMMA = 0x36 // Keyboard , and <
	HID_KEY_DOT = 0x37 // Keyboard . and >
	HID_KEY_SLASH = 0x38 // Keyboard / and ?
	HID_KEY_CAPSLOCK = 0x39 // Keyboard Caps Lock

	HID_KEY_F1 = 0x3a // Keyboard F1
	HID_KEY_F2 = 0x3b // Keyboard F2
	HID_KEY_F3 = 0x3c // Keyboard F3
	HID_KEY_F4 = 0x3d // Keyboard F4
	HID_KEY_F5 = 0x3e // Keyboard F5
	HID_KEY_F6 = 0x3f // Keyboard F6
	HID_KEY_F7 = 0x40 // Keyboard F7
	HID_KEY_F8 = 0x41 // Keyboard F8
	HID_KEY_F9 = 0x42 // Keyboard F9
	HID_KEY_F10 = 0x43 // Keyboard F10
	HID_KEY_F11 = 0x44 // Keyboard F11
	HID_KEY_F12 = 0x45 // Keyboard F12

	HID_KEY_SYSRQ = 0x46 // Keyboard Print Screen
	HID_KEY_SCROLLLOCK = 0x47 // Keyboard Scroll Lock
	HID_KEY_PAUSE = 0x48 // Keyboard Pause
	HID_KEY_INSERT = 0x49 // Keyboard Insert
	HID_KEY_HOME = 0x4a // Keyboard Home
	HID_KEY_PAGEUP = 0x4b // Keyboard Page Up
	HID_KEY_DELETE = 0x4c // Keyboard Delete Forward
	HID_KEY_END = 0x4d // Keyboard End
	HID_KEY_PAGEDOWN = 0x4e // Keyboard Page Down
	HID_KEY_RIGHT = 0x4f // Keyboard Right Arrow
	HID_KEY_LEFT = 0x50 // Keyboard Left Arrow
	HID_KEY_DOWN = 0x51 // Keyboard Down Arrow
	HID_KEY_UP = 0x52 // Keyboard Up Arrow

	HID_KEY_NUMLOCK = 0x53 // Keyboard Num Lock and Clear
	HID_KEY_KPSLASH = 0x54 // Keypad /
	HID_KEY_KPASTERISK = 0x55 // Keypad *
	HID_KEY_KPMINUS = 0x56 // Keypad -
	HID_KEY_KPPLUS = 0x57 // Keypad +
	HID_KEY_KPENTER = 0x58 // Keypad ENTER
	HID_KEY_KP1 = 0x59 // Keypad 1 and End
	HID_KEY_KP2 = 0x5a // Keypad 2 and Down Arrow
	HID_KEY_KP3 = 0x5b // Keypad 3 and PageDn
	HID_KEY_KP4 = 0x5c // Keypad 4 and Left Arrow
	HID_KEY_KP5 = 0x5d // Keypad 5
	HID_KEY_KP6 = 0x5e // Keypad 6 and Right Arrow
	HID_KEY_KP7 = 0x5f // Keypad 7 and Home
	HID_KEY_KP8 = 0x60 // Keypad 8 and Up Arrow
	HID_KEY_KP9 = 0x61 // Keypad 9 and Page Up
	HID_KEY_KP0 = 0x62 // Keypad 0 and Insert
	HID_KEY_KPDOT = 0x63 // Keypad . and Delete

	HID_KEY_102ND = 0x64 // Keyboard Non-US \ and |
	HID_KEY_COMPOSE = 0x65 // Keyboard Application
	HID_KEY_POWER = 0x66 // Keyboard Power
	HID_KEY_KPEQUAL = 0x67 // Keypad =

	HID_KEY_F13 = 0x68 // Keyboard F13
	HID_KEY_F14 = 0x69 // Keyboard F14
	HID_KEY_F15 = 0x6a // Keyboard F15
	HID_KEY_F16 = 0x6b // Keyboard F16
	HID_KEY_F17 = 0x6c // Keyboard F17
	HID_KEY_F18 = 0x6d // Keyboard F18
	HID_KEY_F19 = 0x6e // Keyboard F19
	HID_KEY_F20 = 0x6f // Keyboard F20
	HID_KEY_F21 = 0x70 // Keyboard F21
	HID_KEY_F22 = 0x71 // Keyboard F22
	HID_KEY_F23 = 0x72 // Keyboard F23
	HID_KEY_F24 = 0x73 // Keyboard F24

	HID_KEY_OPEN = 0x74 // Keyboard Execute
	HID_KEY_HELP = 0x75 // Keyboard Help
	HID_KEY_PROPS = 0x76 // Keyboard Menu
	HID_KEY_FRONT = 0x77 // Keyboard Select
	HID_KEY_STOP = 0x78 // Keyboard Stop
	HID_KEY_AGAIN = 0x79 // Keyboard Again
	HID_KEY_UNDO = 0x7a // Keyboard Undo
	HID_KEY_CUT = 0x7b // Keyboard Cut
	HID_KEY_COPY = 0x7c // Keyboard Copy
	HID_KEY_PASTE = 0x7d // Keyboard Paste
	HID_KEY_FIND = 0x7e // Keyboard Find
	HID_KEY_MUTE = 0x7f // Keyboard Mute
	HID_KEY_VOLUMEUP = 0x80 // Keyboard Volume Up
	HID_KEY_VOLUMEDOWN = 0x81 // Keyboard Volume Down
	// = 0x82  Keyboard Locking Caps Lock
	// = 0x83  Keyboard Locking Num Lock
	// = 0x84  Keyboard Locking Scroll Lock
	HID_KEY_KPCOMMA = 0x85 // Keypad Comma
	// = 0x86  Keypad Equal Sign
	HID_KEY_RO = 0x87 // Keyboard International1
	HID_KEY_KATAKANAHIRAGANA = 0x88 // Keyboard International2
	HID_KEY_YEN = 0x89 // Keyboard International3
	HID_KEY_HENKAN = 0x8a // Keyboard International4
	HID_KEY_MUHENKAN = 0x8b // Keyboard International5
	HID_KEY_KPJPCOMMA = 0x8c // Keyboard International6
	// = 0x8d  Keyboard International7
	// = 0x8e  Keyboard International8
	// = 0x8f  Keyboard International9
	HID_KEY_HANGEUL = 0x90 // Keyboard LANG1
	HID_KEY_HANJA = 0x91 // Keyboard LANG2
	HID_KEY_KATAKANA = 0x92 // Keyboard LANG3
	HID_KEY_HIRAGANA = 0x93 // Keyboard LANG4
	HID_KEY_ZENKAKUHANKAKU = 0x94 // Keyboard LANG5
	// = 0x95  Keyboard LANG6
	// = 0x96  Keyboard LANG7
	// = 0x97  Keyboard LANG8
	// = 0x98  Keyboard LANG9
	// = 0x99  Keyboard Alternate Erase
	// = 0x9a  Keyboard SysReq/Attention
	// = 0x9b  Keyboard Cancel
	// = 0x9c  Keyboard Clear
	// = 0x9d  Keyboard Prior
	// = 0x9e  Keyboard Return
	// = 0x9f  Keyboard Separator
	// = 0xa0  Keyboard Out
	// = 0xa1  Keyboard Oper
	// = 0xa2  Keyboard Clear/Again
	// = 0xa3  Keyboard CrSel/Props
	// = 0xa4  Keyboard ExSel

	// = 0xb0  Keypad 00
	// = 0xb1  Keypad 000
	// = 0xb2  Thousands Separator
	// = 0xb3  Decimal Separator
	// = 0xb4  Currency Unit
	// = 0xb5  Currency Sub-unit
	HID_KEY_KPLEFTPAREN = 0xb6 // Keypad (
	HID_KEY_KPRIGHTPAREN = 0xb7 // Keypad )
	// = 0xb8  Keypad {
	// = 0xb9  Keypad }
	// = 0xba  Keypad Tab
	// = 0xbb  Keypad Backspace
	// = 0xbc  Keypad A
	// = 0xbd  Keypad B
	// = 0xbe  Keypad C
	// = 0xbf  Keypad D
	// = 0xc0  Keypad E
	// = 0xc1  Keypad F
	// = 0xc2  Keypad XOR
	// = 0xc3  Keypad ^
	// = 0xc4  Keypad %
	// = 0xc5  Keypad <
	// = 0xc6  Keypad >
	// = 0xc7  Keypad &
	// = 0xc8  Keypad &&
	// = 0xc9  Keypad |
	// = 0xca  Keypad ||
	// = 0xcb  Keypad :
	// = 0xcc  Keypad #
	// = 0xcd  Keypad Space
	// = 0xce  Keypad @
	// = 0xcf  Keypad !
	// = 0xd0  Keypad Memory Store
	// = 0xd1  Keypad Memory Recall
	// = 0xd2  Keypad Memory Clear
	// = 0xd3  Keypad Memory Add
	// = 0xd4  Keypad Memory Subtract
	// = 0xd5  Keypad Memory Multiply
	// = 0xd6  Keypad Memory Divide
	// = 0xd7  Keypad +/-
	// = 0xd8  Keypad Clear
	// = 0xd9  Keypad Clear Entry
	// = 0xda  Keypad Binary
	// = 0xdb  Keypad Octal
	// = 0xdc  Keypad Decimal
	// = 0xdd  Keypad Hexadecimal

	HID_KEY_LEFTCTRL = 0xe0 // Keyboard Left Control
	HID_KEY_LEFTSHIFT = 0xe1 // Keyboard Left Shift
	HID_KEY_LEFTALT = 0xe2 // Keyboard Left Alt
	HID_KEY_LEFTMETA = 0xe3 // Keyboard Left GUI
	HID_KEY_RIGHTCTRL = 0xe4 // Keyboard Right Control
	HID_KEY_RIGHTSHIFT = 0xe5 // Keyboard Right Shift
	HID_KEY_RIGHTALT = 0xe6 // Keyboard Right Alt
	HID_KEY_RIGHTMETA = 0xe7 // Keyboard Right GUI

	HID_KEY_MEDIA_PLAYPAUSE = 0xe8
	HID_KEY_MEDIA_STOPCD = 0xe9
	HID_KEY_MEDIA_PREVIOUSSONG = 0xea
	HID_KEY_MEDIA_NEXTSONG = 0xeb
	HID_KEY_MEDIA_EJECTCD = 0xec
	HID_KEY_MEDIA_VOLUMEUP = 0xed
	HID_KEY_MEDIA_VOLUMEDOWN = 0xee
	HID_KEY_MEDIA_MUTE = 0xef
	HID_KEY_MEDIA_WWW = 0xf0
	HID_KEY_MEDIA_BACK = 0xf1
	HID_KEY_MEDIA_FORWARD = 0xf2
	HID_KEY_MEDIA_STOP = 0xf3
	HID_KEY_MEDIA_FIND = 0xf4
	HID_KEY_MEDIA_SCROLLUP = 0xf5
	HID_KEY_MEDIA_SCROLLDOWN = 0xf6
	HID_KEY_MEDIA_EDIT = 0xf7
	HID_KEY_MEDIA_SLEEP = 0xf8
	HID_KEY_MEDIA_COFFEE = 0xf9
	HID_KEY_MEDIA_REFRESH = 0xfa
	HID_KEY_MEDIA_CALC = 0xfb
)

var (
	UsbKeyToString    = generateKey2Str()
	StringToUsbKey    = generateStr2Key()
	UsbModKeyToString = generateMod2Str()
	StringToUsbModKey = generateStr2Mod()
)

func generateMod2Str() (m2s map[uint8]string) {
	m2s = make(map[uint8]string)
	m2s[HID_MOD_KEY_LEFT_CONTROL] = "MOD_LEFT_CONTROL"
	m2s[HID_MOD_KEY_LEFT_SHIFT] = "MOD_LEFT_SHIFT"
	m2s[HID_MOD_KEY_LEFT_ALT] = "MOD_LEFT_ALT"
	m2s[HID_MOD_KEY_LEFT_GUI] = "MOD_LEFT_GUI"
	m2s[HID_MOD_KEY_RIGHT_CONTROL] = "MOD_RIGHT_CONTROL"
	m2s[HID_MOD_KEY_RIGHT_SHIFT] = "MOD_RIGHT_SHIFT"
	m2s[HID_MOD_KEY_RIGHT_ALT] = "MOD_RIGHT_ALT"
	m2s[HID_MOD_KEY_RIGHT_GUI] = "MOD_RIGHT_GUI"
	return
}

func generateStr2Mod() (s2m map[string]uint8) {
	s2m = make(map[string]uint8)
	s2m["MOD_LEFT_CONTROL"] = HID_MOD_KEY_LEFT_CONTROL
	s2m["MOD_LEFT_SHIFT"] = HID_MOD_KEY_LEFT_SHIFT
	s2m["MOD_LEFT_ALT"] = HID_MOD_KEY_LEFT_ALT
	s2m["MOD_LEFT_GUI"] = HID_MOD_KEY_LEFT_GUI
	s2m["MOD_RIGHT_CONTROL"] =HID_MOD_KEY_RIGHT_CONTROL
	s2m["MOD_RIGHT_SHIFT"] = HID_MOD_KEY_RIGHT_SHIFT
	s2m["MOD_RIGHT_ALT"] = HID_MOD_KEY_RIGHT_ALT
	s2m["MOD_RIGHT_GUI"] = HID_MOD_KEY_RIGHT_GUI
	return
}

func generateKey2Str() (k2s map[uint8]string) {
	k2s = make(map[uint8]string)
	k2s[HID_KEY_RESERVED] = "KEY_RESERVED"
	k2s[HID_KEY_ERROR_ROLLOVER] = "KEY_ERROR_ROLLOVER"
	k2s[HID_KEY_POST_FAIL] = "KEY_POST_FAIL"
	k2s[HID_KEY_ERROR_UNDEFINED] = "KEY_ERROR_UNDEFINED"
	k2s[HID_KEY_A] = "KEY_A" // Keyboard a and A
	k2s[HID_KEY_B] = "KEY_B" // Keyboard b and B
	k2s[HID_KEY_C] = "KEY_C" // Keyboard c and C
	k2s[HID_KEY_D] = "KEY_D" // Keyboard d and D
	k2s[HID_KEY_E] = "KEY_E" // Keyboard e and E
	k2s[HID_KEY_F] = "KEY_F" // Keyboard f and F
	k2s[HID_KEY_G] = "KEY_G" // Keyboard g and G
	k2s[HID_KEY_H] = "KEY_H" // Keyboard h and H
	k2s[HID_KEY_I] = "KEY_I" // Keyboard i and I
	k2s[HID_KEY_J] = "KEY_J" //0x0d // Keyboard j and J
	k2s[HID_KEY_K] = "KEY_K" //0x0e // Keyboard k and K
	k2s[HID_KEY_L] = "KEY_L" //0x0f // Keyboard l and L
	k2s[HID_KEY_M] = "KEY_M" //0x10 // Keyboard m and M
	k2s[HID_KEY_N] = "KEY_N" //0x11 // Keyboard n and N
	k2s[HID_KEY_O] = "KEY_O" //0x12 // Keyboard o and O
	k2s[HID_KEY_P] = "KEY_P" //0x13 // Keyboard p and P
	k2s[HID_KEY_Q] = "KEY_Q" //0x14 // Keyboard q and Q
	k2s[HID_KEY_R] = "KEY_R" //0x15 // Keyboard r and R
	k2s[HID_KEY_S] = "KEY_S" //0x16 // Keyboard s and S
	k2s[HID_KEY_T] = "KEY_T" //0x17 // Keyboard t and T
	k2s[HID_KEY_U] = "KEY_U" //0x18 // Keyboard u and U
	k2s[HID_KEY_V] = "KEY_V" //0x19 // Keyboard v and V
	k2s[HID_KEY_W] = "KEY_W" //0x1a // Keyboard w and W
	k2s[HID_KEY_X] = "KEY_X" //0x1b // Keyboard x and X
	k2s[HID_KEY_Y] = "KEY_Y" //0x1c // Keyboard y and Y
	k2s[HID_KEY_Z] = "KEY_Z" //0x1d // Keyboard z and Z

	k2s[HID_KEY_1] = "KEY_1" //0x1e // Keyboard 1 and !
	k2s[HID_KEY_2] = "KEY_2" //0x1f // Keyboard 2 and @
	k2s[HID_KEY_3] = "KEY_3" //0x20 // Keyboard 3 and #
	k2s[HID_KEY_4] = "KEY_4" //0x21 // Keyboard 4 and $
	k2s[HID_KEY_5] = "KEY_5" //0x22 // Keyboard 5 and %
	k2s[HID_KEY_6] = "KEY_6" //0x23 // Keyboard 6 and ^
	k2s[HID_KEY_7] = "KEY_7" //0x24 // Keyboard 7 and &
	k2s[HID_KEY_8] = "KEY_8" //0x25 // Keyboard 8 and *
	k2s[HID_KEY_9] = "KEY_9" //0x26 // Keyboard 9 and (
	k2s[HID_KEY_0] = "KEY_0" //0x27 // Keyboard 0 and )

	k2s[HID_KEY_ENTER] = "KEY_ENTER" //0x28 // Keyboard Return (ENTER)
	k2s[HID_KEY_ESC] = "KEY_ESC" //0x29 // Keyboard ESCAPE
	k2s[HID_KEY_BACKSPACE] = "KEY_BACKSPACE" //0x2a // Keyboard DELETE (Backspace)
	k2s[HID_KEY_TAB] = "KEY_TAB" //0x2b // Keyboard Tab
	k2s[HID_KEY_SPACE] = "KEY_SPACE" //0x2c // Keyboard Spacebar
	k2s[HID_KEY_MINUS] = "KEY_MINUS" //0x2d // Keyboard - and _
	k2s[HID_KEY_EQUAL] = "KEY_EQUAL" //0x2e // Keyboard = and +
	k2s[HID_KEY_LEFTBRACE] = "KEY_LEFTBRACE" //0x2f // Keyboard [ and {
	k2s[HID_KEY_RIGHTBRACE] = "KEY_RIGHTBRACE" //0x30 // Keyboard ] and }
	k2s[HID_KEY_BACKSLASH] = "KEY_BACKSLASH" //0x31 // Keyboard \ and |
	k2s[HID_KEY_HASHTILDE] = "KEY_HASHTILDE" //0x32 // Keyboard Non-US # and ~
	k2s[HID_KEY_SEMICOLON] = "KEY_SEMICOLON" //0x33 // Keyboard ; and :
	k2s[HID_KEY_APOSTROPHE] = "KEY_APOSTROPHE" //0x34 // Keyboard ' and "
	k2s[HID_KEY_GRAVE] = "KEY_GRAVE" //0x35 // Keyboard ` and ~
	k2s[HID_KEY_COMMA] = "KEY_COMMA" //0x36 // Keyboard , and <
	k2s[HID_KEY_DOT] = "KEY_DOT" //0x37 // Keyboard . and >
	k2s[HID_KEY_SLASH] = "KEY_SLASH" //0x38 // Keyboard / and ?
	k2s[HID_KEY_CAPSLOCK] = "KEY_CAPSLOCK" //0x39 // Keyboard Caps Lock

	k2s[HID_KEY_F1] = "KEY_F1" //0x3a // Keyboard F1
	k2s[HID_KEY_F2] = "KEY_F2" //0x3b // Keyboard F2
	k2s[HID_KEY_F3] = "KEY_F3" //0x3c // Keyboard F3
	k2s[HID_KEY_F4] = "KEY_F4" //0x3d // Keyboard F4
	k2s[HID_KEY_F5] = "KEY_F5" //0x3e // Keyboard F5
	k2s[HID_KEY_F6] = "KEY_F6" //0x3f // Keyboard F6
	k2s[HID_KEY_F7] = "KEY_F7" //0x40 // Keyboard F7
	k2s[HID_KEY_F8] = "KEY_F8" //0x41 // Keyboard F8
	k2s[HID_KEY_F9] = "KEY_F9" //0x42 // Keyboard F9
	k2s[HID_KEY_F10] = "KEY_F10" //0x43 // Keyboard F10
	k2s[HID_KEY_F11] = "KEY_F11" //0x44 // Keyboard F11
	k2s[HID_KEY_F12] = "KEY_F12" //0x45 // Keyboard F12

	k2s[HID_KEY_SYSRQ] = "KEY_SYSRQ" //0x46 // Keyboard Print Screen
	k2s[HID_KEY_SCROLLLOCK] = "KEY_SCROLLLOCK" //0x47 // Keyboard Scroll Lock
	k2s[HID_KEY_PAUSE] = "KEY_PAUSE" //0x48 // Keyboard Pause
	k2s[HID_KEY_INSERT] = "KEY_INSERT" //0x49 // Keyboard Insert
	k2s[HID_KEY_HOME] = "KEY_HOME" //0x4a // Keyboard Home
	k2s[HID_KEY_PAGEUP] = "KEY_PAGEUP" //0x4b // Keyboard Page Up
	k2s[HID_KEY_DELETE] = "KEY_DELETE" //0x4c // Keyboard Delete Forward
	k2s[HID_KEY_END] = "KEY_END" //0x4d // Keyboard End
	k2s[HID_KEY_PAGEDOWN] = "KEY_PAGEDOWN" //0x4e // Keyboard Page Down
	k2s[HID_KEY_RIGHT] = "KEY_RIGHT" //0x4f // Keyboard Right Arrow
	k2s[HID_KEY_LEFT] = "KEY_LEFT" //0x50 // Keyboard Left Arrow
	k2s[HID_KEY_DOWN] = "KEY_DOWN" //0x51 // Keyboard Down Arrow
	k2s[HID_KEY_UP] = "KEY_UP" //0x52 // Keyboard Up Arrow

	k2s[HID_KEY_NUMLOCK] = "KEY_NUMLOCK" //0x53 // Keyboard Num Lock and Clear
	k2s[HID_KEY_KPSLASH] = "KEY_KPSLASH" //0x54 // Keypad /
	k2s[HID_KEY_KPASTERISK] = "KEY_KPASTERISK" //0x55 // Keypad *
	k2s[HID_KEY_KPMINUS] = "KEY_KPMINUS" //0x56 // Keypad -
	k2s[HID_KEY_KPPLUS] = "KEY_KPPLUS" //0x57 // Keypad +
	k2s[HID_KEY_KPENTER] = "KEY_KPENTER" //0x58 // Keypad ENTER
	k2s[HID_KEY_KP1] = "KEY_KP1" //0x59 // Keypad 1 and End
	k2s[HID_KEY_KP2] = "KEY_KP2" //0x5a // Keypad 2 and Down Arrow
	k2s[HID_KEY_KP3] = "KEY_KP3" //0x5b // Keypad 3 and PageDn
	k2s[HID_KEY_KP4] = "KEY_KP4" //0x5c // Keypad 4 and Left Arrow
	k2s[HID_KEY_KP5] = "KEY_KP5" //0x5d // Keypad 5
	k2s[HID_KEY_KP6] = "KEY_KP6" //0x5e // Keypad 6 and Right Arrow
	k2s[HID_KEY_KP7] = "KEY_KP7" //0x5f // Keypad 7 and Home
	k2s[HID_KEY_KP8] = "KEY_KP8" //0x60 // Keypad 8 and Up Arrow
	k2s[HID_KEY_KP9] = "KEY_KP9" //0x61 // Keypad 9 and Page Up
	k2s[HID_KEY_KP0] = "KEY_KP0" //0x62 // Keypad 0 and Insert
	k2s[HID_KEY_KPDOT] = "KEY_KPDOT" //0x63 // Keypad . and Delete

	k2s[HID_KEY_102ND] = "KEY_102ND" //0x64 // Keyboard Non-US \ and |
	k2s[HID_KEY_COMPOSE] = "KEY_COMPOSE" //0x65 // Keyboard Application
	k2s[HID_KEY_POWER] = "KEY_POWER" //0x66 // Keyboard Power
	k2s[HID_KEY_KPEQUAL] = "KEY_KPEQUAL" //0x67 // Keypad =

	k2s[HID_KEY_F13] = "KEY_F13" //0x68 // Keyboard F13
	k2s[HID_KEY_F14] = "KEY_F14" //0x69 // Keyboard F14
	k2s[HID_KEY_F15] = "KEY_F15" //0x6a // Keyboard F15
	k2s[HID_KEY_F16] = "KEY_F16" //0x6b // Keyboard F16
	k2s[HID_KEY_F17] = "KEY_F17" //0x6c // Keyboard F17
	k2s[HID_KEY_F18] = "KEY_F18" //0x6d // Keyboard F18
	k2s[HID_KEY_F19] = "KEY_F19" //0x6e // Keyboard F19
	k2s[HID_KEY_F20] = "KEY_F20" //0x6f // Keyboard F20
	k2s[HID_KEY_F21] = "KEY_F21" //0x70 // Keyboard F21
	k2s[HID_KEY_F22] = "KEY_F22" //0x71 // Keyboard F22
	k2s[HID_KEY_F23] = "KEY_F23" //0x72 // Keyboard F23
	k2s[HID_KEY_F24] = "KEY_F24" //0x73 // Keyboard F24

	k2s[HID_KEY_OPEN] = "KEY_OPEN" //0x74 // Keyboard Execute
	k2s[HID_KEY_HELP] = "KEY_HELP" //0x75 // Keyboard Help
	k2s[HID_KEY_PROPS] = "KEY_PROPS" //0x76 // Keyboard Menu
	k2s[HID_KEY_FRONT] = "KEY_FRONT" //0x77 // Keyboard Select
	k2s[HID_KEY_STOP] = "KEY_STOP" //0x78 // Keyboard Stop
	k2s[HID_KEY_AGAIN] = "KEY_AGAIN" //0x79 // Keyboard Again
	k2s[HID_KEY_UNDO] = "KEY_UNDO" //0x7a // Keyboard Undo
	k2s[HID_KEY_CUT] = "KEY_CUT" //0x7b // Keyboard Cut
	k2s[HID_KEY_COPY] = "KEY_COPY" //0x7c // Keyboard Copy
	k2s[HID_KEY_PASTE] = "KEY_PASTE" //0x7d // Keyboard Paste
	k2s[HID_KEY_FIND] = "KEY_FIND" //0x7e // Keyboard Find
	k2s[HID_KEY_MUTE] = "KEY_MUTE" //0x7f // Keyboard Mute
	k2s[HID_KEY_VOLUMEUP] = "KEY_VOLUMEUP" //0x80 // Keyboard Volume Up
	k2s[HID_KEY_VOLUMEDOWN] = "KEY_VOLUMEDOWN" //0x81 // Keyboard Volume Down
	// = 0x82  Keyboard Locking Caps Lock
	// = 0x83  Keyboard Locking Num Lock
	// = 0x84  Keyboard Locking Scroll Lock
	k2s[HID_KEY_KPCOMMA] = "KEY_KPCOMMA" //0x85 // Keypad Comma
	// = 0x86  Keypad Equal Sign
	k2s[HID_KEY_RO] = "KEY_RO" //0x87 // Keyboard International1
	k2s[HID_KEY_KATAKANAHIRAGANA] = "KEY_KATAKANAHIRAGANA" //0x88 // Keyboard International2
	k2s[HID_KEY_YEN] = "KEY_YEN" //0x89 // Keyboard International3
	k2s[HID_KEY_HENKAN] = "KEY_HENKAN" //0x8a // Keyboard International4
	k2s[HID_KEY_MUHENKAN] = "KEY_MUHENKAN" //0x8b // Keyboard International5
	k2s[HID_KEY_KPJPCOMMA] = "KEY_KPJPCOMMA" //0x8c // Keyboard International6
	// = 0x8d  Keyboard International7
	// = 0x8e  Keyboard International8
	// = 0x8f  Keyboard International9
	k2s[HID_KEY_HANGEUL] = "KEY_HANGEUL" //0x90 // Keyboard LANG1
	k2s[HID_KEY_HANJA] = "KEY_HANJA" //0x91 // Keyboard LANG2
	k2s[HID_KEY_KATAKANA] = "KEY_KATAKANA" //0x92 // Keyboard LANG3
	k2s[HID_KEY_HIRAGANA] = "KEY_HIRAGANA" //0x93 // Keyboard LANG4
	k2s[HID_KEY_ZENKAKUHANKAKU] = "KEY_ZENKAKUHANKAKU" //0x94 // Keyboard LANG5
	// = 0x95  Keyboard LANG6
	// = 0x96  Keyboard LANG7
	// = 0x97  Keyboard LANG8
	// = 0x98  Keyboard LANG9
	// = 0x99  Keyboard Alternate Erase
	// = 0x9a  Keyboard SysReq/Attention
	// = 0x9b  Keyboard Cancel
	// = 0x9c  Keyboard Clear
	// = 0x9d  Keyboard Prior
	// = 0x9e  Keyboard Return
	// = 0x9f  Keyboard Separator
	// = 0xa0  Keyboard Out
	// = 0xa1  Keyboard Oper
	// = 0xa2  Keyboard Clear/Again
	// = 0xa3  Keyboard CrSel/Props
	// = 0xa4  Keyboard ExSel

	// = 0xb0  Keypad 00
	// = 0xb1  Keypad 000
	// = 0xb2  Thousands Separator
	// = 0xb3  Decimal Separator
	// = 0xb4  Currency Unit
	// = 0xb5  Currency Sub-unit
	k2s[HID_KEY_KPLEFTPAREN] = "KEY_KPLEFTPAREN" //0xb6 // Keypad (
	k2s[HID_KEY_KPRIGHTPAREN] = "KEY_KPRIGHTPAREN" //0xb7 // Keypad )
	// = 0xb8  Keypad {
	// = 0xb9  Keypad }
	// = 0xba  Keypad Tab
	// = 0xbb  Keypad Backspace
	// = 0xbc  Keypad A
	// = 0xbd  Keypad B
	// = 0xbe  Keypad C
	// = 0xbf  Keypad D
	// = 0xc0  Keypad E
	// = 0xc1  Keypad F
	// = 0xc2  Keypad XOR
	// = 0xc3  Keypad ^
	// = 0xc4  Keypad %
	// = 0xc5  Keypad <
	// = 0xc6  Keypad >
	// = 0xc7  Keypad &
	// = 0xc8  Keypad &&
	// = 0xc9  Keypad |
	// = 0xca  Keypad ||
	// = 0xcb  Keypad :
	// = 0xcc  Keypad #
	// = 0xcd  Keypad Space
	// = 0xce  Keypad @
	// = 0xcf  Keypad !
	// = 0xd0  Keypad Memory Store
	// = 0xd1  Keypad Memory Recall
	// = 0xd2  Keypad Memory Clear
	// = 0xd3  Keypad Memory Add
	// = 0xd4  Keypad Memory Subtract
	// = 0xd5  Keypad Memory Multiply
	// = 0xd6  Keypad Memory Divide
	// = 0xd7  Keypad +/-
	// = 0xd8  Keypad Clear
	// = 0xd9  Keypad Clear Entry
	// = 0xda  Keypad Binary
	// = 0xdb  Keypad Octal
	// = 0xdc  Keypad Decimal
	// = 0xdd  Keypad Hexadecimal

	k2s[HID_KEY_LEFTCTRL] = "KEY_LEFTCTRL" //0xe0 // Keyboard Left Control
	k2s[HID_KEY_LEFTSHIFT] = "KEY_LEFTSHIFT" //0xe1 // Keyboard Left Shift
	k2s[HID_KEY_LEFTALT] = "KEY_LEFTALT" //0xe2 // Keyboard Left Alt
	k2s[HID_KEY_LEFTMETA] = "KEY_LEFTMETA" //0xe3 // Keyboard Left GUI
	k2s[HID_KEY_RIGHTCTRL] = "KEY_RIGHTCTRL" //0xe4 // Keyboard Right Control
	k2s[HID_KEY_RIGHTSHIFT] = "KEY_RIGHTSHIFT" //0xe5 // Keyboard Right Shift
	k2s[HID_KEY_RIGHTALT] = "KEY_RIGHTALT" //0xe6 // Keyboard Right Alt
	k2s[HID_KEY_RIGHTMETA] = "KEY_RIGHTMETA" //0xe7 // Keyboard Right GUI

	k2s[HID_KEY_MEDIA_PLAYPAUSE] = "KEY_MEDIA_PLAYPAUSE" //0xe8
	k2s[HID_KEY_MEDIA_STOPCD] = "KEY_MEDIA_STOPCD" //0xe9
	k2s[HID_KEY_MEDIA_PREVIOUSSONG] = "KEY_MEDIA_PREVIOUSSONG" //0xea
	k2s[HID_KEY_MEDIA_NEXTSONG] = "KEY_MEDIA_NEXTSONG" //0xeb
	k2s[HID_KEY_MEDIA_EJECTCD] = "KEY_MEDIA_EJECTCD" //0xec
	k2s[HID_KEY_MEDIA_VOLUMEUP] = "KEY_MEDIA_VOLUMEUP" //0xed
	k2s[HID_KEY_MEDIA_VOLUMEDOWN] = "KEY_MEDIA_VOLUMEDOWN" //0xee
	k2s[HID_KEY_MEDIA_MUTE] = "KEY_MEDIA_MUTE" //0xef
	k2s[HID_KEY_MEDIA_WWW] = "KEY_MEDIA_WWW" //0xf0
	k2s[HID_KEY_MEDIA_BACK] = "KEY_MEDIA_BACK" //0xf1
	k2s[HID_KEY_MEDIA_FORWARD] = "KEY_MEDIA_FORWARD" //0xf2
	k2s[HID_KEY_MEDIA_STOP] = "KEY_MEDIA_STOP" //0xf3
	k2s[HID_KEY_MEDIA_FIND] = "KEY_MEDIA_FIND" //0xf4
	k2s[HID_KEY_MEDIA_SCROLLUP] = "KEY_MEDIA_SCROLLUP" //0xf5
	k2s[HID_KEY_MEDIA_SCROLLDOWN] = "KEY_MEDIA_SCROLLDOWN" //0xf6
	k2s[HID_KEY_MEDIA_EDIT] = "KEY_MEDIA_EDIT" //0xf7
	k2s[HID_KEY_MEDIA_SLEEP] = "KEY_MEDIA_SLEEP" //0xf8
	k2s[HID_KEY_MEDIA_COFFEE] = "KEY_MEDIA_COFFEE" //0xf9
	k2s[HID_KEY_MEDIA_REFRESH] = "KEY_MEDIA_REFRESH" //0xfa
	k2s[HID_KEY_MEDIA_CALC] = "KEY_MEDIA_CALC" //0xfb

	return
}

func generateStr2Key() (s2k map[string]uint8) {
	s2k = make(map[string]uint8)
	s2k["KEY_RESERVED"] = HID_KEY_RESERVED
	s2k["KEY_ERROR_ROLLOVER"] = HID_KEY_ERROR_ROLLOVER
	s2k["KEY_POST_FAIL"] = HID_KEY_POST_FAIL
	s2k["KEY_ERROR_UNDEFINED"] = HID_KEY_ERROR_UNDEFINED
	s2k["KEY_A"] = HID_KEY_A // Keyboard a and A
	s2k["KEY_B"] = HID_KEY_B // Keyboard b and B
	s2k["KEY_C"] = HID_KEY_C // Keyboard c and C
	s2k["KEY_D"] = HID_KEY_D // Keyboard d and D
	s2k["KEY_E"] = HID_KEY_E // Keyboard e and E
	s2k["KEY_F"] = HID_KEY_F // Keyboard f and F
	s2k["KEY_G"] = HID_KEY_G // Keyboard g and G
	s2k["KEY_H"] = HID_KEY_H // Keyboard h and H
	s2k["KEY_I"] = HID_KEY_I // Keyboard i and I
	s2k["KEY_J"] = HID_KEY_J //0x0d // Keyboard j and J
	s2k["KEY_K"] = HID_KEY_K //0x0e // Keyboard k and K
	s2k["KEY_L"] = HID_KEY_L //0x0f // Keyboard l and L
	s2k["KEY_M"] = HID_KEY_M //0x10 // Keyboard m and M
	s2k["KEY_N"] = HID_KEY_N //0x11 // Keyboard n and N
	s2k["KEY_O"] = HID_KEY_O //0x12 // Keyboard o and O
	s2k["KEY_P"] = HID_KEY_P //0x13 // Keyboard p and P
	s2k["KEY_Q"] = HID_KEY_Q //0x14 // Keyboard q and Q
	s2k["KEY_R"] = HID_KEY_R //0x15 // Keyboard r and R
	s2k["KEY_S"] = HID_KEY_S //0x16 // Keyboard s and S
	s2k["KEY_T"] = HID_KEY_T //0x17 // Keyboard t and T
	s2k["KEY_U"] = HID_KEY_U //0x18 // Keyboard u and U
	s2k["KEY_V"] = HID_KEY_V //0x19 // Keyboard v and V
	s2k["KEY_W"] = HID_KEY_W //0x1a // Keyboard w and W
	s2k["KEY_X"] = HID_KEY_X //0x1b // Keyboard x and X
	s2k["KEY_Y"] = HID_KEY_Y //0x1c // Keyboard y and Y
	s2k["KEY_Z"] = HID_KEY_Z //0x1d // Keyboard z and Z

	s2k["KEY_1"] = HID_KEY_1 //0x1e // Keyboard 1 and !
	s2k["KEY_2"] = HID_KEY_2 //0x1f // Keyboard 2 and @
	s2k["KEY_3"] = HID_KEY_3 //0x20 // Keyboard 3 and #
	s2k["KEY_4"] = HID_KEY_4 //0x21 // Keyboard 4 and $
	s2k["KEY_5"] = HID_KEY_5 //0x22 // Keyboard 5 and %
	s2k["KEY_6"] = HID_KEY_6 //0x23 // Keyboard 6 and ^
	s2k["KEY_7"] = HID_KEY_7 //0x24 // Keyboard 7 and &
	s2k["KEY_8"] = HID_KEY_8 //0x25 // Keyboard 8 and *
	s2k["KEY_9"] = HID_KEY_9 //0x26 // Keyboard 9 and (
	s2k["KEY_0"] = HID_KEY_0 //0x27 // Keyboard 0 and )

	s2k["KEY_ENTER"] = HID_KEY_ENTER //0x28 // Keyboard Return (ENTER)
	s2k["KEY_ESC"] = HID_KEY_ESC //0x29 // Keyboard ESCAPE
	s2k["KEY_BACKSPACE"] = HID_KEY_BACKSPACE //0x2a // Keyboard DELETE (Backspace)
	s2k["KEY_TAB"] = HID_KEY_TAB //0x2b // Keyboard Tab
	s2k["KEY_SPACE"] = HID_KEY_SPACE //0x2c // Keyboard Spacebar
	s2k["KEY_MINUS"] = HID_KEY_MINUS //0x2d // Keyboard - and _
	s2k["KEY_EQUAL"] = HID_KEY_EQUAL //0x2e // Keyboard = and +
	s2k["KEY_LEFTBRACE"] = HID_KEY_LEFTBRACE //0x2f // Keyboard [ and {
	s2k["KEY_RIGHTBRACE"] = HID_KEY_RIGHTBRACE //0x30 // Keyboard "] and }
	s2k["KEY_BACKSLASH"] = HID_KEY_BACKSLASH //0x31 // Keyboard \ and |
	s2k["KEY_HASHTILDE"] = HID_KEY_HASHTILDE //0x32 // Keyboard Non-US # and ~
	s2k["KEY_SEMICOLON"] = HID_KEY_SEMICOLON //0x33 // Keyboard ; and :
	s2k["KEY_APOSTROPHE"] = HID_KEY_APOSTROPHE //0x34 // Keyboard ' and "
	s2k["KEY_GRAVE"] = HID_KEY_GRAVE //0x35 // Keyboard ` and ~
	s2k["KEY_COMMA"] = HID_KEY_COMMA //0x36 // Keyboard , and <
	s2k["KEY_DOT"] = HID_KEY_DOT //0x37 // Keyboard . and >
	s2k["KEY_SLASH"] = HID_KEY_SLASH //0x38 // Keyboard / and ?
	s2k["KEY_CAPSLOCK"] = HID_KEY_CAPSLOCK //0x39 // Keyboard Caps Lock

	s2k["KEY_F1"] = HID_KEY_F1 //0x3a // Keyboard F1
	s2k["KEY_F2"] = HID_KEY_F2 //0x3b // Keyboard F2
	s2k["KEY_F3"] = HID_KEY_F3 //0x3c // Keyboard F3
	s2k["KEY_F4"] = HID_KEY_F4 //0x3d // Keyboard F4
	s2k["KEY_F5"] = HID_KEY_F5 //0x3e // Keyboard F5
	s2k["KEY_F6"] = HID_KEY_F6 //0x3f // Keyboard F6
	s2k["KEY_F7"] = HID_KEY_F7 //0x40 // Keyboard F7
	s2k["KEY_F8"] = HID_KEY_F8 //0x41 // Keyboard F8
	s2k["KEY_F9"] = HID_KEY_F9 //0x42 // Keyboard F9
	s2k["KEY_F10"] = HID_KEY_F10 //0x43 // Keyboard F10
	s2k["KEY_F11"] = HID_KEY_F11 //0x44 // Keyboard F11
	s2k["KEY_F12"] = HID_KEY_F12 //0x45 // Keyboard F12

	s2k["KEY_SYSRQ"] = HID_KEY_SYSRQ //0x46 // Keyboard Print Screen
	s2k["KEY_SCROLLLOCK"] = HID_KEY_SCROLLLOCK //0x47 // Keyboard Scroll Lock
	s2k["KEY_PAUSE"] = HID_KEY_PAUSE //0x48 // Keyboard Pause
	s2k["KEY_INSERT"] = HID_KEY_INSERT //0x49 // Keyboard Insert
	s2k["KEY_HOME"] = HID_KEY_HOME //0x4a // Keyboard Home
	s2k["KEY_PAGEUP"] = HID_KEY_PAGEUP //0x4b // Keyboard Page Up
	s2k["KEY_DELETE"] = HID_KEY_DELETE //0x4c // Keyboard Delete Forward
	s2k["KEY_END"] = HID_KEY_END //0x4d // Keyboard End
	s2k["KEY_PAGEDOWN"] = HID_KEY_PAGEDOWN //0x4e // Keyboard Page Down
	s2k["KEY_RIGHT"] = HID_KEY_RIGHT //0x4f // Keyboard Right Arrow
	s2k["KEY_LEFT"] = HID_KEY_LEFT //0x50 // Keyboard Left Arrow
	s2k["KEY_DOWN"] = HID_KEY_DOWN //0x51 // Keyboard Down Arrow
	s2k["KEY_UP"] = HID_KEY_UP //0x52 // Keyboard Up Arrow

	s2k["KEY_NUMLOCK"] = HID_KEY_NUMLOCK //0x53 // Keyboard Num Lock and Clear
	s2k["KEY_KPSLASH"] = HID_KEY_KPSLASH //0x54 // Keypad /
	s2k["KEY_KPASTERISK"] = HID_KEY_KPASTERISK //0x55 // Keypad *
	s2k["KEY_KPMINUS"] = HID_KEY_KPMINUS //0x56 // Keypad -
	s2k["KEY_KPPLUS"] = HID_KEY_KPPLUS //0x57 // Keypad +
	s2k["KEY_KPENTER"] = HID_KEY_KPENTER //0x58 // Keypad ENTER
	s2k["KEY_KP1"] = HID_KEY_KP1 //0x59 // Keypad 1 and End
	s2k["KEY_KP2"] = HID_KEY_KP2 //0x5a // Keypad 2 and Down Arrow
	s2k["KEY_KP3"] = HID_KEY_KP3 //0x5b // Keypad 3 and PageDn
	s2k["KEY_KP4"] = HID_KEY_KP4 //0x5c // Keypad 4 and Left Arrow
	s2k["KEY_KP5"] = HID_KEY_KP5 //0x5d // Keypad 5
	s2k["KEY_KP6"] = HID_KEY_KP6 //0x5e // Keypad 6 and Right Arrow
	s2k["KEY_KP7"] = HID_KEY_KP7 //0x5f // Keypad 7 and Home
	s2k["KEY_KP8"] = HID_KEY_KP8 //0x60 // Keypad 8 and Up Arrow
	s2k["KEY_KP9"] = HID_KEY_KP9 //0x61 // Keypad 9 and Page Up
	s2k["KEY_KP0"] = HID_KEY_KP0 //0x62 // Keypad 0 and Insert
	s2k["KEY_KPDOT"] = HID_KEY_KPDOT //0x63 // Keypad . and Delete

	s2k["KEY_102ND"] = HID_KEY_102ND //0x64 // Keyboard Non-US \ and |
	s2k["KEY_COMPOSE"] = HID_KEY_COMPOSE //0x65 // Keyboard Application
	s2k["KEY_POWER"] = HID_KEY_POWER //0x66 // Keyboard Power
	s2k["KEY_KPEQUAL"] = HID_KEY_KPEQUAL //0x67 // Keypad =

	s2k["KEY_F13"] = HID_KEY_F13 //0x68 // Keyboard F13
	s2k["KEY_F14"] = HID_KEY_F14 //0x69 // Keyboard F14
	s2k["KEY_F15"] = HID_KEY_F15 //0x6a // Keyboard F15
	s2k["KEY_F16"] = HID_KEY_F16 //0x6b // Keyboard F16
	s2k["KEY_F17"] = HID_KEY_F17 //0x6c // Keyboard F17
	s2k["KEY_F18"] = HID_KEY_F18 //0x6d // Keyboard F18
	s2k["KEY_F19"] = HID_KEY_F19 //0x6e // Keyboard F19
	s2k["KEY_F20"] = HID_KEY_F20 //0x6f // Keyboard F20
	s2k["KEY_F21"] = HID_KEY_F21 //0x70 // Keyboard F21
	s2k["KEY_F22"] = HID_KEY_F22 //0x71 // Keyboard F22
	s2k["KEY_F23"] = HID_KEY_F23 //0x72 // Keyboard F23
	s2k["KEY_F24"] = HID_KEY_F24 //0x73 // Keyboard F24

	s2k["KEY_OPEN"] = HID_KEY_OPEN //0x74 // Keyboard Execute
	s2k["KEY_HELP"] = HID_KEY_HELP //0x75 // Keyboard Help
	s2k["KEY_PROPS"] = HID_KEY_PROPS //0x76 // Keyboard Menu
	s2k["KEY_FRONT"] = HID_KEY_FRONT //0x77 // Keyboard Select
	s2k["KEY_STOP"] = HID_KEY_STOP //0x78 // Keyboard Stop
	s2k["KEY_AGAIN"] = HID_KEY_AGAIN //0x79 // Keyboard Again
	s2k["KEY_UNDO"] = HID_KEY_UNDO //0x7a // Keyboard Undo
	s2k["KEY_CUT"] = HID_KEY_CUT //0x7b // Keyboard Cut
	s2k["KEY_COPY"] = HID_KEY_COPY //0x7c // Keyboard Copy
	s2k["KEY_PASTE"] = HID_KEY_PASTE //0x7d // Keyboard Paste
	s2k["KEY_FIND"] = HID_KEY_FIND //0x7e // Keyboard Find
	s2k["KEY_MUTE"] = HID_KEY_MUTE //0x7f // Keyboard Mute
	s2k["KEY_VOLUMEUP"] = HID_KEY_VOLUMEUP //0x80 // Keyboard Volume Up
	s2k["KEY_VOLUMEDOWN"] = HID_KEY_VOLUMEDOWN //0x81 // Keyboard Volume Down
	// = 0x82  Keyboard Locking Caps Lock
	// = 0x83  Keyboard Locking Num Lock
	// = 0x84  Keyboard Locking Scroll Lock
	s2k["KEY_KPCOMMA"] = HID_KEY_KPCOMMA //0x85 // Keypad Comma
	// = 0x86  Keypad Equal Sign
	s2k["KEY_RO"] = HID_KEY_RO //0x87 // Keyboard International1
	s2k["KEY_KATAKANAHIRAGANA"] = HID_KEY_KATAKANAHIRAGANA //0x88 // Keyboard International2
	s2k["KEY_YEN"] = HID_KEY_YEN //0x89 // Keyboard International3
	s2k["KEY_HENKAN"] = HID_KEY_HENKAN //0x8a // Keyboard International4
	s2k["KEY_MUHENKAN"] = HID_KEY_MUHENKAN //0x8b // Keyboard International5
	s2k["KEY_KPJPCOMMA"] = HID_KEY_KPJPCOMMA //0x8c // Keyboard International6
	// = 0x8d  Keyboard International7
	// = 0x8e  Keyboard International8
	// = 0x8f  Keyboard International9
	s2k["KEY_HANGEUL"] = HID_KEY_HANGEUL //0x90 // Keyboard LANG1
	s2k["KEY_HANJA"] = HID_KEY_HANJA //0x91 // Keyboard LANG2
	s2k["KEY_KATAKANA"] = HID_KEY_KATAKANA //0x92 // Keyboard LANG3
	s2k["KEY_HIRAGANA"] = HID_KEY_HIRAGANA //0x93 // Keyboard LANG4
	s2k["KEY_ZENKAKUHANKAKU"] = HID_KEY_ZENKAKUHANKAKU //0x94 // Keyboard LANG5
	// = 0x95  Keyboard LANG6
	// = 0x96  Keyboard LANG7
	// = 0x97  Keyboard LANG8
	// = 0x98  Keyboard LANG9
	// = 0x99  Keyboard Alternate Erase
	// = 0x9a  Keyboard SysReq/Attention
	// = 0x9b  Keyboard Cancel
	// = 0x9c  Keyboard Clear
	// = 0x9d  Keyboard Prior
	// = 0x9e  Keyboard Return
	// = 0x9f  Keyboard Separator
	// = 0xa0  Keyboard Out
	// = 0xa1  Keyboard Oper
	// = 0xa2  Keyboard Clear/Again
	// = 0xa3  Keyboard CrSel/Props
	// = 0xa4  Keyboard ExSel

	// = 0xb0  Keypad 00
	// = 0xb1  Keypad 000
	// = 0xb2  Thousands Separator
	// = 0xb3  Decimal Separator
	// = 0xb4  Currency Unit
	// = 0xb5  Currency Sub-unit
	s2k["KEY_KPLEFTPAREN"] = HID_KEY_KPLEFTPAREN //0xb6 // Keypad (
	s2k["KEY_KPRIGHTPAREN"] = HID_KEY_KPRIGHTPAREN //0xb7 // Keypad )
	// = 0xb8  Keypad {
	// = 0xb9  Keypad }
	// = 0xba  Keypad Tab
	// = 0xbb  Keypad Backspace
	// = 0xbc  Keypad A
	// = 0xbd  Keypad B
	// = 0xbe  Keypad C
	// = 0xbf  Keypad D
	// = 0xc0  Keypad E
	// = 0xc1  Keypad F
	// = 0xc2  Keypad XOR
	// = 0xc3  Keypad ^
	// = 0xc4  Keypad %
	// = 0xc5  Keypad <
	// = 0xc6  Keypad >
	// = 0xc7  Keypad &
	// = 0xc8  Keypad &&
	// = 0xc9  Keypad |
	// = 0xca  Keypad ||
	// = 0xcb  Keypad :
	// = 0xcc  Keypad #
	// = 0xcd  Keypad Space
	// = 0xce  Keypad @
	// = 0xcf  Keypad !
	// = 0xd0  Keypad Memory Store
	// = 0xd1  Keypad Memory Recall
	// = 0xd2  Keypad Memory Clear
	// = 0xd3  Keypad Memory Add
	// = 0xd4  Keypad Memory Subtract
	// = 0xd5  Keypad Memory Multiply
	// = 0xd6  Keypad Memory Divide
	// = 0xd7  Keypad +/-
	// = 0xd8  Keypad Clear
	// = 0xd9  Keypad Clear Entry
	// = 0xda  Keypad Binary
	// = 0xdb  Keypad Octal
	// = 0xdc  Keypad Decimal
	// = 0xdd  Keypad Hexadecimal

	s2k["KEY_LEFTCTRL"] = HID_KEY_LEFTCTRL //0xe0 // Keyboard Left Control
	s2k["KEY_LEFTSHIFT"] = HID_KEY_LEFTSHIFT //0xe1 // Keyboard Left Shift
	s2k["KEY_LEFTALT"] = HID_KEY_LEFTALT //0xe2 // Keyboard Left Alt
	s2k["KEY_LEFTMETA"] = HID_KEY_LEFTMETA //0xe3 // Keyboard Left GUI
	s2k["KEY_RIGHTCTRL"] = HID_KEY_RIGHTCTRL //0xe4 // Keyboard Right Control
	s2k["KEY_RIGHTSHIFT"] = HID_KEY_RIGHTSHIFT //0xe5 // Keyboard Right Shift
	s2k["KEY_RIGHTALT"] = HID_KEY_RIGHTALT //0xe6 // Keyboard Right Alt
	s2k["KEY_RIGHTMETA"] = HID_KEY_RIGHTMETA //0xe7 // Keyboard Right GUI

	s2k["KEY_MEDIA_PLAYPAUSE"] = HID_KEY_MEDIA_PLAYPAUSE //0xe8
	s2k["KEY_MEDIA_STOPCD"] = HID_KEY_MEDIA_STOPCD //0xe9
	s2k["KEY_MEDIA_PREVIOUSSONG"] = HID_KEY_MEDIA_PREVIOUSSONG //0xea
	s2k["KEY_MEDIA_NEXTSONG"] = HID_KEY_MEDIA_NEXTSONG //0xeb
	s2k["KEY_MEDIA_EJECTCD"] = HID_KEY_MEDIA_EJECTCD //0xec
	s2k["KEY_MEDIA_VOLUMEUP"] = HID_KEY_MEDIA_VOLUMEUP //0xed
	s2k["KEY_MEDIA_VOLUMEDOWN"] = HID_KEY_MEDIA_VOLUMEDOWN //0xee
	s2k["KEY_MEDIA_MUTE"] = HID_KEY_MEDIA_MUTE //0xef
	s2k["KEY_MEDIA_WWW"] = HID_KEY_MEDIA_WWW //0xf0
	s2k["KEY_MEDIA_BACK"] = HID_KEY_MEDIA_BACK //0xf1
	s2k["KEY_MEDIA_FORWARD"] = HID_KEY_MEDIA_FORWARD //0xf2
	s2k["KEY_MEDIA_STOP"] = HID_KEY_MEDIA_STOP //0xf3
	s2k["KEY_MEDIA_FIND"] = HID_KEY_MEDIA_FIND //0xf4
	s2k["KEY_MEDIA_SCROLLUP"] = HID_KEY_MEDIA_SCROLLUP //0xf5
	s2k["KEY_MEDIA_SCROLLDOWN"] = HID_KEY_MEDIA_SCROLLDOWN //0xf6
	s2k["KEY_MEDIA_EDIT"] = HID_KEY_MEDIA_EDIT //0xf7
	s2k["KEY_MEDIA_SLEEP"] = HID_KEY_MEDIA_SLEEP //0xf8
	s2k["KEY_MEDIA_COFFEE"] = HID_KEY_MEDIA_COFFEE //0xf9
	s2k["KEY_MEDIA_REFRESH"] = HID_KEY_MEDIA_REFRESH //0xfa
	s2k["KEY_MEDIA_CALC"] = HID_KEY_MEDIA_CALC //0xfb

	return
}