package reg

import "github.com/mp3cko/registry/access"

// WithRegistry use the given registry for a single op
//
// Valid:
//
//	Get[T](WithRegistry(r)) // get from r
//
//	Set(val, WithRegistry(r)) // set in r
//
//	GetAll(WithRegistry(r)) // get all from r
//
//	Unset[T](WithRegistry(r)) // unsert from r
//
// Invalid:
//
//	NewRegistry(WithRegistry(r)) // returns ErrNotSupported, use cloning options for that
func (t *optionsBuilder) WithRegistry(r *registry) *optionsBuilder {
	return t.and(withRegistryOption(r))
}

// WithUniqueType is a unique constraint on the type, it ensures that each type can be registered only once
//
// Valid:
//
//	NewRegistry(WithUniqueType()) // ensures that only a single instance can be registered per type inside the registry
//
//	Get[T](WithUniqueType()) // returns ErrNotUniqueType if multiple instances are already registered, which can happen if option was not passed in the constructor)
//
//	Set(val, WithUniqueType()) // returns ErrNotUniqueType if an instance is already registered for a type
//
//	GetAll(WithUniqueType()) // get all unique instances
//
//	Unset[T](WithUniqueType()) // returns ErrNotUniqueType if type is not unique
func (t *optionsBuilder) WithUniqueType() *optionsBuilder {
	return t.and(withUniqueTypeOption())
}

// WithUniqueName is a unique constraint on name (per type)
//
// Valid:
//
//	NewRegistry(WithUniqueName()) // ensures that a name can only be registered once per type inside the registry
//
//	Set(val, WithUniqueName()) // returns ErrNotUniqueName if an instance with the same name is already registered
//
// Invalid:
//
//	Get[T](WithUniqueName()) // returns ErrNotSupported, use WithUniqueType()
//
//	GetAll(WithUniqueName()) // returns ErrNotSupported, use WithUniqueType()
//
//	Unset[T](WithUniqueName()) // returns ErrNotSupported, use WithUniqueType()
func (t *optionsBuilder) WithUniqueName() *optionsBuilder {
	return t.and(withUniqueNamesOption())
}

// WithName defines instance name for operation
// It is supported in all operations
//
// Valid:
//
//	NewRegistry(WithName("example")) // sets the default name for instances inside the registry (if not set, default name is an empty string)
//
//	Get[T](WithName("example")) // returns the instance with the name "example" if it exists
//
//	Set(val, WithName("example")) // sets the instance with the name "example" inside the registry
//
//	GetAll(WithName("example")) // returns the instance with the name "example" if it exists
//
//	Unset[T](WithName("example")) // unsets the instance of T with name "example"
func (t *optionsBuilder) WithName(n string) *optionsBuilder {
	return t.and(withNameOption(n))

}

// WithAccessibility enforce minimum accessibility level of types.
//
// Best used at registry level, it will then require all types to be **at least** as visible as the setting.
// The default registry requires minimum of
//
//	access.AccessibleInsidePackage
//
// This is very useful because it is possible to register a type that you cannot name and therefore it is impossible to retrieve it later (except with GetAll but that is not type safe).
//
// This can happen when you hold an external unexported type. You can work around this by setting it as an interface (either provided by external package or your own). For example:
//
//	innaccesibleType := external.New() /** an innaccesible external type, for example: */ *external.privateType
//	Set[external.AccessibleInterface](innaccesibleType)
//
// if done like that later you can retrieve it using
//
//	Get[external.AccessibleInterface]()
//
// # Valid:
//
//	NewRegistry(WithAccessibility(access.AccessibleEverywhere)) // makes sure that all registered instances are accessible everywhere
//
//	Set(val, WithAccessibility(access.AccessibleInsidePackage)) // make sure that the instance being set is accessible at least in the package where it is registered
//
//	GetAll(WithAccessibility(access.AccessibleInsidePackage)) // returns all instances with the given accessibility. Types are checked for acessibility in the callers package
//
// # Invalid:
//
//	Get[T](WithAccessibility(access.AccessibleEverywhere) // returns ErrNotSupported as it doesnt have a valid use case. The type is already registered and if you can name it, it is accessible
//
//	Unset[T](WithAccessibility(access.AccessibleEverywhere) // returns ErrNotSupported as it doesnt have a valid use case. The type is already registered and if you can name it, it is accessible
func (t *optionsBuilder) WithAccessibility(a access.Accessibility) *optionsBuilder {
	return t.and(withAccessibilityOption(a))
}

