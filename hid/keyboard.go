package hid

import (
	"encoding/json"
	"io"
	"log"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"fmt"
	"time"
	"math/rand"
	"regexp"
	"path/filepath"
	"sync"
	"context"
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
	lock *sync.Mutex
	DevicePath           string
	ActiveLanguageLayout *HIDKeyboardLanguageMap
	LanguageMaps         map[string]*HIDKeyboardLanguageMap //available language maps
	LEDWatcher           *KeyboardLEDStateWatcher
	KeyDelay             int
	KeyDelayJitter       int
	ctx context.Context
	cancel context.CancelFunc

	kbdOutFile *os.File
}




func NewKeyboard(ctx context.Context, devicePath string, resourcePath string) (keyboard *HIDKeyboard, err error) {
	//ToDo: check existence of deviceFile (+ is writable)

	ctx,cancel := context.WithCancel(ctx)

	keyboard = &HIDKeyboard{
		lock: &sync.Mutex{},
		DevicePath: devicePath,
		KeyDelay: 0,
		KeyDelayJitter: 0,
		ctx: ctx,
		cancel: cancel,
	}

	//Load available language maps
	err = keyboard.LoadLanguageMapDir(resourcePath)
	if err != nil {return nil, err}

	//Init LED sate
	keyboard.LEDWatcher, err = NewLEDStateWatcher(ctx, devicePath)
	if err != nil {return nil, err}

	//Open dev file for writing
	keyboard.kbdOutFile, err = os.OpenFile(devicePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil,err
	}


	return
}

func (kbd *HIDKeyboard) Close() {
	kbd.kbdOutFile.Close()
	kbd.LEDWatcher.Stop()
}


func (kbd *HIDKeyboard) LoadLanguageMapDir(dirpath string) (err error) {
	folder,err := filepath.Abs(dirpath)
	if err != nil { return err }
	var mapFiles []string
	err = filepath.Walk(string(folder),  func(path string, info os.FileInfo, err error) error {
		if err != nil { return err } // prevent panic due to access failures
		if !info.IsDir() && (strings.ToLower(filepath.Ext(info.Name())) == ".json") {
			fp := path
			abs,pErr := filepath.Abs(fp)
			if pErr == nil {
				mapFiles = append(mapFiles, abs)
			} else {
				//print warning
				log.Printf("Ignoring map file '%s', retrieving absolute path failed!\n", fp)
			}
		}
		return nil
	})
	if err != nil { return err }

	//mapFiles contains all absolute path of files with extension ".json"
	var commonMAP *HIDKeyboardLanguageMap
	for _,mapFile := range mapFiles {
		kbdmap, lErr := loadKeyboardLanguageMapFromFile(mapFile)
		if lErr != nil {
			//Warn on error
			log.Printf("Skipping language map file '%s' due to load error: %v\n", mapFile, lErr)
			continue
		}
		if strings.ToUpper(kbdmap.Name) == "COMMON" {
			//this is the map with common keys
			//mapping in this file will be reflected to all other maps, in case the ARE NOT ALREADY PRESENT
			commonMAP = kbdmap
		} else {
			if kbd.LanguageMaps == nil {
				kbd.LanguageMaps = make(map[string]*HIDKeyboardLanguageMap)
			}
			kbd.LanguageMaps[strings.ToUpper(kbdmap.Name)] = kbdmap

			if kbd.ActiveLanguageLayout == nil && kbdmap.Name != "COMMON" {
				kbd.ActiveLanguageLayout = kbdmap
			}
		}
	}

	// If no language maps beside "COMMON" have been loaded, return error
	if len(kbd.LanguageMaps) == 0 {
		if commonMAP == nil {
			return errors.New("Couldn't load any language map")
		} else {
			return errors.New("Couldn't load any language map, beside 'COMMON' map")
		}

	}

	// At this point, all map files not named "COMMON" should be added to kbd.LanguageMaps
	// In case a map with name "COMMON" was found, it should be stored in commonMap
	// If commonMap was found, the contained mappings are added to the other language maps,
	// in case the dedicated mapping doesn't exist already. F.e. the mapping for "F1" is only
	// needed in map "COMMON" and added to all other maps, except they specify the "F1" mapping
	// themselves.
	if commonMAP != nil {
		//iterate over all common mappings
		for name,reports := range commonMAP.Mapping {
			//iterate over all loaded maps
			for _,lMap := range kbd.LanguageMaps {
				//check if the mapping isn't already present and add it if needed
				if _,alreadyPresent := lMap.Mapping[name]; !alreadyPresent {
					lMap.Mapping[name] = reports
				}
			}
		}
	}

	return
}

