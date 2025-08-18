// Package access provides utilities for determining the accessibility of types in Go.
package access

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Accessibility int

const (
	// Accessibility is not defined
	AccessibilityUndefined Accessibility = iota

	// The type is not accessible
	NotAccessible

	// The type is accessible only within its own package
	AccessibleInsidePackage

	// The type is accessible from any package
	AccessibleEverywhere
)

type Namedness int

const (
	NamednessUndefined Namedness = iota
	AnonymousType
	NamedType
)

func (t Accessibility) String() string {
	switch t {
	case AccessibilityUndefined:
		return "accessibility undefined"
	case NotAccessible:
		return "not accessible"
	case AccessibleInsidePackage:
		return "accessible inside package"
	case AccessibleEverywhere:
		return "accessible everywhere"
	default:
		err, _ := fmt.Printf("unknown accessibility: %d\n", t)
		panic(err)
	}
}

func (t Namedness) String() string {
	switch t {
	case NamednessUndefined:
		return "namedness undefined"
	case AnonymousType:
		return "anonymous type"
	case NamedType:
		return "named type"
	default:
		err, _ := fmt.Printf("unknown namedness: %d\n", t)
		panic(err)
	}
}

func Info[T any](_ T) (Namedness, Accessibility) {
	rt := reflect.TypeFor[T]()

	return getNamedness(rt), getAccessability(rt)
}

func getAccessability(rt reflect.Type) Accessibility {
	callerFunc := getCallerFuncName(3)
	callerPkg := extractCallerPKG(callerFunc)

	if accessibleEverywhere(rt) {
		return AccessibleEverywhere
	}

	if accessibleFromPackage(rt, callerPkg) {
		return AccessibleInsidePackage
	}
	return NotAccessible
}

func getNamedness(rt reflect.Type) Namedness {
	for rt != nil && rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt != nil && rt.Name() != "" {
		return NamedType
	}

	return AnonymousType
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// accessibleEverywhere reports whether rt can be accessed from any package other than its defining one.
//
// For named types: only exported types are accessible (predeclared types are also accessible).
//
// For unnamed composite types: all referenced named types must be exported or predeclared.
//
//	accessibleEverywhere(rt) // do not pass the seen param
func accessibleEverywhere(rt reflect.Type, seen ...map[reflect.Type]bool) bool {
	if seen == nil {
		seen = []map[reflect.Type]bool{
			make(map[reflect.Type]bool),
		}
	}

	if rt == nil {
		return true
	}

	if seen[0][rt] {
		return true
	}

	// avoid infinite recursion
	seen[0][rt] = true

	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if name := rt.Name(); name != "" {
		// Predeclared types are accessible
		if rt.PkgPath() == "" {
			return true
		}
		return isExported(name)
	}

	// Handle unnamed types
	switch rt.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Pointer:
		return accessibleEverywhere(rt.Elem(), seen...)
	case reflect.Map:
		return accessibleEverywhere(rt.Key(), seen...) && accessibleEverywhere(rt.Elem(), seen...)
	case reflect.Func:
		for i := 0; i < rt.NumIn(); i++ {
			if !accessibleEverywhere(rt.In(i), seen...) {
				return false
			}
		}
		for i := 0; i < rt.NumOut(); i++ {
			if !accessibleEverywhere(rt.Out(i), seen...) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			if !accessibleEverywhere(rt.Field(i).Type, seen...) {
				return false
			}
		}
		return true
	case reflect.Interface:
		for i := 0; i < rt.NumMethod(); i++ {
			if !accessibleEverywhere(rt.Method(i).Type, seen...) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// acessableFromPackage reports whether rt can be accessed from the specified package.
//
// For named types: exported types are accessible from anywhere; unexported types are accessible only within their defining package. Predeclared types are accessible.
//
// For unnamed composite types: all referenced named types must be accessible from pkg.
//
//	accessibleFromPackage(rt) // do not pass the seen param
func accessibleFromPackage(rt reflect.Type, pkgPath string, seen ...map[reflect.Type]bool) bool {
	if seen == nil {
		seen = []map[reflect.Type]bool{
			make(map[reflect.Type]bool),
		}
	}

	if rt == nil {
		return true
	}

	if seen[0][rt] {
		return true
	}

	// avoid infinite recursion
	seen[0][rt] = true

	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if name := rt.Name(); name != "" {
		// Predeclared types are accessible
		if rt.PkgPath() == "" {
			return true
		}
		// Exported types are accessible from anywhere
		if isExported(name) {
			return true
		}

		// Unexported named types are accessible only within their package
		return rt.PkgPath() == pkgPath
	}

	// Handle unnamed types
	switch rt.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Pointer:
		return accessibleFromPackage(rt.Elem(), pkgPath, seen...)
	case reflect.Map:
		return accessibleFromPackage(rt.Key(), pkgPath, seen...) && accessibleFromPackage(rt.Elem(), pkgPath, seen...)
	case reflect.Func:
		for i := 0; i < rt.NumIn(); i++ {
			if !accessibleFromPackage(rt.In(i), pkgPath, seen...) {
				return false
			}
		}
		for i := 0; i < rt.NumOut(); i++ {
			if !accessibleFromPackage(rt.Out(i), pkgPath, seen...) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			if !accessibleFromPackage(rt.Field(i).Type, pkgPath, seen...) {
				return false
			}
		}
		return true
	case reflect.Interface:
		for i := 0; i < rt.NumMethod(); i++ {
			if !accessibleFromPackage(rt.Method(i).Type, pkgPath, seen...) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// getCallerFuncName returns the name of the caller function/frame at the given skip.
func getCallerFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return fn.Name()
}

// extractCallerPKG returns the package import path from runtime.FuncForPC(pc).Name()
func extractCallerPKG(callerFunc string) string {
	if callerFunc == "" {
		return ""
	}

	var start int
	// look for the last slash in path
	if slash: = strings.LastIndex(callerFunc, "/"); slash >= 0 {
		start = slash + 1

		// return everything after it until the first dot // ex. "full/import/path" for "full/import/path.(*Type).Method" 
		if dot := strings.Index(callerFunc[start:], "."); dot >= 0 {
			return callerFunc[:start+dot]
		}
	}
	

	// no slash; best-effort trim at the last dot. // ex. "noslashmod" for "noslashmod.Function"
	if dot := strings.LastIndex(callerFunc, "."); dot >= 0 {
		return callerFunc[:dot]
	}

	// main.main for example
	return callerFunc
}