// WithNamedness controls if unnamed(anonymous types) are allowed. Primitive types are always allowed.
//
// It can lead to unexpected behavior and is not ergonomic. For example:
//
//	type someInterface{ someMethod() }  // named interface
//	Set[interface{ someMethod() }](someInterface(nil)) // registered under an anonymous type
//	Get[someInterface]() // won't return the instance from above
//
// # Valid:
//
//	NewRegistry(WithNamedness(access.NamedType)) // ensures all registered types are named types
//
//	Set(val, WithNamedness(access.NamedType)) // fails if val has an anonymous type
//
//	GetAll(WithNamedness(access.AnonymousType)) // returns all instances with anonymous types
//
// # Invalid:
//
//	Get[T](WithNamedness(access.NamedType)) // returns ErrNotSupported as it doesn't have a valid use case. You are not constraining namedness here since you know exactly what you are passing to Get
//
//	Unset[T](WithNamedness(access.NamedType)) // returns ErrNotSupported as it doesn't have a valid use case. You are not constraining namedness here since you know exactly what you are passing to Unset
func (t *optionsBuilder) WithNamedness(n access.Namedness) *optionsBuilder {
	return t.and(withNamednessOption(n))
}

// WithCloneConfig copies configuration from the provided registry
//
// Valid:
//
//	// If called with other registry modifying options that conflict with source will throw ErrBadOption
//	NewRegistry(WithCloneConfig(src)) // copies config from src to the new registry
//
// Invalid:
//
//	Get[T](WithCloneConfig(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneConfig(src)) // returns ErrNotSupported
//
//	GetAll(WithCloneConfig(src)) // returns ErrNotSupported
//
//	Unset[T](WithCloneConfig(src)) // returns ErrNotSupported
//
// This option is applied 3rd to last, just before [WithCloneRegistry] and [WithCloneEntries]
func (t *optionsBuilder) WithCloneConfig(src *registry) *optionsBuilder {
	return t.and(withCloneConfigOption(src))
}

// WithCloneEntries copies all entries from the provided registry into the new registry.
//
// Valid:
//
//	NewRegistry(WithCloneEntries(src)) // copies all entries from src to the new registry
//
// Invalid:
//
//	Get[T](WithCloneEntries(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneEntries(src)) // returns ErrNotSupported
//
//	GetAll(WithNamedness(access.AnonymousType)) // returns ErrNotSupported
//
//	Unset[T](WithNamedness(access.NamedType)) // returns ErrNotSupported
//
// This option is applied second to last, before [WithCloneRegistry]
func (t *optionsBuilder) WithCloneEntries(src *registry) *optionsBuilder {
	return t.and(withCloneEntriesOption(src))
}

// WithCloneRegistry copies configuration and entries from the provided registry
//
// Valid:
//
//	// If called with other registry modifying options that conflict with source will throw ErrBadOption
//	NewRegistry(WithCloneRegistry(src)) // copies config from src to the new registry
//
// Invalid:
//
//	Get[T](WithCloneConfig(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneConfig(src)) // returns ErrNotSupported
//
//	GetAll(WithCloneConfig(src)) // returns ErrNotSupported
//
//	Unset[T](WithCloneConfig(src)) // returns ErrNotSupported
//
// This option always applies last to check if other incompatible options have been called before it
func (t *optionsBuilder) WithCloneRegistry(src *registry) *optionsBuilder {
	return t.and(withCloneRegistryOption(src))
}