func (kbd *HIDKeyboard) LoadLanguageMapFromFile(filepath string) (err error) {
	//if this is the first map loaded, set as active Map
	kbdmap, err := loadKeyboardLanguageMapFromFile(filepath)
	if err != nil { return err }

	if kbd.LanguageMaps == nil {
		kbd.LanguageMaps = make(map[string]*HIDKeyboardLanguageMap)
	}
	kbd.LanguageMaps[strings.ToUpper(kbdmap.Name)] = kbdmap

	if kbd.ActiveLanguageLayout == nil && kbdmap.Name != "COMMON" {
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

func (kbd *HIDKeyboard) StringToPressKeyCombo(comboStr string) (err error) {
	report,err := kbd.StringToKeyCombo(comboStr)
	if err != nil { return err }
	seq := []KeyboardOutReport{*report} //convert to single report sequence
	err = kbd.PressKeySequence(seq)
	return
}


func (kbd *HIDKeyboard) StringToKeyCombo(comboStr string) (result *KeyboardOutReport, err error) {
	//ToDo: Check if keyboard device file exists
	if kbd.ActiveLanguageLayout == nil {
		return nil,errors.New("No language mapping active, couldn't send key combo!")
	}

	//split key sequence describe by string into single key names
	keyNames := rpSplit.Split(comboStr, -1)
	if len(keyNames) == 0 {
		return nil,errors.New("No keys to press")
	}
	if len(keyNames) == 1 && len(keyNames[0]) == 0 {
		return nil,errors.New("No keys to press")
	}

	//fmt.Printf("KeyNames %d: %+v\n", len(keyNames), keyNames)

	//try to convert splitted keynames to reports
	comboReports := make([]*KeyboardOutReport,0)
	for _,keyname := range keyNames {
		if len(keyname) == 0 { continue } //ignore empty keynames
		report,err := kbd.mapKeyNameToReport(keyname)
		if err == nil {
			//fmt.Printf("Keyname '%s' mapped to report %+v\n", keyname, report)
			comboReports = append(comboReports, report)
		} else {
			return nil,errors.New(fmt.Sprintf("Couldn't build key combo '%s' because of mapping error in keyname '%s': %v", comboStr, keyname, err))
		}
	}

	//fmt.Printf("Combo reports for '%s': %+v\n", comboStr, comboReports)

	//combine reports
	result,err = combineReports(comboReports)
	return
}

func (kbd *HIDKeyboard) StringToPressKeySequence(str string) (err error) {

	//ToDo: Check if keyboard device file exists
	if kbd.ActiveLanguageLayout == nil {
		return errors.New("No language mapping active, couldn't send key sequence!")
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

// mapKeyNameToReport tries to translate a key expressed by a string description to a SINGLE
// report (with respect to the chosen language map), which could be sent to a keyboard device.
// Most printable characters like 'a' or 'A' could be represented by a single rune (f.e. `a` or `A`).
// The parameter `keyName` is of type string (instead of rune) because there're keys which
// couldn't be described with a single rune, for example: 'F1', 'ESCAPE' ...
//
// mapKeyNameToReport translates uppercase alphabetical keys [A-Z] to the respective lower case,
// before trying to map, in order to avoid fetching reports with [SHIFT] modifiers (this assures
// that 'A' gets mapped to the USB key KEY_A, not to the USB key KEY_A combined with the [SHIFT]
// modifier).
//
// There're result reports containing only modifiers (f.e. if keyName = 'CTRL'). Such reports could be used
// to build KeyCombos, by mixing them together.
//
// The language maps consist mappings from UTF-8 runes to reports sequences and keyNames to single reports.
// Examples for sequences are mostly printable runes which are built from multiple sequential key presses
// (mostly started with DEAD KEYS) like the `^` rune on a german keyboard layout, which has to be created by
// pressing [^] followed by [SPACE].
// The purpose of mapKeyNameToReport is to resolve the given keyname to A SINGLE REPORT, thus report sequences
// are truncated to the first report only, before returned. This is the trade-off between assuring to return a
// single report per keyname versus managing separated mapping files for rune-mapping and key-mapping.
func (kbd *HIDKeyboard) mapKeyNameToReport(keyName string) (report *KeyboardOutReport,err error) {
	// Assure keyName contains no spaces, else error
	r := rpSplit.Split(keyName, -1)
	if len(r) > 1 {
		return nil, errors.New(fmt.Sprintf("Error mapping keyName '%s', unallowed contains spaces!", keyName))
	}
	keyName = strings.ToUpper(r[0]) //reassign trimmed, upper case version


	// If keyName consists of a single upper case letter, translate to lowercase
	if rpSingleUpperLetter.MatchString(keyName) { keyName = strings.ToLower(keyName)}

	// Try to find a matching mapping in current language map
	if kbd.ActiveLanguageLayout == nil {
		return nil, errors.New("No language mapping selected")
	}
	if reports,found := kbd.ActiveLanguageLayout.Mapping[keyName]; found {
		//report(s) found, return only first one
		if len(reports) > 0 {
			return &reports[0], nil
		} else {
			return nil, errors.New(fmt.Sprintf("Mapping for key '%s' found in language map named '%s', but mapping is empty!", keyName, kbd.ActiveLanguageLayout.Name))
		}
	} else {
		return nil, errors.New(fmt.Sprintf("Couldn't find mapping for key '%s' in language map named '%s'!", keyName, kbd.ActiveLanguageLayout.Name))
	}
}

// combineReports combines a slice of output reports into a single report (for key combinations).
// The following rules apply:
// 1) Modifiers are combined with logical or
// 2) Unique keys are filled into the keys array, one-by-one
// 3) Duplicated keys are ignore
// 4) Only the first 6 keys are regarded (without duplicates), the rest is ignored
func combineReports(reports []*KeyboardOutReport) (result *KeyboardOutReport, err error) {
	r := KeyboardOutReport{}
	keys := make(map[byte]bool)
	keyCount := 0
	maxKeys := 6

	ADDREPORTLOOP:
	for _,report := range reports {
		//Add modifiers
		r.Modifiers |= report.Modifiers

		// add keys to map
		// Note: This could be interrupted in the middle of a report if too many keys are contained, while the
		// modifiers of this report are already applied. This is a corner case, we don't take care of (happens f.e.
		// if the first report contains 2 keys and the second one 6 keys with modifiers)
		for _,key := range report.Keys {
			if key != 0 { //Ignore "no key"
				if !keys[key] {
					keys[key] = true
					keyCount++
					if keyCount >= maxKeys {
						break ADDREPORTLOOP
					}
				}
			}
		}
	}

	//keys should contain maxKeys at max
	keyCount = 0
	for k,_ := range keys {
		r.Keys[keyCount] = k
		keyCount++
		if keyCount >= maxKeys { break }
	}

	return &r, nil
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
	//Synchronize the whole sequence output, as this is considered atomar.
	// f.e. This is used to type out character which are brought up a sequence consisting of
	// [some deadkey + modifiers, some normal key + modifier, zero report for key release]
	// if another key is pressed right after the deadkey, by a different go routine, we end
	// up with unpredictable output
	kbd.lock.Lock()
	defer kbd.lock.Unlock()

	//iterate over reports and send them
	for _,rep := range reports {
		//err = rep.WriteTo(kbd.DevicePath)
		err = rep.WriteToFile(kbd.kbdOutFile)
		if err != nil { return err }
	}
	//append an empty report to release all keys
	//err = KeyboardReportEmpty.WriteTo(kbd.DevicePath)
	err = KeyboardReportEmpty.WriteToFile(kbd.kbdOutFile)
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

func loadKeyboardLanguageMapFromFile(filePath string) (result *HIDKeyboardLanguageMap, err error) {
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
	var o map[string]interface{}
	if err := json.Unmarshal(b,&o); err != nil {
		return err
	}


	for k,v := range o {
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

func (kr KeyboardOutReport) String() string {
	bytes,err := kr.MarshalJSON()
	if err == nil {
		return string(bytes)
	} else {
		return fmt.Sprintf("%+v", kr) //ToDo: check if this works or calls a loop
	}
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

// Accepts *os.File, in contrast to WriteTo() this allows keeping the file open
func (rep KeyboardOutReport) WriteToFile(file *os.File) (err error) {
	data := rep.Serialize()
	n, err := file.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	return err
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

