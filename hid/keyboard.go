package hid

import (
	"encoding/json"
	"log"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"fmt"
	"time"
	"math/rand"
	"regexp"
)

var (
	KeyboardReportEmpty = NewKeyboardOutReport(0)
	ErrTimeout = errors.New("Timeout reached")
)

var (
	//regex
	rpSplit = regexp.MustCompile("(?m)\\s+")
	rpSingleUpperLetter = regexp.MustCompile("(?m)^[A-Z]$")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type HIDKeyboard struct {
	DevicePath           string
	ActiveLanguageLayout *HIDKeyboardLanguageMap
	LanguageMaps         map[string]*HIDKeyboardLanguageMap //available language maps
	LEDWatcher           *HIDKeyboardLEDStateWatcher
	KeyDelay             int
	KeyDelayJitter       int
}




func NewKeyboard(devicePath string, resourcePath string) (keyboard *HIDKeyboard, err error) {
	keyboard = &HIDKeyboard{}
	keyboard.DevicePath = devicePath
	keyboard.KeyDelay = 0
	keyboard.KeyDelayJitter = 0

	//ToDo: Load whole language map folder, for now single layout testing
	err = keyboard.LoadLanguageMapFromFile(resourcePath + "/common.json")
	if err != nil {return nil, err}
	err = keyboard.LoadLanguageMapFromFile(resourcePath + "/DE_ASCII.json")
	if err != nil {return nil, err}

	//Init LED sate
	keyboard.LEDWatcher, err = newHIDKeyboardLEDStateWatcher(devicePath)
	if err != nil {return nil, err}

	return
}


func (kbd *HIDKeyboard) LoadLanguageMapFromFile(filepath string) (err error) {
	//if this is the first map loaded, set as active Map
	kbdmap, err := LoadKeyboardLanguageMapFromFile(filepath)
	if err != nil { return err }

	if kbd.LanguageMaps == nil {
		kbd.LanguageMaps = make(map[string]*HIDKeyboardLanguageMap)
	}
	kbd.LanguageMaps[strings.ToUpper(kbdmap.Name)] = kbdmap

	if kbd.ActiveLanguageLayout == nil {
		kbd.ActiveLanguageLayout = kbdmap
	}

	return nil
}

func (kbd HIDKeyboard) ListLanguageMapNames() (mapNames []string) {
	mapNames = make([]string, len(kbd.LanguageMaps))

	i := 0
	for k := range kbd.LanguageMaps {
		mapNames[i] = k
		i++
	}
	return mapNames
}

func (kbd *HIDKeyboard) SetActiveLanguageMap(name string) (err error) {
	if m, ok := kbd.LanguageMaps[strings.ToUpper(name)]; ok {
		kbd.ActiveLanguageLayout = m
	} else {
		return errors.New(fmt.Sprintf("Language map with name '%s' isn't loaded!", name))
	}
	return nil
}

func (kbd *HIDKeyboard) SendString(str string) (err error) {
	//ToDo: Check if keyboard device file exists
	if kbd.ActiveLanguageLayout == nil {
		return errors.New("No language mapping active, couldn't send string!")
	}
	for _,runeVal := range str {
		strRune := string(runeVal)
		if reports,found := kbd.ActiveLanguageLayout.Mapping[strRune]; found {
			//log.Printf("Sending reports (%T): %v\n", reports, reports)
			err = kbd.PressKeySequence(reports)
			if err != nil {
				//Abort typing
				return err
			}
		} else {
			log.Printf("HID keyboard warning: Couldn't send charcter '%q' (0x%x) because it is not defined in language map '%s', skipping ...", strRune, strRune, kbd.ActiveLanguageLayout.Name)
		}


	}
	return nil
}

// mapKeyStringToReports tries to translate a key expressed by a string description to a SINGLE
// report (with respect to the chosen language map), which could be sent to a keyboard device.
// Most printable characters like 'a' or 'A' could be represented by a single rune (f.e. `a` or `A`).
// mapKeyStringToReports translates uppercase alphabetical keys [A-Z] to the respective lower case,
// before trying to map, in order to avoid fetching reports with [SHIFT] modifiers (this assures
// that 'A' gets mapped to the USB key KEY_A, not to the USB key KEY_A combined with the [SHIFT]
// modifier).
// The parameter `keyDescription` is of type string (instead of rune) because there're keys which
// couldn't be described with a single rune, for example: 'F1', 'ESCAPE' ...
//
// mapKeyStringToReports could return a report, containing only modifiers (f.e. if
// keyDescription = 'CTRL'). Such reports could be used to build KeyCombos, by mixing them together.
//
// The language maps could contain mappings, containing multiple reports, as they sometime represent
// printable runes consisting of a sequence of multiple keys (f.e. `^` in the German layout maps
// to a report slice of the key for [^] followed by the key [SPACE], which is needed to print the character).
// mapKeyStringToReports returns ONLY THE FIRST REPORT of such slices, as this is closer to the representation.
// Additionally, single reports could be combimned into a key-combo, which wouldn't be possible with a ordered
// sequence of reports.
func (kbd *HIDKeyboard) TmapKeyStringToReports(keyDescription string) (report *KeyboardOutReport,err error) {
	// Assure keyDescription contains no spaces, else error
	r := rpSplit.Split(keyDescription, -1)
	if len(r) > 1 {
		return nil, errors.New("keyDescription mustn't contain spaces")
	}
	keyDescription = strings.ToUpper(r[0]) //reassign trimmed, upper case version


	// If keyDescription consists of a single upper case letter, translate to lowercase
	if rpSingleUpperLetter.MatchString(keyDescription) { keyDescription = strings.ToLower(keyDescription)}

	// Try to find a matching mapping in 1) current language map, followed by 2) common map (the latter
	// holds mappings like 'F1', 'CTRL' etc which are more or less language independent. The common
	// map is only accessed, if there was no successful mapping in the chosen language map (priority
	// as more specialized)
	c,ok := kbd.LanguageMaps["COMMON"]
	if !ok { return nil,errors.New("Keyboardmap 'common' not found")}
	common := c.Mapping

	reports := common[keyDescription]
	if len(reports) > 1 {
		//store first report as result
		report = &reports[0]
	}

	fmt.Printf("Descr '%s': %+v\n", keyDescription, )
	return
}

// PressKeySequence writes the output reports given in `reports` to the keyboard device in sequential
// order. A all empty report is automatically appended in order to release all keys after finishing
// the sequence (press in contrast to hold).
//
// There's a clear reason to use sequences of reports. For example the character 'à' on a German keyboard
// layout, is created by pressing the key with [`] in combination with [SHIFT], followed by the key [A].
// To represent the underlying key sequence three reports are needed:
// 1) A report containing the key [`] (key equal from US layout) along with the [SHIFT] modifier
// 2) A report containing the key [A] (the [A] key results in lower case 'a', as no [SHIFT] modifier is used)
// 3) A report containing no key and no modifier representing the release of all keys
//
// It is worth mentioning, that a single report could hold 8 different modifiers and up to 6 dedicated keys,
// for the keyboard type used here. Anyway, packing the keys [A] and [`] in a single report, along with the
// [SHIFT] modifier, would lead to a different result (if there's a result at all). The reason is, that the
// pressing order of the two keys [A] and [`] couldn't be determined anymore, neither would it be possible
// to distinguish if the [SHIFT] modifier should be combined with [A], [`] or both.
//
// As shown, even a single character could be represented by several reports in sequential order! This is why
// PressKeySequence is needed.
//
// A key combination, in contrast to a sequence, combines several keys in a single report (f.e. CTRL+ALT+A)
func (kbd *HIDKeyboard) PressKeySequence(reports []KeyboardOutReport) (err error) {
	//iterate over reports and send them
	for _,rep := range reports {
		err = rep.WriteTo(kbd.DevicePath)
		if err != nil { return err }
	}
	//append an empty report to release all keys
	err = KeyboardReportEmpty.WriteTo(kbd.DevicePath)
	if err != nil { return err }

	//Delay after keypress
	delay := kbd.KeyDelay
	if kbd.KeyDelayJitter > 0 { delay += rand.Intn(kbd.KeyDelayJitter)}
	if delay > 0 { time.Sleep(time.Millisecond * time.Duration(delay)) }

	return nil
}



type HIDKeyboardLanguageMap struct {
	Name string
	Description string
	Mapping map[string][]KeyboardOutReport
}

func (klm *HIDKeyboardLanguageMap) StoreToFile(filePath string) (err error) {
	//create JSON representation
	mapJson, err := json.MarshalIndent(klm, "", "\t")
	if err != nil { return err }
	//Write to file
	return ioutil.WriteFile(filePath, mapJson, os.ModePerm)
}

func LoadKeyboardLanguageMapFromFile(filePath string) (result *HIDKeyboardLanguageMap, err error) {
	result = &HIDKeyboardLanguageMap{}
	mapJson, err := ioutil.ReadFile(filePath)
	if err != nil { return nil,err }
	err = json.Unmarshal(mapJson, result)
	return
}

type KeyboardOutReport struct {
	Modifiers byte
	//Reserved byte
	Keys [6]byte
}


func (kr *KeyboardOutReport) UnmarshalJSON(b []byte) error {
	var o interface{}
	if err := json.Unmarshal(b,&o); err != nil {
		return err
	}

	m := o.(map[string]interface{})
	for k,v := range m {
		//log.Printf("key: %v, val %v (%T)\n", k, v, v)

		switch strings.ToLower(k) {
		case "modifiers":
			switch vv := v.(type) {
			case []interface{}:
				for _, modValIface := range vv {
					//convert modifier back from string to uint8 representation

					switch modVal := modValIface.(type) {
					case string:
						if modInt, ok := StringToUsbModKey[modVal]; ok {
							kr.Modifiers |= modInt
						} else {
							return errors.New(fmt.Sprintf("The value '%s' couldn't be translated to a valid modifier key\n", modVal))
						}
						//log.Printf("Mod: %v (%T)", modVal, modVal)
					case float64:
						modInt := uint8(modVal)
						if _,ok := UsbModKeyToString[modInt]; !ok && modInt != 0 {
							return errors.New(fmt.Sprintf("The value '%v' isn't valid for a modifier key\n", modVal))
						}

						kr.Modifiers |= modInt
					default:
						return errors.New(fmt.Sprintf("The value '%v' of type '%T' isn't a valid type for a modifier key\n", modVal, modVal))
					}
				}
			default:
				return errors.New(fmt.Sprintf("Unintended type for 'Modifiers', has to be array of modifier strings, but %v was given\n", vv))
			}
		case "keys":
			switch vv := v.(type) {
			case []interface{}:
				for i, keyValIface := range vv {
					if i > len(kr.Keys) - 1 {
						return errors.New(fmt.Sprintf("The key '%v' at index %d exceeds the maximum key count per report, which is 6!\n", keyValIface, i))
					}
					switch keyVal := keyValIface.(type) {
					case string:

						if keyInt, ok := StringToUsbKey[keyVal]; ok {
							kr.Keys[i] = keyInt
						} else {
							return errors.New(fmt.Sprintf("The value '%s' couldn't be translated to a valid key\n", keyVal))
						}

						//log.Printf("Key '%s' (%T) at index %d\n", keyVal, keyVal, i)
					case float64:
						keyInt := uint8(keyVal)
						if _,ok := UsbKeyToString[keyInt]; !ok && keyInt != 0 {
							return errors.New(fmt.Sprintf("The value '%v' isn't valid for a key\n", keyVal))
						}

						kr.Keys[i] = keyInt
					default:
						return errors.New(fmt.Sprintf("The value '%v' of type '%T' at index %d isn't a valid type for a key array\n", keyVal, keyVal, i))
					}
				}
			default:
				return errors.New(fmt.Sprintf("Unintended type in for 'Keys', has to be array of key strings, but %v was given\n", vv))
			}
		}
	}


	return nil
}


func (kr *KeyboardOutReport) MarshalJSON() ([]byte, error) {
	keys := []string{}
	modifiers := []string{}

	if kr.Modifiers & HID_MOD_KEY_LEFT_CONTROL > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_LEFT_CONTROL]) }
	if kr.Modifiers & HID_MOD_KEY_LEFT_SHIFT > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_LEFT_SHIFT]) }
	if kr.Modifiers & HID_MOD_KEY_LEFT_ALT > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_LEFT_ALT]) }
	if kr.Modifiers & HID_MOD_KEY_LEFT_GUI > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_LEFT_GUI]) }
	if kr.Modifiers & HID_MOD_KEY_RIGHT_CONTROL > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_RIGHT_CONTROL]) }
	if kr.Modifiers & HID_MOD_KEY_RIGHT_SHIFT > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_RIGHT_SHIFT]) }
	if kr.Modifiers & HID_MOD_KEY_RIGHT_ALT > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_RIGHT_ALT]) }
	if kr.Modifiers & HID_MOD_KEY_RIGHT_GUI > 0 { modifiers = append(modifiers, UsbModKeyToString[HID_MOD_KEY_RIGHT_GUI]) }

	for _,key := range kr.Keys {
		if key == 0 {break} //abort on first 0x00 key code
		if keyStr, ok := UsbKeyToString[uint8(key)]; ok {
			keys = append(keys, keyStr)
			//log.Println(keyStr)
		} else {
			log.Printf("Warning: No string representation for USB key with value '%d', key ignored during JSON marshaling.\n", key)
		}
	}

	result := struct{
		Modifiers []string
		Keys []string
	}{
		Keys:keys,
		Modifiers:modifiers,
	}
	return json.Marshal(result)
}



