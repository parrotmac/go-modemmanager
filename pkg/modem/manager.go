package modem

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/godbus/dbus"
	"go.uber.org/zap"

	"github.com/parrotmac/rusted/pkg/device/modem"
)

type Manager struct {
	SystemBus *dbus.Conn
	Logger    *zap.Logger

	modemObjectPaths []dbus.ObjectPath
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

// func getPropertyAsString(object dbus.BusObject, path string) (*string, error) {
// 	v, err := object.GetProperty(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if v.Signature() != dbus.SignatureOf(reflect.String) {
// 		return nil, BadTypeSignatureErr
// 	}
// 	str, ok := v.Value().(string)
// 	if !ok {
// 		return nil, BadCastErr
// 	}
// 	return &str, err
// }
//
// func getPropertyAsBool(object dbus.BusObject, path string) (*bool, error) {
// 	v, err := object.GetProperty(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if v.Signature() != dbus.SignatureOf(reflect.Bool) {
// 		return nil, BadTypeSignatureErr
// 	}
// 	b, ok := v.Value().(bool)
// 	if !ok {
// 		return nil, BadCastErr
// 	}
// 	return &b, nil
// }
//
// func getPropertyAsInt(object dbus.BusObject, path string) (*int, error) {
// 	v, err := object.GetProperty(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if v.Signature() != dbus.SignatureOf(types.Int) || v.Signature() != dbus.SignatureOf(types.Int64) {
// 		return nil, errors.New("bad")
// 	}
// 	i, ok := v.Value().(int)
// 	if !ok {
// 		return nil, BadCastErr
// 	}
// 	return &i, err
// }

// func queryBusForBearer(conn *dbus.Conn, bearerPath dbus.ObjectPath) (Bearer, error) {
// 	bearerObj := conn.Object(ModemManagerService, bearerPath)
// 	bearer := Bearer{
// 		Connected: false,
// 		Suspended: false,
// 		Interface: "",
// 	}
//
// 	connected, err := getPropertyAsBool(bearerObj, fmt.Sprintf("%s.%s", objectPathBearer, "Connected"))
// 	if err == nil {
// 		bearer.Connected = *connected
// 	}
//
// 	suspended, err := getPropertyAsBool(bearerObj, fmt.Sprintf("%s.%s", objectPathBearer, "Suspended"))
// 	if err == nil {
// 		bearer.Suspended = *suspended
// 	}
//
// 	iface, err := getPropertyAsString(bearerObj, fmt.Sprintf("%s.%s", objectPathBearer, "Properties"))
// 	if err == nil {
// 		bearer.Interface = *iface
// 	}
//
// 	return bearer, nil
// }

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

func (mgr *Manager) GetModemList() ([]Modem, error) {
	modems := []Modem{}
	for _, modemPath := range mgr.modemObjectPaths {
		modem := &Modem{}
		err := mgr.queryBusForProperties(mgr.SystemBus, modemPath, objectPathModem, modem)
		if err != nil {
			return nil, err
		}
		modems = append(modems, *modem)
	}
	return modems, nil

	// modemManager := conn.Object("org.freedesktop.ModemManager1", "/org/freedesktop/ModemManager1/Modem/131")
	//
	// variant, err := modemManager.GetProperty("org.freedesktop.ModemManager1.Modem.Model")
	// if err != nil {
	// 	mgr.Logger.Fatal(err)
	// }
	// mgr.Logger.Println("modem model:", variant.Value().(string))
	//
	// simVariant, err := modemManager.GetProperty("org.freedesktop.ModemManager1.Modem.Sim")
	// if err != nil {
	// 	mgr.Logger.Fatalln(err)
	// }
	//
	// simPath := simVariant.Value().(dbus.ObjectPath)
	// mgr.Logger.Println(simPath)
	//
	// simInfo := conn.Object(ModemManagerService, simPath)
	// simOperator, err := simInfo.GetProperty("org.freedesktop.ModemManager1.Sim.OperatorName")
	// if err != nil {
	// 	mgr.Logger.Fatalln(err)
	// }
	// mgr.Logger.Println(simOperator.Value().(string))
	//
	// modemBearerPath := dbus.ObjectPath("/org/freedesktop/ModemManager1/Bearer/0")
	// bearer, err := queryBusForBearer(conn, modemBearerPath)
	// if err != nil {
	// 	mgr.Logger.Fatalln(err)
	// }
	// mgr.Logger.Println(bearer)
	//
	// modem := &Modem{}
	// err = mgr.queryBusForProperties(conn, modemPaths[0], "org.freedesktop.ModemManager1.Modem", modem)
	// if err != nil {
	// 	mgr.Logger.Fatalln(err)
	// }
	// mgr.Logger.Printf("%+v", modem)
	//
	//
	// return nil, nil
}

func (mgr *Manager) GetBearer(path dbus.ObjectPath) (Bearer, error) {
	b := &Bearer{}
	err := mgr.queryBusForProperties(mgr.SystemBus, path, modem.BearerPath, b)
	if err != nil {
		return Bearer{}, err
	}
	return *b, nil
}

func (mgr *Manager) GetSim(path dbus.ObjectPath) (Sim, error) {
	b := &Sim{}
	err := mgr.queryBusForProperties(mgr.SystemBus, path, modem.SimPath, b)
	if err != nil {
		return Sim{}, err
	}
	return *b, nil
}

func (mgr *Manager) Scan() error {
	modemPaths, err := mgr.findModemsOnBus(mgr.SystemBus, ModemManagerService, PathModemManager)
	if err != nil {
		return err
	}

	mgr.Logger.Debug("managed_objects", zap.Any("object_listing", modemPaths))

	mgr.modemObjectPaths = modemPaths
	return nil
}
