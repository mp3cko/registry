package reg

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
// This option respects other options and is applied near the end of the chain
func (t *optionsBuilder) WithCloneEntries(src *registry) *optionsBuilder {
	return t.and(withCloneEntriesOption(src))
}

// WithCloneConfig copies configuration from the provided registry
//
// Valid:
//
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
// This option is applied near the end of the chain
func (t *optionsBuilder) WithCloneConfig(src *registry) *optionsBuilder {
	return t.and(withCloneConfigOption(src))
}

// WithCloneRegistry copies configuration and entries from the provided registry
//
// Valid:
//
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
