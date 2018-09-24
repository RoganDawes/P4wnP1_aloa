// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/mame82/hvue"
)

func InitComponentsWiFi() {
	hvue.NewComponent(
		"wifi",
		hvue.Template(templateWiFi),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowStoreModal bool   `js:"showStoreModal"`
				ShowLoadModal bool   `js:"showLoadModal"`
				TemplateName   string `js:"templateName"`
			}{Object: O()}
			data.ShowStoreModal = false
			data.ShowLoadModal = false
			data.TemplateName = ""
			return &data
		}),
		hvue.Computed("settings", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state").Get("wifiState").Get("CurrentSettings")
		}),
		hvue.Computed("wifiState", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state").Get("wifiState")
		}),
		hvue.Computed("wifiStateIcon", func(vm *hvue.VM) interface{} {
			mode := vm.Get("$store").Get("state").Get("wifiState").Get("mode").Int()
			switch mode {
			case int(pb.WiFiStateMode_STA_CONNECTED):
				// Client mode, connected
				return "wifi"
			case int(pb.WiFiStateMode_AP_UP):
				// Access Point running
				return "wifi_tethering"
			default:
				return "portable_wifi_off"
			}
		}),
		hvue.Computed("wifiStateText", func(vm *hvue.VM) interface{} {
			mode := vm.Get("$store").Get("state").Get("wifiState").Get("mode").Int()
			ssid := vm.Get("$store").Get("state").Get("wifiState").Get("ssid").String()
			channel := vm.Get("$store").Get("state").Get("wifiState").Get("channel").String()
			switch mode {
			case int(pb.WiFiStateMode_STA_CONNECTED):
				res := "Connected to network"
				if len(ssid) > 0 {
					res += ": '" + ssid + "' on channel " + channel
				}
				return res
			case int(pb.WiFiStateMode_AP_UP):
				res := "Access Point running on channel " + channel
				if len(ssid) > 0 {
					res += ": '" + ssid + "'"
				}
				return res
			default:
				return "Not connected"
			}
		}),
		hvue.ComputedWithGetSet("enabled",
			func(vm *hvue.VM) interface{} {
				return !vm.Get("settings").Get("disabled").Bool()
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Get("settings").Set("disabled", !newValue.Bool())
			},
		),
		hvue.ComputedWithGetSet("enableNexmon",
			func(vm *hvue.VM) interface{} {
				return vm.Get("settings").Get("nexmon").Bool()
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Get("settings").Set("nexmon", newValue.Bool())
			},
		),

		hvue.Computed("wifiAuthModes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val, _ := range pb.WiFiAuthMode_name {
				mode := struct {
					*js.Object
					Label string `js:"label"`
					Value int    `js:"value"`
				}{Object: O()}
				mode.Value = val
				switch pb.WiFiAuthMode(val) {
				case pb.WiFiAuthMode_WPA2_PSK:
					mode.Label = "WPA2"
				case pb.WiFiAuthMode_OPEN:
					mode.Label = "Open"
				default:
					mode.Label = "Unknown"
				}
				modes.Call("push", mode)
			}
			return modes
		}),
		hvue.Computed("wifiModes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val, _ := range pb.WiFiWorkingMode_name {
				mode := struct {
					*js.Object
					Label string `js:"label"`
					Value int    `js:"value"`
				}{Object: O()}
				mode.Value = val
				switch pb.WiFiWorkingMode(val) {
				case pb.WiFiWorkingMode_AP:
					mode.Label = "Access Point (AP)"
				case pb.WiFiWorkingMode_STA:
					mode.Label = "Station (Client)"
				case pb.WiFiWorkingMode_STA_FAILOVER_AP:
					mode.Label = "Client with Failover to AP"
				default:
					continue
				}
				modes.Call("push", mode)
			}
			return modes
		}),
		hvue.Computed("mode_ap", func(vm *hvue.VM) interface{} { return pb.WiFiWorkingMode_AP }),
		hvue.Computed("mode_sta", func(vm *hvue.VM) interface{} { return pb.WiFiWorkingMode_STA }),
		hvue.Computed("mode_failover", func(vm *hvue.VM) interface{} { return pb.WiFiWorkingMode_STA_FAILOVER_AP }),
		hvue.Method("reset",
			func(vm *hvue.VM) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_WIFI_STATE)
			}),
		hvue.Method("deploy",
			func(vm *hvue.VM, wifiSettings *jsWiFiSettings) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_WIFI_SETTINGS, wifiSettings)
			}),
		hvue.Method("updateStoredSettingsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST)
			}),
		hvue.Method("store",
			func(vm *hvue.VM) {
				sReq := NewWifiRequestSettingsStorage()
				sReq.TemplateName = vm.Get("templateName").String()
				sReq.Settings = &jsWiFiSettings{
					Object: vm.Get("$store").Get("state").Get("wifiState").Get("CurrentSettings"),
				}
				println("Storing :", sReq)
				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_WIFI_SETTINGS, sReq)
				vm.Set("showStoreModal", false)
			}),
		hvue.Computed("deploying",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("deployingWifiSettings")
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST)
		}),

	)
}

