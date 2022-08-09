package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"github.com/adamnite/go-adamnite/log15"
)

var (
	contextType      = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	subscriptionType = reflect.TypeOf(Subscription{})
	stringType       = reflect.TypeOf("")
)

type adamniteServiceRegistry struct {
	mu       sync.Mutex
	services map[string]adamniteService
}

type adamniteService struct {
	name          string
	callbacks     map[string]*rpcCallback
	subscriptions map[string]*rpcCallback
}

type rpcCallback struct {
	function    reflect.Value
	receiver    reflect.Value
	argTypes    []reflect.Type
	isSubscribe bool
	errPos      int
	hasCtx      bool
}

func (asr *adamniteServiceRegistry) registerName(name string, receiver interface{}) error {
	receiverVal := reflect.ValueOf(receiver)
	if name == "" {
		return fmt.Errorf("no service name for type %s", receiverVal.Type().String())
	}

	callbacks := suitableCallbacks(receiverVal)
	if len(callbacks) == 0 {
		return fmt.Errorf("service %T doesn't have any suitable methods/subscriptions to expose", receiver)
	}

	asr.mu.Lock()
	defer asr.mu.Unlock()
	if asr.services == nil {
		asr.services = make(map[string]adamniteService)
	}
	svc, ok := asr.services[name]
	if !ok {
		svc = adamniteService{
			name:          name,
			callbacks:     make(map[string]*rpcCallback),
			subscriptions: make(map[string]*rpcCallback),
		}
		asr.services[name] = svc
	}
	for name, cb := range callbacks {
		if cb.isSubscribe {
			svc.subscriptions[name] = cb
		} else {
			svc.callbacks[name] = cb
		}
	}
	return nil
}

// callback returns the callback corresponding to the given RPC method name.
func (r *adamniteServiceRegistry) callback(method string) *rpcCallback {
	elem := strings.SplitN(method, serviceMethodSeparator, 2)
	if len(elem) != 2 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.services[elem[0]].callbacks[elem[1]]
}

func suitableCallbacks(receiver reflect.Value) map[string]*rpcCallback {
	typ := receiver.Type()
	callbacks := make(map[string]*rpcCallback)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if method.PkgPath != "" {
			continue // method not exported
		}
		cb := newCallback(receiver, method.Func)
		if cb == nil {
			continue // function invalid
		}
		name := formatName(method.Name)
		callbacks[name] = cb
	}
	return callbacks
}

func newCallback(receiver, fn reflect.Value) *rpcCallback {
	fntype := fn.Type()
	c := &rpcCallback{function: fn, receiver: receiver, errPos: -1, isSubscribe: isPubSub(fntype)}
	// Determine parameter types. They must all be exported or builtin types.
	c.makeArgTypes()

	// Verify return types. The function must return at most one error
	// and/or one other non-error value.
	outs := make([]reflect.Type, fntype.NumOut())
	for i := 0; i < fntype.NumOut(); i++ {
		outs[i] = fntype.Out(i)
	}
	if len(outs) > 2 {
		return nil
	}
	// If an error is returned, it must be the last returned value.
	switch {
	case len(outs) == 1 && isErrorType(outs[0]):
		c.errPos = 0
	case len(outs) == 2:
		if isErrorType(outs[0]) || !isErrorType(outs[1]) {
			return nil
		}
		c.errPos = 1
	}
	return c
}

// makeArgTypes composes the argTypes list.
func (c *rpcCallback) makeArgTypes() {
	fntype := c.function.Type()
	// Skip receiver and context.Context parameter (if present).
	firstArg := 0
	if c.receiver.IsValid() {
		firstArg++
	}
	if fntype.NumIn() > firstArg && fntype.In(firstArg) == contextType {
		c.hasCtx = true
		firstArg++
	}
	// Add all remaining parameters.
	c.argTypes = make([]reflect.Type, fntype.NumIn()-firstArg)
	for i := firstArg; i < fntype.NumIn(); i++ {
		c.argTypes[i-firstArg] = fntype.In(i)
	}
}

// call invokes the callback.
func (c *rpcCallback) call(ctx context.Context, method string, args []reflect.Value) (res interface{}, errRes error) {
	// Create the argument slice.
	fullargs := make([]reflect.Value, 0, 2+len(args))
	if c.receiver.IsValid() {
		fullargs = append(fullargs, c.receiver)
	}
	if c.hasCtx {
		fullargs = append(fullargs, reflect.ValueOf(ctx))
	}
	fullargs = append(fullargs, args...)

	// Catch panic while running the callback.
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log15.Error("RPC method " + method + " crashed: " + fmt.Sprintf("%v\n%s", err, buf))
			errRes = errors.New("method handler crashed")
		}
	}()
	// Run the callback.
	results := c.function.Call(fullargs)
	if len(results) == 0 {
		return nil, nil
	}
	if c.errPos >= 0 && !results[c.errPos].IsNil() {
		// Method has returned non-nil error value.
		err := results[c.errPos].Interface().(error)
		return reflect.Value{}, err
	}
	return results[0].Interface(), nil
}

// Is t context.Context or *context.Context?
func isContextType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == contextType
}

// Does t satisfy the error interface?
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

// Is t Subscription or *Subscription?
func isSubscriptionType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == subscriptionType
}

// isPubSub tests whether the given method has as as first argument a context.Context and
// returns the pair (Subscription, error).
func isPubSub(methodType reflect.Type) bool {
	// numIn(0) is the receiver type
	if methodType.NumIn() < 2 || methodType.NumOut() != 2 {
		return false
	}
	return isContextType(methodType.In(1)) &&
		isSubscriptionType(methodType.Out(0)) &&
		isErrorType(methodType.Out(1))
}

// formatName converts to first character of name to lowercase.
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}

// subscription returns a subscription callback in the given service.
func (r *adamniteServiceRegistry) subscription(service, name string) *rpcCallback {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.services[service].subscriptions[name]
}
