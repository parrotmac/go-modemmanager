package modem

import (
	"fmt"

	"github.com/godbus/dbus"
	"github.com/pkg/errors"
)

var BadTypeSignatureErr = errors.New("bad type signature")
var BadCastErr = errors.New("bad type assertion")

const ModemManagerService = "org.freedesktop.ModemManager1"

// Expected to be in the bus
// Most other objects such as Modems or SIMs should be queried
const PathModemManager = "/org/freedesktop/ModemManager1"
const MethodGetManagedObjects = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"

// org.freedesktop.ModemManager1.Modem
const objectPathModem = "org.freedesktop.ModemManager1.Modem"

type Modem struct {
	Bearers      []dbus.ObjectPath `propertyPath:"Bearers" json:"bearer_paths"`
	Sim          dbus.ObjectPath   `propertyPath:"Sim" json:"sim_path"`
	Model        string            `propertyPath:"Model" json:"model"`
	Manufacturer string            `propertyPath:"Manufacturer" json:"manufacturer"`
	Device       string            `propertyPath:"Device" json:"device"`
	// Ports                []struct{string, int} `propertyPath:"Ports" json:"ports"` // TODO
	PrimaryPort         string             `propertyPath:"PrimaryPort" json:"primary_port"`
	OwnNumbers          []string           `propertyPath:"OwnNumbers" json:"own_numbers"`
	SoftwareRevision    string             `propertyPath:"Revision" json:"revision"`
	HardwareRevision    string             `propertyPath:"HardwareRevision" json:"hardware_revision"`
	EquipmentIdentifier string             `propertyPath:"EquipmentIdentifier" json:"equipment_identifier"`
	Drivers             []string           `propertyPath:"Drivers" json:"drivers"`
	ModemSignalQuality  ModemSignalQuality `propertyPath:"SignalQuality" json:"signal_quality"`
}

// FIXME Currently special-cased
// Is there enough data returned by the D-Bus API to generalize this?
type ModemSignalQuality struct {
	Percent uint32 `json:"percent"` /* 0 - 100 signal stength */
	Recent  bool   `json:"recent"`  /* whether reading is recent */
}

// Signal information requires activation
const objectPathSignal = "org.freedesktop.ModemManager1.Modem.Signal"

type Signal struct {
	Cdma map[string]dbus.Variant // rssi, ecio
	Evdo map[string]dbus.Variant // rssi, ecio, sinr, io
	Gsm  map[string]dbus.Variant // rssi
	Umts map[string]dbus.Variant // rssi, rsrp, ecio
	Lte  map[string]dbus.Variant // rssi, rsrp, rsrq, snr
}

const objectPathModem3gpp = "org.freedesktop.ModemManager1.Modem.Modem3gpp"

type Modem3gpp struct {
	Imei string `propertyPath:"Imei" json:"imei"`
}

const objectPathSim = "org.freedesktop.ModemManager1.Sim"

type Sim struct {
	Imsi               string `propertyPath:"Imsi" json:"imsi"`
	OperatorIdentifier string `propertyPath:"OperatorIdentifier" json:"operator_identifier"`
	OperatorName       string `propertyPath:"OperatorName" json:"operator_name"`
	SimIdentifier      string `propertyPath:"SimIdentifier" json:"sim_identifier"`
}

const objectPathBearer = "org.freedesktop.ModemManager1.Bearer"

type Bearer struct {
	Connected bool   `propertyPath:"Connected" json:"connected"`
	Suspended bool   `propertyPath:"Suspended" json:"suspended"`
	Interface string `propertyPath:"Interface" json:"interface"`
}

const objectPathModemLocation = "org.freedesktop.ModemManager1.Modem.Location"

var callModemLocationGetLocation = fmt.Sprintf("%s.%s", objectPathModemLocation, "GetLocation")
var errNoLocation = errors.New("no location found")

type Location struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
	LAC string `json:"lac"`
	CID string `json:"cid"`
	TAC string `json:"tac"`
}
