package utils

// 常用的基本类型转指针，指针转基本类型工具函数

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// Int8Ptr returns a pointer to the given int8.
func Int8Ptr(i int8) *int8 {
	return &i
}

// Int16Ptr returns a pointer to the given int16.
func Int16Ptr(i int16) *int16 {
	return &i
}

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int {
	return &i
}

// Int32Ptr returns a pointer to the given int32.
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr returns a pointer to the given int64.
func Int64Ptr(i int64) *int64 {
	return &i
}

// Uint8Ptr returns a pointer to the given uint8.
func Uint8Ptr(u uint8) *uint8 {
	return &u
}

// Uint16Ptr returns a pointer to the given uint16.
func Uint16Ptr(u uint16) *uint16 {
	return &u
}

// UintPtr returns a pointer to the given uint.
func UintPtr(u uint) *uint {
	return &u
}

// Uint32Ptr returns a pointer to the given uint32.
func Uint32Ptr(u uint32) *uint32 {
	return &u
}

// Uint64Ptr returns a pointer to the given uint64.
func Uint64Ptr(u uint64) *uint64 {
	return &u
}

// UintptrPtr returns a pointer to the given uintptr.
func UintptrPtr(u uintptr) *uintptr {
	return &u
}

// BytePtr returns a pointer to the given byte.
func BytePtr(b byte) *byte {
	return &b
}

// RunePtr returns a pointer to the given rune.
func RunePtr(r rune) *rune {
	return &r
}

// Float32Ptr returns a pointer to the given float32.
func Float32Ptr(f float32) *float32 {
	return &f
}

// Float64Ptr returns a pointer to the given float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

// Complex64Ptr returns a pointer to the given complex64.
func Complex64Ptr(c complex64) *complex64 {
	return &c
}

// Complex128Ptr returns a pointer to the given complex128.
func Complex128Ptr(c complex128) *complex128 {
	return &c
}

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool {
	return &b
}

// PtrStringValue returns the dereferenced string value, or an empty string if nil.
func PtrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PtrInt8Value returns the dereferenced int8 value, or 0 if nil.
func PtrInt8Value(i *int8) int8 {
	if i == nil {
		return 0
	}
	return *i
}

// PtrInt16Value returns the dereferenced int16 value, or 0 if nil.
func PtrInt16Value(i *int16) int16 {
	if i == nil {
		return 0
	}
	return *i
}

// PtrIntValue returns the dereferenced int value, or 0 if nil.
func PtrIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// PtrInt32Value returns the dereferenced int32 value, or 0 if nil.
func PtrInt32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// PtrInt64Value returns the dereferenced int64 value, or 0 if nil.
func PtrInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// PtrUint8Value returns the dereferenced uint8 value, or 0 if nil.
func PtrUint8Value(u *uint8) uint8 {
	if u == nil {
		return 0
	}
	return *u
}

// PtrUint16Value returns the dereferenced uint16 value, or 0 if nil.
func PtrUint16Value(u *uint16) uint16 {
	if u == nil {
		return 0
	}
	return *u
}

// PtrUintValue returns the dereferenced uint value, or 0 if nil.
func PtrUintValue(u *uint) uint {
	if u == nil {
		return 0
	}
	return *u
}

// PtrUint32Value returns the dereferenced uint32 value, or 0 if nil.
func PtrUint32Value(u *uint32) uint32 {
	if u == nil {
		return 0
	}
	return *u
}

// PtrUint64Value returns the dereferenced uint64 value, or 0 if nil.
func PtrUint64Value(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}

// PtrUintptrValue returns the dereferenced uintptr value, or 0 if nil.
func PtrUintptrValue(u *uintptr) uintptr {
	if u == nil {
		return 0
	}
	return *u
}

// PtrByteValue returns the dereferenced byte value, or 0 if nil.
func PtrByteValue(b *byte) byte {
	if b == nil {
		return 0
	}
	return *b
}

// PtrRuneValue returns the dereferenced rune value, or 0 if nil.
func PtrRuneValue(r *rune) rune {
	if r == nil {
		return 0
	}
	return *r
}

// PtrFloat32Value returns the dereferenced float32 value, or 0 if nil.
func PtrFloat32Value(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

// PtrFloat64Value returns the dereferenced float64 value, or 0 if nil.
func PtrFloat64Value(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// PtrComplex64Value returns the dereferenced complex64 value, or 0 if nil.
func PtrComplex64Value(c *complex64) complex64 {
	if c == nil {
		return 0
	}
	return *c
}

// PtrComplex128Value returns the dereferenced complex128 value, or 0 if nil.
func PtrComplex128Value(c *complex128) complex128 {
	if c == nil {
		return 0
	}
	return *c
}

// PtrBoolValue returns the dereferenced bool value, or false if nil.
func PtrBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
