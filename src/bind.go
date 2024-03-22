package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// our own RPC engine ontop of the one that webview already provides
// because the webview one is blocking and we want to be able to call
// functions that take a while to complete (eg make a network request)

func (app *App) bindHandler(method string, callId int, params string) error {
	handler, ok := app.bindings[method]
	if app.options.Verbose {
		fmt.Println("RPC call:", method, callId, params)
	}
	prefix := fmt.Sprintf("if(saged[%d]){", callId)
	suffix := fmt.Sprintf(";delete saged[%d]}", callId)

	if !ok {
		return errors.New("invalid method: " + method)
	}
	go func() {
		result, err := handler(params)
		if err != nil {
			app.ui.eval(prefix + fmt.Sprintf("saged[%d].b(new Error(%q))", callId, err.Error()) + suffix)
			return
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Failed to marshal result of RPC function", method, err)
			app.ui.eval(prefix + fmt.Sprintf("saged[%d].b(new Error('result marshal failed'));", callId) + suffix)
			return
		}
		app.ui.eval(prefix + fmt.Sprintf("saged[%d].a(%s)", callId, string(encoded)) + suffix)
	}()
	return nil
}

func (app *App) bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)

	if v.Kind() != reflect.Func {
		return errors.New("not a function")
	}

	if n := v.Type().NumOut(); n > 2 {
		return errors.New("too many return values")
	}

	binding := func(req string) (interface{}, error) {
		raw := []json.RawMessage{}
		if err := json.Unmarshal([]byte(req), &raw); err != nil {
			return nil, err
		}
		isVariadic := v.Type().IsVariadic()
		numIn := v.Type().NumIn()
		if (isVariadic && len(raw) < numIn-1) || (!isVariadic && len(raw) < numIn) {
			return nil, errors.New("wrong number of arguments")
		}
		args := []reflect.Value{}
		for i := range raw {
			if !isVariadic && i >= numIn {
				break
			}
			var arg reflect.Value
			if isVariadic && i >= numIn-1 {
				arg = reflect.New(v.Type().In(numIn - 1).Elem())
			} else {
				arg = reflect.New(v.Type().In(i))
			}
			if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
				return nil, err
			}
			args = append(args, arg.Elem())
		}
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		results := v.Call(args)

		switch len(results) {
		case 0:
			return nil, nil
		case 1:
			if results[0].Type().Implements(errorType) {
				if results[0].IsNil() {
					return nil, nil
				}
				return nil, results[0].Interface().(error)
			}
			return results[0].Interface(), nil
		case 2:
			if results[1].Type().Implements(errorType) {
				if results[1].IsNil() {
					return results[0].Interface(), nil
				}
				return nil, results[1].Interface().(error)
			}
			return nil, errors.New("second return value must be an error")
		default:
			return nil, errors.New("too many return values")
		}
	}

	app.bindings[name] = binding
	return nil
}
