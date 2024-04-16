package bindings

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sag-enhanced/native-app/src/file"
	"github.com/sag-enhanced/native-app/src/options"
	"github.com/sag-enhanced/native-app/src/ui"
)

type Bindings struct {
	options *options.Options
	ui      ui.UII
	fm      *file.FileManager
}

func NewBindings(options *options.Options, ui ui.UII, fm *file.FileManager) *Bindings {
	return &Bindings{options, ui, fm}
}

// our own RPC engine ontop of the one that webview already provides
// because the webview one is blocking and we want to be able to call
// functions that take a while to complete (eg make a network request)

func (b *Bindings) BindHandler(method string, callId int, params string) error {
	if b.options.Verbose {
		fmt.Println("RPC call:", method, callId, params)
	}
	prefix := fmt.Sprintf("if(saged[%d]){", callId)
	suffix := fmt.Sprintf(";delete saged[%d]}", callId)

	methodName := strings.ToUpper(method[:1]) + method[1:]

	bindingType := reflect.TypeOf(b)
	binding, ok := bindingType.MethodByName(methodName)
	if !ok {
		return fmt.Errorf("method not found: %s (%s)", method, methodName)
	}

	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(params), &raw); err != nil {
		return err
	}
	if len(raw)+1 < binding.Type.NumIn() {
		return fmt.Errorf("wrong number of arguments (got %d, expected %d)", len(raw), binding.Type.NumIn())
	}

	args := []reflect.Value{reflect.ValueOf(b)}
	for i := range raw {
		arg := reflect.New(binding.Type.In(1 + i))
		if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
			return err
		}
		args = append(args, arg.Elem())
	}

	go func() {
		result, err := parseResults(binding.Func.Call(args))

		if err != nil {
			b.ui.Eval(prefix + fmt.Sprintf("saged[%d].b(new Error(%q))", callId, err.Error()) + suffix)
			return
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Failed to marshal result of RPC function", method, err)
			b.ui.Eval(prefix + fmt.Sprintf("saged[%d].b(new Error('result marshal failed'));", callId) + suffix)
			return
		}
		b.ui.Eval(prefix + fmt.Sprintf("saged[%d].a(%s)", callId, string(encoded)) + suffix)
	}()
	return nil
}

func parseResults(results []reflect.Value) (interface{}, error) {
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		if err, ok := results[0].Interface().(error); ok {
			return nil, err
		}
		return results[0].Interface(), nil
	}
	if len(results) == 2 {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}
		return results[0].Interface(), nil
	}
	return nil, errors.New("too many return values")
}