func (rep KeyboardOutReport) Serialize() (out []byte) {
	out = []byte {
		rep.Modifiers,
		byte(0),
		rep.Keys[0],
		rep.Keys[1],
		rep.Keys[2],
		rep.Keys[3],
		rep.Keys[4],
		rep.Keys[5],
	}
	return
}

func (rep KeyboardOutReport) Deserialize(data []byte) (err error) {
	if len(data) != 8 {
		err = errors.New("Wrong data length, keyboard out report has to be 8 bytes in length")
	}
	rep = KeyboardOutReport{
		Modifiers: data[0],
		//data[1] should be empty, we ignore it
		Keys: [6]byte{
			data[2],
			data[3],
			data[4],
			data[5],
			data[6],
			data[7],
		},
	}
	return
}

func (rep KeyboardOutReport) WriteTo(filePath string) (err error) {
	return ioutil.WriteFile(filePath, rep.Serialize(), os.ModePerm) //Serialize Report and write to specified file
}



func NewKeyboardOutReport(modifiers byte, keys ...byte) (res KeyboardOutReport) {
	res = KeyboardOutReport{
		Keys: [6]byte {0, 0, 0, 0, 0, 0,},
	}
	res.Modifiers = modifiers
	for i, key := range keys {
		if i < 6 {
			res.Keys[i] = key
		}
	}
	return
}
