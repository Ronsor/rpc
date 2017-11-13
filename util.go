package rpc
import "reflect"
import "errors"

func inCall(fn interface{}, params ... interface{}) (result []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = []interface{}{r}
			//err = r.(string)
		}
	} ()
	f := reflect.ValueOf(fn)
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
//		println("p")
//		println(k)
		xp := param
		vt := reflect.ValueOf(xp)
//		println(f.Type().In(k).String(), vt.Type().String())
		if vt.Type().String()[0] != 'u' {
		switch f.Type().In(k).String() {
			case "int": xp = int(vt.Int()); break
			case "int64": xp = int64(vt.Int()); break
			case "int32": xp = int32(vt.Int()); break
		}
		} else {
		switch f.Type().In(k).String() {
			case "int": xp = int(vt.Uint()); break
			case "int64": xp = int64(vt.Uint()); break
			case "int32": xp = int32(vt.Uint()); break
		}
		}
		in[k] = reflect.ValueOf(xp)
		
	}
	result2 := f.Call(in)
	result = make([]interface{}, len(result2))
	for k, v := range result2 {
		result[k] = v.Interface()
	}
	return
}

