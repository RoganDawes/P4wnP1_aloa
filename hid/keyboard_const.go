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
