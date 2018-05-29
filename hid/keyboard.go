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
)

var (
	KeyboardReportEmpty = NewKeyboardOutReport(0)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type HIDKeyboard struct {
	DevicePath string
	ActiveLanguageLayout *HIDKeyboardLanguageMap
	LanguageMaps map[string]*HIDKeyboardLanguageMap //available language maps
	KeyDelay int
	KeyDelayJitter int
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
			err = kbd.PressKeyCombo(reports)
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

func (kbd *HIDKeyboard) PressKeyCombo(reports []KeyboardOutReport) (err error) {
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
