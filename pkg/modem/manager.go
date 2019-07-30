package modem

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
	"go.uber.org/zap"
)

type Manager struct {
	SystemBus *dbus.Conn
	Logger    *zap.Logger
}

func (mgr *Manager) findModemsOnBus(conn *dbus.Conn, destination string, path dbus.ObjectPath) ([]dbus.ObjectPath, error) {
	managedObjectPaths := []dbus.ObjectPath{}

	// interface{} is hiding  -->  map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	managedObjectsResponse := make(map[dbus.ObjectPath]interface{})

	mm := conn.Object(destination, path)

	err := mm.Call(MethodGetManagedObjects, 0).Store(&managedObjectsResponse)
	if err != nil {
		mgr.Logger.Error("dbus_get-managed-objects.failure", zap.Error(err))
		return nil, err
	}

	for path := range managedObjectsResponse {
		managedObjectPaths = append(managedObjectPaths, path)
	}
	return managedObjectPaths, nil
}

func (mgr *Manager) queryBusForProperties(conn *dbus.Conn, objPath dbus.ObjectPath, propertyBasePath string, dstPtr interface{}) error {
	obj := conn.Object(ModemManagerService, objPath)
	dst := reflect.ValueOf(dstPtr)
	dstType := reflect.TypeOf(dstPtr)

	/*
		This is a half-baked implementation, but...
		The idea here is to use struct tags to dictate which properties should be pulled from dbus
	*/
	for i := 0; i < dstType.Elem().NumField(); i++ {
		propertyName := dst.Type().Elem().Field(i).Tag.Get("propertyPath")
		if propertyName == "" {
			continue
		}

		// TODO Isolate actual fetch
		variant, err := obj.GetProperty(fmt.Sprintf("%s.%s", propertyBasePath, propertyName))
		if err != nil {
			// TODO: log as error
			mgr.Logger.Error("get_property.query_failure", zap.Error(err))
			continue
		}

		variantValue := variant.Value()
		field := dst.Elem().Field(i)
		switch field.Kind() {
		case reflect.Bool:
			field.SetBool(variantValue.(bool))
		case reflect.Int64, reflect.Int: // FIXME
			field.SetInt(variantValue.(int64))
		case reflect.String:
			// Maybe just replace everything with this...
			// Yeah, this actually works :-/
			switch field.Type() {
			case reflect.TypeOf(dbus.ObjectPath("")):
				field.Set(reflect.ValueOf(variantValue.(dbus.ObjectPath)))
			case reflect.TypeOf(string("")):
				field.SetString(variantValue.(string))
			}
		case reflect.Slice:
			switch field.Type() {
			case reflect.TypeOf([]string{}):
				values := variantValue.([]string)
				valuePtr := reflect.ValueOf(&[]string{})
				value := valuePtr.Elem()
				for _, v := range values {
					value.Set(reflect.Append(value, reflect.ValueOf(v)))
				}
				field.Set(value)
			case reflect.TypeOf([]dbus.ObjectPath{}):
				values := variantValue.([]dbus.ObjectPath)
				valuePtr := reflect.ValueOf(&[]dbus.ObjectPath{})
				value := valuePtr.Elem()
				for _, v := range values {
					value.Set(reflect.Append(value, reflect.ValueOf(v)))
				}
				field.Set(value)
			}
			break
		case reflect.ValueOf(ModemSignalQuality{}).Kind(): // Not yet grasping how to remove weird special cases...
			variantAsInterfaceSlice := variant.Value().([]interface{})
			sigQuality := ModemSignalQuality{
				Percent: variantAsInterfaceSlice[0].(uint32),
				Recent:  variantAsInterfaceSlice[1].(bool),
			}
			field.Set(reflect.ValueOf(sigQuality))
		case reflect.Map:
			continue
		case reflect.Struct:
			continue
		}

	}

	dstPtr = &dst
	return nil
}

/*
conn: dbus connection
objPath: modem path -- e.g. /org/freedesktop/ModemManager1/Modem/94
*/
func (mgr *Manager) getModemSignal(conn *dbus.Conn, modemPath dbus.ObjectPath) (Signal, error) {
	_ = conn.Object(ModemManagerService, modemPath)

	/*

		Left as a stub, since this information isn't normally made available without first calling `Setup()` on the Signal object

	*/

	return Signal{}, errors.New("unavailable")
}

func extractHexEncodedUint(encodedVal string) (string, error) {
	val, err := strconv.ParseUint(encodedVal, 16, 32)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(val)), nil
}

func (mgr *Manager) GetModem(path dbus.ObjectPath) (Modem, error) {
	m := &Modem{}
	err := mgr.queryBusForProperties(mgr.SystemBus, path, objectPathModem, m)
	if err != nil {
		return Modem{}, err
	}
	return *m, nil
}

func (mgr *Manager) GetBearer(path dbus.ObjectPath) (Bearer, error) {
	b := &Bearer{}
	err := mgr.queryBusForProperties(mgr.SystemBus, path, objectPathBearer, b)
	if err != nil {
		return Bearer{}, err
	}
	return *b, nil
}

func (mgr *Manager) GetSim(path dbus.ObjectPath) (Sim, error) {
	b := &Sim{}
	err := mgr.queryBusForProperties(mgr.SystemBus, path, objectPathSim, b)
	if err != nil {
		return Sim{}, err
	}
	return *b, nil
}

func (mgr *Manager) CallGetModemLocation(path dbus.ObjectPath) (*Location, error) {
	// e.g. {1: '310,260,417B,1411502,0'}
	loc := &Location{}

	bus := mgr.SystemBus.Object(ModemManagerService, path)

	resp := make(map[uint32]dbus.Variant)
	err := bus.CallWithContext(context.TODO(), callModemLocationGetLocation, 0).Store(resp)
	if err != nil {
		return nil, err
	}

	if len(resp) < 1 {
		return nil, errNoLocation
	}

	for _, v := range resp {
		respStr := v.Value().(string)
		locParts := strings.Split(respStr, ",")

		loc.MCC = locParts[0]
		loc.MNC = locParts[1]

		lac, err := extractHexEncodedUint(locParts[2])
		if err != nil {
			mgr.Logger.Debug("failed to decode lac hex string", zap.Error(err))
		}
		loc.LAC = lac

		cellId, err := extractHexEncodedUint(locParts[3])
		if err != nil {
			mgr.Logger.Debug("failed to decode cell id hex string", zap.Error(err))
		}
		loc.CID = cellId

		tac, err := extractHexEncodedUint(locParts[4])
		if err != nil {
			mgr.Logger.Debug("failed to decode cell tac hex string", zap.Error(err))
		}
		loc.TAC = tac

		break
	}
	return loc, nil
}

func (mgr *Manager) GetManagedModems() ([]dbus.ObjectPath, error) {
	modemPaths, err := mgr.findModemsOnBus(mgr.SystemBus, ModemManagerService, PathModemManager)
	if err != nil {
		return nil, err
	}
	return modemPaths, nil
}