const templateWiFi = `
<q-page padding>


<div class="row gutter-sm">
	<q-modal v-model="showLoadModal">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					Load WiFi settings
				</q-toolbar-title>
			</q-toolbar>

			<q-list>
				<q-item link tag="label" v-for="tname in this.$store.state.StoredWifiSettingsList" :key="tname">
					<q-item-side>
						<q-radio v-model="templateName" :val="tname"/>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>{{ tname }}</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="secondary" v-close-overlay label="close" />
							<q-btn color="primary" label="load" />
							<q-btn color="primary" label="deploy" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>
	<q-modal v-model="showStoreModal">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					Store current WiFi Settings
				</q-toolbar-title>
			</q-toolbar>
			<q-list>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Template name</q-item-tile>
						<q-item-tile>
							<q-input v-model="templateName" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="secondary" v-close-overlay label="close" />
							<q-btn color="primary" @click="store" label="store" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>
	<div class="col-lg-4">
	<q-card class="full-height">
		<q-card-title>
			WiFi settings
		</q-card-title>

		<q-item>
        	<q-item-side :icon="wifiStateIcon" color="primary"></q-item-side>
			<q-item-main>
				<q-item-tile label>{{ wifiStateText }}</q-item-tile>
			</q-item-main>
		</q-item>
	
		<q-card-actions>
			<q-btn :loading="deploying" color="primary" @click="deploy(settings)" label="deploy"></q-btn>
			<q-btn color="secondary" @click="reset" label="reset"></q-btn>
			<q-btn color="primary" @click="updateStoredSettingsList(); showStoreModal=true" label="store"></q-btn>
			<q-btn color="primary" @click="updateStoredSettingsList(); showLoadModal=true" label="load"></q-btn>
		</q-card-actions>

		<q-list link>
			<q-item-separator />
			<q-list-header>Generic</q-list-header>
			<q-item tag="label">
				<q-item-side>
					<q-toggle v-model="enabled"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>Enabled</q-item-tile>
					<q-item-tile sublabel>Enable/Disable WiFi</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label">
				<q-item-side>
					<q-toggle v-model="enableNexmon"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>Nexmon</q-item-tile>
					<q-item-tile sublabel>Enable/Disable modified nexmon firmware (needed for WiFi covert channel and KARMA)</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label">
				<q-item-main>
					<q-item-tile label>Regulatory domain</q-item-tile>
					<q-item-tile sublabel>Regulatory domain according to ISO/IEC 3166-1 alpha2 (example "US")</q-item-tile>
					<q-item-tile>
						<q-input v-model="settings.reg" inverted></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label">
				<q-item-main>
					<q-item-tile label>Working Mode</q-item-tile>
					<q-item-tile sublabel>Work as Access Point or Client</q-item-tile>
					<q-item-tile>
						<q-select v-model="settings.mode" :options="wifiModes" color="secondary" inverted></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

		</q-list>
	</q-card>
	</div>

	<div class="col-lg-4" v-if="settings.mode == mode_sta || settings.mode == mode_failover">
	<q-card class="full-height">
		<q-card-title>
			WiFi client settings
		</q-card-title>

		<q-list link>
				<q-item-separator />
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>SSID</q-item-tile>
						<q-item-tile sublabel>Network name to connect</q-item-tile>
						<q-item-tile>
							<q-input v-model="settings.staBssList[0].ssid" color="primary" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Pre shared key</q-item-tile>
						<q-item-tile sublabel>If empty, a network with Open Authentication is assumed (Warning: PLAIN TRANSMISSION)</q-item-tile>
						<q-item-tile>
							<q-input v-model="settings.staBssList[0].psk" type="password" color="primary" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>

			<template v-if="settings.mode == mode_failover">
				<q-item>
					<q-item-main>
	  				<q-alert type="warning">
						If the SSID provided for client mode couldn't be connected, an attempt is started to fail over to Access Point mode with the respective settings.
					</q-alert>
					</q-item-main>
				</q-item>
			</template>
		</q-list>
	</q-card>
	</div>

	<div class="col-lg-4" v-if="settings.mode == mode_ap || settings.mode == mode_failover">
	<q-card class="full-height">
		<q-card-title>
			WiFi Access Point settings
		</q-card-title>

		<q-list link>


			<template>
				<q-item-separator />
				<q-list-header>Access Point settings</q-list-header>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Channel</q-item-tile>
						<q-item-tile sublabel>Must exist in regulatory domain (example 13)</q-item-tile>
						<q-item-tile>
							<q-input v-model="settings.channel" type="number" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>

				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Authentication Mode</q-item-tile>
						<q-item-tile sublabel>Authentication Mode for Access Point (ignored for client mode)</q-item-tile>
						<q-item-tile>
							<q-select v-model="settings.authMode" :options="wifiAuthModes" color="primary" inverted></q-select>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>SSID</q-item-tile>
						<q-item-tile sublabel>Network name (Service Set Identifier)</q-item-tile>
						<q-item-tile>
							<q-input v-model="settings.apBss.ssid" color="primary" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="settings.hideSsid"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Hide SSID</q-item-tile>
						<q-item-tile sublabel>Access Point doesn't send beacons with its SSID</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Pre shared key</q-item-tile>
						<q-item-tile sublabel>Warning: PLAIN TRANSMISSION</q-item-tile>
						<q-item-tile>
							<q-input v-model="settings.apBss.psk" type="password" color="primary" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</template>
		</q-list>
	</q-card>
	</div>


</div>
</q-page>	

`
