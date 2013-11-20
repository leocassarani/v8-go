package v8

/*
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "reflect"
import "sync"

type PropertyAttribute int

const (
	PA_None       PropertyAttribute = 0
	PA_ReadOnly                     = 1 << 0
	PA_DontEnum                     = 1 << 1
	PA_DontDelete                   = 1 << 2
)

type AccessControl int

// Access control specifications.
//
// Some accessors should be accessible across contexts.  These
// accessors have an explicit access control parameter which specifies
// the kind of cross-context access that should be allowed.
//
// Additionally, for security, accessors can prohibit overwriting by
// accessors defined in JavaScript.  For objects that have such
// accessors either locally or in their prototype chain it is not
// possible to overwrite the accessor by using __defineGetter__ or
// __defineSetter__ from JavaScript code.
//
const (
	AC_DEFAULT               AccessControl = 0
	AC_ALL_CAN_READ                        = 1
	AC_ALL_CAN_WRITE                       = 1 << 1
	AC_PROHIBITS_OVERWRITING               = 1 << 2
)

// A JavaScript object (ECMA-262, 4.3.3)
//
type Object struct {
	*Value
}

func (e *Engine) NewObject() *Value {
	return newValue(C.V8_NewObject(e.self))
}

func (o *Object) SetProperty(key string, value *Value, attribs PropertyAttribute) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return C.V8_Object_SetProperty(
		o.self, (*C.char)(keyPtr), C.int(len(key)), value.self, C.int(attribs),
	) == 1
}

func (o *Object) GetProperty(key string) *Value {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return newValue(C.V8_Object_GetProperty(
		o.self, (*C.char)(keyPtr), C.int(len(key)),
	))
}

func (o *Object) SetElement(index int, value *Value) bool {
	return C.V8_Object_SetElement(
		o.self, C.uint32_t(index), value.self,
	) == 1
}

func (o *Object) GetElement(index int) *Value {
	return newValue(C.V8_Object_GetElement(o.self, C.uint32_t(index)))
}

func (o *Object) GetPropertyAttributes(key string) PropertyAttribute {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return PropertyAttribute(C.V8_Object_GetPropertyAttributes(
		o.self, (*C.char)(keyPtr), C.int(len(key)),
	))
}

// Sets a local property on this object bypassing interceptors and
// overriding accessors or read-only properties.
//
// Note that if the object has an interceptor the property will be set
// locally, but since the interceptor takes precedence the local property
// will only be returned if the interceptor doesn't return a value.
//
// Note also that this only works for named properties.
func (o *Object) ForceSetProperty(key string, value *Value, attribs PropertyAttribute) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return C.V8_Object_ForceSetProperty(o.self,
		(*C.char)(keyPtr), C.int(len(key)), value.self, C.int(attribs),
	) == 1
}

func (o *Object) HasProperty(key string) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return C.V8_Object_HasProperty(
		o.self, (*C.char)(keyPtr), C.int(len(key)),
	) == 1
}

func (o *Object) DeleteProperty(key string) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return C.V8_Object_DeleteProperty(
		o.self, (*C.char)(keyPtr), C.int(len(key)),
	) == 1
}

// Delete a property on this object bypassing interceptors and
// ignoring dont-delete attributes.
func (o *Object) ForceDeleteProperty(key string) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	return C.V8_Object_ForceDeleteProperty(
		o.self, (*C.char)(keyPtr), C.int(len(key)),
	) == 1
}

func (o *Object) HasElement(index int) bool {
	return C.V8_Object_HasElement(
		o.self, C.uint32_t(index),
	) == 1
}

func (o *Object) DeleteElement(index int) bool {
	return C.V8_Object_DeleteElement(
		o.self, C.uint32_t(index),
	) == 1
}

// Returns an array containing the names of the enumerable properties
// of this object, including properties from prototype objects.  The
// array returned by this method contains the same values as would
// be enumerated by a for-in statement over this object.
//
func (o *Object) GetPropertyNames() *Array {
	return newValue(C.V8_Object_GetPropertyNames(o.self)).ToArray()
}

// This function has the same functionality as GetPropertyNames but
// the returned array doesn't contain the names of properties from
// prototype objects.
//
func (o *Object) GetOwnPropertyNames() *Array {
	return newValue(C.V8_Object_GetOwnPropertyNames(o.self)).ToArray()
}

// Get the prototype object.  This does not skip objects marked to
// be skipped by __proto__ and it does not consult the security
// handler.
//
func (o *Object) GetPrototype() *Object {
	return newValue(C.V8_Object_GetPrototype(o.self)).ToObject()
}

// Set the prototype object.  This does not skip objects marked to
// be skipped by __proto__ and it does not consult the security
// handler.
//
func (o *Object) SetPrototype(proto *Object) bool {
	return C.V8_Object_SetPrototype(o.self, proto.self) == 1
}

// An instance of the built-in array constructor (ECMA-262, 15.4.2).
//
type Array struct {
	*Object
}

func (e *Engine) NewArray(length int) *Array {
	return newValue(C.V8_NewArray(
		e.self, C.int(length),
	)).ToArray()
}

func (a *Array) Length() int {
	return int(C.V8_Array_Length(a.self))
}

type RegExpFlags int

// Regular expression flag bits. They can be or'ed to enable a set
// of flags.
//
const (
	RF_None       RegExpFlags = 0
	RF_Global                 = 1
	RF_IgnoreCase             = 2
	RF_Multiline              = 4
)

type RegExp struct {
	*Object
	pattern       string
	patternCached bool
	flags         RegExpFlags
	flagsCached   bool
}

// Creates a regular expression from the given pattern string and
// the flags bit field. May throw a JavaScript exception as
// described in ECMA-262, 15.10.4.1.
//
// For example,
//   NewRegExp("foo", RF_Global | RF_Multiline)
//
// is equivalent to evaluating "/foo/gm".
//
func (e *Engine) NewRegExp(pattern string, flags RegExpFlags) *Value {
	patternPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&pattern)).Data)

	return newValue(C.V8_NewRegExp(
		e.self, (*C.char)(patternPtr), C.int(len(pattern)), C.int(flags),
	))
}

// Returns the value of the source property: a string representing
// the regular expression.
func (r *RegExp) Pattern() string {
	if !r.patternCached {
		cstring := C.V8_RegExp_Pattern(r.self)
		r.pattern = C.GoString(cstring)
		r.patternCached = true
		C.free(unsafe.Pointer(cstring))
	}
	return r.pattern
}

// Returns the flags bit field.
//
func (r *RegExp) Flags() RegExpFlags {
	if !r.flagsCached {
		r.flags = RegExpFlags(C.V8_RegExp_Flags(r.self))
		r.flagsCached = true
	}
	return r.flags
}

// A JavaScript function object (ECMA-262, 15.3).
//
type Function struct {
	*Object
}

type FunctionCallback func(FunctionCallbackInfo)

type FunctionTemplate struct {
	sync.Mutex
	id       int
	engine   *Engine
	callback FunctionCallback
}

func (e *Engine) NewFunctionTemplate(callback FunctionCallback) *FunctionTemplate {
	ft := &FunctionTemplate{
		id:       e.funcTemplateId + 1,
		engine:   e,
		callback: callback,
	}
	e.funcTemplateId += 1
	e.funcTemplates[ft.id] = ft
	return ft
}

func (ft *FunctionTemplate) Dispose() {
	ft.Lock()
	defer ft.Unlock()
	if ft.id > 0 {
		delete(ft.engine.funcTemplates, ft.id)
		ft.id = 0
		ft.engine = nil
	}
}

func (ft *FunctionTemplate) NewFunction() *Value {
	ft.Lock()
	defer ft.Unlock()
	if ft.engine == nil {
		return nil
	}
	return newValue(C.V8_NewFunction(
		ft.engine.self, unsafe.Pointer(&(ft.callback)),
	))
}

//export go_function_callback
func go_function_callback(info, callback unsafe.Pointer) {
	callbackFunc := *(*func(FunctionCallbackInfo))(callback)
	callbackFunc(FunctionCallbackInfo{info, ReturnValue{}})
}

func (f *Function) Call(args ...*Value) *Value {
	argv := make([]unsafe.Pointer, len(args))
	for i, arg := range args {
		argv[i] = arg.self
	}
	return newValue(C.V8_Function_Call(
		f.self, C.int(len(args)),
		unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&argv)).Data),
	))
}

// Function and property return value
//
type ReturnValue struct {
	self unsafe.Pointer
}

func (rv ReturnValue) Set(value *Value) {
	C.V8_ReturnValue_Set(rv.self, value.self)
}

func (rv ReturnValue) SetBoolean(value bool) {
	valueInt := 0
	if value {
		valueInt = 1
	}
	C.V8_ReturnValue_SetBoolean(rv.self, C.int(valueInt))
}

func (rv ReturnValue) SetNumber(value float64) {
	C.V8_ReturnValue_SetNumber(rv.self, C.double(value))
}

func (rv ReturnValue) SetInt32(value int32) {
	C.V8_ReturnValue_SetInt32(rv.self, C.int32_t(value))
}

func (rv ReturnValue) SetUint32(value uint32) {
	C.V8_ReturnValue_SetUint32(rv.self, C.uint32_t(value))
}

func (rv ReturnValue) SetString(value string) {
	valuePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&value)).Data)
	C.V8_ReturnValue_SetString(rv.self, (*C.char)(valuePtr), C.int(len(value)))
}

func (rv ReturnValue) SetNull() {
	C.V8_ReturnValue_SetNull(rv.self)
}

func (rv ReturnValue) SetUndefined() {
	C.V8_ReturnValue_SetUndefined(rv.self)
}

// Function callback info
//
type FunctionCallbackInfo struct {
	self        unsafe.Pointer
	returnValue ReturnValue
}

func (fc FunctionCallbackInfo) Get(i int) *Value {
	return newValue(C.V8_FunctionCallbackInfo_Get(fc.self, C.int(i)))
}

func (fc FunctionCallbackInfo) Length() int {
	return int(C.V8_FunctionCallbackInfo_Length(fc.self))
}

func (fc FunctionCallbackInfo) Callee() *Function {
	return newValue(C.V8_FunctionCallbackInfo_Callee(fc.self)).ToFunction()
}

func (fc FunctionCallbackInfo) This() *Object {
	return newValue(C.V8_FunctionCallbackInfo_This(fc.self)).ToObject()
}

func (fc FunctionCallbackInfo) Holder() *Object {
	return newValue(C.V8_FunctionCallbackInfo_Holder(fc.self)).ToObject()
}

func (fc *FunctionCallbackInfo) ReturnValue() ReturnValue {
	if fc.returnValue.self == nil {
		fc.returnValue.self = C.V8_FunctionCallbackInfo_ReturnValue(fc.self)
	}
	return fc.returnValue
}

type ObjectTemplate struct {
	sync.Mutex
	id        int
	engine    *Engine
	accessors map[string]*accessorInfo
}

type accessorInfo struct {
	key     string
	getter  GetterCallback
	setter  SetterCallback
	attribs PropertyAttribute
}

func (e *Engine) NewObjectTemplate() *ObjectTemplate {
	ot := &ObjectTemplate{
		id:        e.objectTemplateId + 1,
		engine:    e,
		accessors: make(map[string]*accessorInfo),
	}
	e.objectTemplateId += 1
	e.objectTemplates[ot.id] = ot
	return ot
}

func (ot *ObjectTemplate) Dispose() {
	ot.Lock()
	defer ot.Unlock()

	if ot.id > 0 {
		delete(ot.engine.objectTemplates, ot.id)
		ot.id = 0
		ot.engine = nil
	}
}

func (ot *ObjectTemplate) NewObject() *Value {
	ot.Lock()
	defer ot.Unlock()

	if ot.engine == nil {
		return nil
	}

	value := ot.engine.NewObject()

	if value == nil {
		return nil
	}

	object := value.ToObject()
	for _, info := range ot.accessors {
		object.setAccessor(info)
	}

	return value
}

func (ot *ObjectTemplate) WrapObject(value *Value) {
	ot.Lock()
	defer ot.Unlock()

	object := value.ToObject()
	for _, info := range ot.accessors {
		object.setAccessor(info)
	}
}

func (ot *ObjectTemplate) SetAccessor(key string, getter GetterCallback, setter SetterCallback, attribs PropertyAttribute) {
	ot.accessors[key] = &accessorInfo{
		key:     key,
		getter:  getter,
		setter:  setter,
		attribs: attribs,
	}
}

// Property getter callback info
//
type GetterCallbackInfo struct {
	self        unsafe.Pointer
	returnValue ReturnValue
}

func (g GetterCallbackInfo) This() *Object {
	return newValue(C.V8_GetterCallbackInfo_This(g.self)).ToObject()
}

func (g GetterCallbackInfo) Holder() *Object {
	return newValue(C.V8_GetterCallbackInfo_Holder(g.self)).ToObject()
}

func (g *GetterCallbackInfo) ReturnValue() ReturnValue {
	if g.returnValue.self == nil {
		g.returnValue.self = C.V8_GetterCallbackInfo_ReturnValue(g.self)
	}
	return g.returnValue
}

// Property setter callback info
//
type SetterCallbackInfo struct {
	self unsafe.Pointer
}

func (s SetterCallbackInfo) This() *Object {
	return newValue(C.V8_SetterCallbackInfo_This(s.self)).ToObject()
}

func (s SetterCallbackInfo) Holder() *Object {
	return newValue(C.V8_SetterCallbackInfo_Holder(s.self)).ToObject()
}

type GetterCallback func(name string, info GetterCallbackInfo)

type SetterCallback func(name string, value *Value, info SetterCallbackInfo)

//export go_getter_callback
func go_getter_callback(key *C.char, length C.int, info, callback unsafe.Pointer) {
	name := &reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(key)),
		Len:  int(length),
	}
	gname := *((*string)(unsafe.Pointer(name)))
	(*(*GetterCallback)(callback))(gname, GetterCallbackInfo{info, ReturnValue{}})
}

//export go_setter_callback
func go_setter_callback(key *C.char, length C.int, value, info, callback unsafe.Pointer) {
	name := &reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(key)),
		Len:  int(length),
	}
	gname := *((*string)(unsafe.Pointer(name)))
	(*(*SetterCallback)(callback))(gname, newValue(value), SetterCallbackInfo{info})
}

func (o *Object) setAccessor(info *accessorInfo) bool {
	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&info.key)).Data)
	return C.V8_Object_SetAccessor(
		o.self,
		(*C.char)(keyPtr), C.int(len(info.key)),
		unsafe.Pointer(&(info.getter)),
		unsafe.Pointer(&(info.setter)),
		C.int(info.attribs),
	) == 1
}