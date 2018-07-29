package main

import (
	"github.com/mame82/P4wnP1_go/service"
	"log"
	"fmt"
)

var iw_scan_out = `BSS 4e:66:41:a0:5b:35(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2437
	beacon interval: 100 TUs
	capability: ESS Privacy SpectrumMgmt ShortSlotTime RadioMeasure (0x1511)
	signal: -43.01 dBm
	last seen: 13740 ms ago
	SSID: Android AP8905
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0
	DS Parameter set: channel 6
	TIM: DTIM Count 0 DTIM Period 2 Bitmap Control 0x0 Bitmap[0] 0x0
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	Power constraint: 0 dB
	TPC report: TX power: 17 dBm
	ERP: <no flags>
	Extended supported rates: 6.0 9.0 12.0 48.0
	RSN:	 * Version: 1
		 * Group cipher: CCMP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 16-PTKSA-RC 1-GTKSA-RC (0x000c)
	HT capabilities:
		Capabilities: 0x1ad
			RX LDPC
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			No DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 4 usec (0x05)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 6
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 0
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Extended capabilities: Extended Channel Switching, BSS Transition, 6
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS 5c:dc:96:b4:59:af(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2437
	beacon interval: 100 TUs
	capability: ESS Privacy ShortSlotTime RadioMeasure (0x1411)
	signal: -38.00 dBm
	last seen: 10 ms ago
	SSID: WLAN-579086
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0
	DS Parameter set: channel 6
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 22 dBm
	ERP: <no flags>
	ERP D4.0: <no flags>
	RSN:	 * Version: 1
		 * Group cipher: CCMP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 16-PTKSA-RC 1-GTKSA-RC (0x000c)
	Extended supported rates: 6.0 9.0 12.0 48.0
	HT capabilities:
		Capabilities: 0x19fe
			HT20/HT40
			SM Power Save disabled
			RX Greenfield
			RX HT20 SGI
			RX HT40 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 7935 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 16 usec (0x07)
		HT RX MCS rate indexes supported: 0-15, 32
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 6
		 * secondary channel offset: below
		 * STA channel width: any
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 0
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Overlapping BSS scan params:
		 * passive dwell: 20 TUs
		 * active dwell: 10 TUs
		 * channel width trigger scan interval: 300 s
		 * scan passive total per channel: 200 TUs
		 * scan active total per channel: 20 TUs
		 * BSS width channel transition delay factor: 5
		 * OBSS Scan Activity Threshold: 0.25 %
	Extended capabilities: HT Information Exchange Supported, Extended Channel Switching, 6
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: 7470bfaa-0621-242d-7915-13686f9ded23
		 * Manufacturer: ARCADYAN
		 * Model: DT724
		 * Model Number: 1.0
		 * Serial Number: 888
		 * Primary Device Type: 6-0050f204-1
		 * Device name: DT724
		 * Config methods:
		 * RF Bands: 0x3
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20
	WMM:	 * Parameter version 1
		 * u-APSD
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS 4e:66:41:a0:5b:35(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2462
	beacon interval: 100 TUs
	capability: ESS SpectrumMgmt ShortSlotTime RadioMeasure (0x1501)
	signal: -35.00 dBm
	last seen: 10 ms ago
	SSID: AndroidAP8905
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0
	DS Parameter set: channel 11
	TIM: DTIM Count 0 DTIM Period 2 Bitmap Control 0x0 Bitmap[0] 0x0
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	Power constraint: 0 dB
	TPC report: TX power: 17 dBm
	ERP: <no flags>
	Extended supported rates: 6.0 9.0 12.0 48.0
	HT capabilities:
		Capabilities: 0x1ad
			RX LDPC
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			No DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 4 usec (0x05)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 11
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 0
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Extended capabilities: Extended Channel Switching, BSS Transition, 6
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS 88:e3:ab:9d:38:f5(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2412
	beacon interval: 100 TUs
	capability: ESS Privacy ShortSlotTime RadioMeasure (0x1411)
	signal: -84.00 dBm
	last seen: 0 ms ago
	SSID: WLAN-WGF3FG
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0 
	DS Parameter set: channel 1
	ERP: Barker_Preamble_Mode
	ERP D4.0: Barker_Preamble_Mode
	RSN:	 * Version: 1
		 * Group cipher: CCMP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 16-PTKSA-RC 1-GTKSA-RC (0x000c)
	Extended supported rates: 6.0 9.0 12.0 48.0 
	HT capabilities:
		Capabilities: 0x19ac
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 7935 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 1
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 0
		 * HT protection: no
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Overlapping BSS scan params:
		 * passive dwell: 20 TUs
		 * active dwell: 10 TUs
		 * channel width trigger scan interval: 300 s
		 * scan passive total per channel: 200 TUs
		 * scan active total per channel: 20 TUs
		 * BSS width channel transition delay factor: 5
		 * OBSS Scan Activity Threshold: 0.25 %
	Extended capabilities: HT Information Exchange Supported, Extended Channel Switching, 6
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: 13082394-62d2-8063-e9de-c42e3a2d1071
		 * Manufacturer: Broadcom
		 * Model: Broadcom
		 * Model Number: 123456
		 * Serial Number: 1234
		 * Primary Device Type: 6-0050f204-1
		 * Device name: BroadcomAP
		 * Config methods:
		 * RF Bands: 0x3
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20
	WMM:	 * Parameter version 1
		 * u-APSD
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS b8:bc:1b:63:33:d9(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2412
	beacon interval: 100 TUs
	capability: ESS Privacy ShortSlotTime RadioMeasure (0x1411)
	signal: -81.00 dBm
	last seen: 0 ms ago
	SSID: schmidtderhit
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0 
	DS Parameter set: channel 1
	ERP: Barker_Preamble_Mode
	ERP D4.0: Barker_Preamble_Mode
	RSN:	 * Version: 1
		 * Group cipher: CCMP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 16-PTKSA-RC 1-GTKSA-RC (0x000c)
	Extended supported rates: 6.0 9.0 12.0 48.0 
	HT capabilities:
		Capabilities: 0x19ac
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 7935 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 1
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 0
		 * HT protection: no
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Overlapping BSS scan params:
		 * passive dwell: 20 TUs
		 * active dwell: 10 TUs
		 * channel width trigger scan interval: 300 s
		 * scan passive total per channel: 200 TUs
		 * scan active total per channel: 20 TUs
		 * BSS width channel transition delay factor: 5
		 * OBSS Scan Activity Threshold: 0.25 %
	Extended capabilities: HT Information Exchange Supported, Extended Channel Switching, 6
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: 8e998705-3ed2-399e-e43d-0d4c0a412fec
		 * Manufacturer: Broadcom
		 * Model: Broadcom
		 * Model Number: 123456
		 * Serial Number: 1234
		 * Primary Device Type: 6-0050f204-1
		 * Device name: BroadcomAP
		 * Config methods:
		 * RF Bands: 0x3
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20
	WMM:	 * Parameter version 1
		 * u-APSD
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS 88:e3:ab:9d:38:f6(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2412
	beacon interval: 100 TUs
	capability: ESS ShortSlotTime RadioMeasure (0x1401)
	signal: -84.00 dBm
	last seen: 0 ms ago
	SSID: Telekom_FON
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0 
	DS Parameter set: channel 1
	ERP: Barker_Preamble_Mode
	ERP D4.0: Barker_Preamble_Mode
	Extended supported rates: 6.0 9.0 12.0 48.0 
	HT capabilities:
		Capabilities: 0x19ac
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 7935 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 1
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 0
		 * HT protection: no
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Overlapping BSS scan params:
		 * passive dwell: 20 TUs
		 * active dwell: 10 TUs
		 * channel width trigger scan interval: 300 s
		 * scan passive total per channel: 200 TUs
		 * scan active total per channel: 20 TUs
		 * BSS width channel transition delay factor: 5
		 * OBSS Scan Activity Threshold: 0.25 %
	Extended capabilities: HT Information Exchange Supported, Extended Channel Switching, 6
	WMM:	 * Parameter version 1
		 * u-APSD
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS b8:bc:1b:63:33:da(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2412
	beacon interval: 100 TUs
	capability: ESS ShortSlotTime RadioMeasure (0x1401)
	signal: -83.00 dBm
	last seen: 0 ms ago
	SSID: Telekom_FON
	Supported rates: 1.0* 2.0* 5.5* 11.0* 18.0 24.0 36.0 54.0 
	DS Parameter set: channel 1
	ERP: Barker_Preamble_Mode
	ERP D4.0: Barker_Preamble_Mode
	Extended supported rates: 6.0 9.0 12.0 48.0 
	HT capabilities:
		Capabilities: 0x19ac
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 7935 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 1
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 0
		 * HT protection: no
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Overlapping BSS scan params:
		 * passive dwell: 20 TUs
		 * active dwell: 10 TUs
		 * channel width trigger scan interval: 300 s
		 * scan passive total per channel: 200 TUs
		 * scan active total per channel: 20 TUs
		 * BSS width channel transition delay factor: 5
		 * OBSS Scan Activity Threshold: 0.25 %
	Extended capabilities: HT Information Exchange Supported, Extended Channel Switching, 6
	WMM:	 * Parameter version 1
		 * u-APSD
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
BSS 24:65:11:89:1c:28(on wlan0) -- associated
	TSF: 0 usec (0d, 00:00:00)
	freq: 2462
	beacon interval: 100 TUs
	capability: ESS Privacy ShortPreamble ShortSlotTime (0x0431)
	signal: -40.00 dBm
	last seen: 0 ms ago
	SSID: spycki3
	Supported rates: 1.0* 2.0* 5.5* 11.0* 6.0* 9.0 12.0* 18.0 
	DS Parameter set: channel 11
	TIM: DTIM Count 0 DTIM Period 1 Bitmap Control 0x0 Bitmap[0] 0x1
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	ERP: <no flags>
	RSN:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 1-PTKSA-RC 1-GTKSA-RC (0x0000)
	Extended supported rates: 24.0* 36.0 48.0 54.0 
	HT capabilities:
		Capabilities: 0x11ce
			HT20/HT40
			SM Power Save disabled
			RX HT40 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT RX MCS rate indexes supported: 0-15
		HT TX MCS rate indexes are undefined
	HT operation:
		 * primary channel: 11
		 * secondary channel offset: below
		 * STA channel width: any
		 * RIFS: 1
		 * HT protection: 20 MHz
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	WPA:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: TKIP
		 * Authentication suites: PSK
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * UUID: 0cb4b30f-b5d6-7e18-4ec8-246511891c10
		 * RF Bands: 0x3
BSS c0:25:06:eb:8e:7d(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2437
	beacon interval: 100 TUs
	capability: ESS Privacy ShortPreamble ShortSlotTime (0x0431)
	signal: -84.00 dBm
	last seen: 0 ms ago
	SSID: FRITZ!Box Fon WLAN 7390
	Supported rates: 1.0* 2.0* 5.5* 11.0* 6.0* 9.0 12.0* 18.0 
	DS Parameter set: channel 6
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	ERP: <no flags>
	Extended supported rates: 24.0* 36.0 48.0 54.0 
	HT capabilities:
		Capabilities: 0x18c
			HT20
			SM Power Save disabled
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			No DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT TX/RX MCS rate indexes supported: 0-15
	HT operation:
		 * primary channel: 6
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 0
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Extended capabilities: 6
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
	RSN:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 1-PTKSA-RC 1-GTKSA-RC (0x0000)
	WPA:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: TKIP
		 * Authentication suites: PSK
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: 76fc7d2e-2e0b-5c3b-7135-c02506eb8e7c
		 * Manufacturer: AVM
		 * Model: FBox
		 * Model Number: 0000
		 * Serial Number: 0000
		 * Primary Device Type: 6-0050f204-1
		 * Device name: FBox
		 * Config methods: Display, PBC, Keypad
		 * RF Bands: 0x3
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20
BSS c2:25:06:eb:8e:7d(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2437
	beacon interval: 100 TUs
	capability: ESS Privacy ShortPreamble ShortSlotTime (0x0431)
	signal: -82.00 dBm
	last seen: 0 ms ago
	SSID: SDK Gast
	Supported rates: 1.0* 2.0* 5.5* 11.0* 6.0* 9.0 12.0* 18.0 
	DS Parameter set: channel 6
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	ERP: <no flags>
	Extended supported rates: 24.0* 36.0 48.0 54.0 
	HT capabilities:
		Capabilities: 0x18c
			HT20
			SM Power Save disabled
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			No DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT TX/RX MCS rate indexes supported: 0-15
	HT operation:
		 * primary channel: 6
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 0
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Extended capabilities: 6
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
	RSN:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 1-PTKSA-RC 1-GTKSA-RC (0x0000)
	WPA:	 * Version: 1
		 * Group cipher: TKIP
		 * Pairwise ciphers: TKIP
		 * Authentication suites: PSK
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: dfb77a7a-25b2-93f6-7ba8-01e7cd0bdb51
		 * Manufacturer: AVM
		 * Model: FBox
		 * Model Number: 0000
		 * Serial Number: 0000
		 * Primary Device Type: 6-0050f204-1
		 * Device name: FBox
		 * Config methods: Display, PBC, Keypad
		 * RF Bands: 0x3
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20
BSS e0:28:6d:dd:b1:fc(on wlan0)
	TSF: 0 usec (0d, 00:00:00)
	freq: 2462
	beacon interval: 100 TUs
	capability: ESS Privacy ShortPreamble (0x0031)
	signal: -75.00 dBm
	last seen: 0 ms ago
	SSID: FRITZ!Box 7362 SL
	Supported rates: 1.0* 2.0* 5.5* 11.0* 6.0* 9.0 12.0* 18.0 
	DS Parameter set: channel 11
	Country: DE	Environment: Indoor/Outdoor
		Channels [1 - 13] @ 20 dBm
	ERP: <no flags>
	Extended supported rates: 24.0* 36.0 48.0 54.0 
	HT capabilities:
		Capabilities: 0x1ad
			RX LDPC
			HT20
			SM Power Save disabled
			RX HT20 SGI
			TX STBC
			RX STBC 1-stream
			Max AMSDU length: 3839 bytes
			No DSSS/CCK HT40
		Maximum RX AMPDU length 65535 bytes (exponent: 0x003)
		Minimum RX AMPDU time spacing: 8 usec (0x06)
		HT TX/RX MCS rate indexes supported: 0-23
	HT operation:
		 * primary channel: 11
		 * secondary channel offset: no secondary
		 * STA channel width: 20 MHz
		 * RIFS: 1
		 * HT protection: no
		 * non-GF present: 1
		 * OBSS non-GF present: 0
		 * dual beacon: 0
		 * dual CTS protection: 0
		 * STBC beacon: 0
		 * L-SIG TXOP Prot: 0
		 * PCO active: 0
		 * PCO phase: 0
	Extended capabilities: 6
	WMM:	 * Parameter version 1
		 * BE: CW 15-1023, AIFSN 3
		 * BK: CW 15-1023, AIFSN 7
		 * VI: CW 7-15, AIFSN 2, TXOP 3008 usec
		 * VO: CW 3-7, AIFSN 2, TXOP 1504 usec
	RSN:	 * Version: 1
		 * Group cipher: CCMP
		 * Pairwise ciphers: CCMP
		 * Authentication suites: PSK
		 * Capabilities: 1-PTKSA-RC 1-GTKSA-RC (0x0000)
	WPS:	 * Version: 1.0
		 * Wi-Fi Protected Setup State: 2 (Configured)
		 * Response Type: 3 (AP)
		 * UUID: 80ffd276-a392-0a84-6421-e0286dddb1fc
		 * Manufacturer: AVM
		 * Model: FBox
		 * Model Number: 0000
		 * Serial Number: 0000
		 * Primary Device Type: 6-0050f204-1
		 * Device name: FBox
		 * Config methods: Display, PBC, Keypad
		 * RF Bands: 0x1
		 * Unknown TLV (0x1049, 6 bytes): 00 37 2a 00 01 20`

 func main() {
	//res, err := service.ParseIwScan(iw_scan_out)
	res, err := service.WifiScan("wlan0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed scan result:\n%v\n", res)
	err = service.WifiCreateWpaSupplicantConfigFile("spycki1 2 3", "a b c c b b", "/tmp/wpa_supplicant.conf")
	if err != nil { log.Fatal(err)}

 }