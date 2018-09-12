// +build js

package main

import (
	"github.com/HuckRidgeSW/hvue"
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
)

func InitComponentsWiFi() {
	hvue.NewComponent(
		"wifi",
		hvue.Template(templateWiFi),
		hvue.Computed("settings", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state").Get("wifiSettings")
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
			for val,_ := range pb.WiFiAuthMode_name {
				mode := struct {
					*js.Object
					Label string `js:"label"`
					Value int  `js:"value"`
				}{Object:O()}
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
			for val,_ := range pb.WiFiWorkingMode_name {
				mode := struct {
					*js.Object
					Label string `js:"label"`
					Value int  `js:"value"`
				}{Object:O()}
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
		hvue.Computed("mode_ap", func(vm *hvue.VM) interface{} {return pb.WiFiWorkingMode_AP}),
		hvue.Computed("mode_sta", func(vm *hvue.VM) interface{} {return pb.WiFiWorkingMode_STA}),
		hvue.Computed("mode_failover", func(vm *hvue.VM) interface{} {return pb.WiFiWorkingMode_STA_FAILOVER_AP}),
		hvue.Method("reset",
			func(vm *hvue.VM) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_WIFI_SETTINGS_FROM_DEPLOYED)
			}),
		hvue.Method("deploy",
			func(vm *hvue.VM, wifiSettings *jsWiFiSettings) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_WIFI_SETTINGS, wifiSettings)
			}),
		hvue.Computed("deploying",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("deployingWifiSettings")
			}),

	)
}

const templateWiFi = `
<q-page padding>
<div class="row gutter-sm">
	<div class="col-lg-4">
	<q-card class="full-height">
		<q-card-title>
			WiFi settings
		</q-card-title>
	
		<q-card-actions>
			<q-btn :loading="deploying" color="primary" @click="deploy(settings)" label="deploy"></q-btn>
			<q-btn color="secondary" @click="reset" label="reset"></q-btn>
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

	<div class="col-lg-12">
	<q-card class="full-height">
		<q-card-title>
			<q-icon name="alarm" /><q-icon name="alarm" />
		</q-card-title>
		WiFiSettings {{ settings }} <br> 
		WiFiState {{ $store.state.wifiConnectionState }} 
	</q-card>
	</div>

</div>
</q-page>	

`