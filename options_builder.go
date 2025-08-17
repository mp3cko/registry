package reg

// WithRegistry uses the provided registry for a single operation
//
//	Get[T](WithRegistry(r)) // uses r for this Get call
//
//	Set(val, WithRegistry(r)) // uses r for this Set call
//
//	NewRegistry(WithRegistry(r)) // returns ErrNotSupported
func (t *optionsBuilder) WithRegistry(r *registry) *optionsBuilder {
	return t.and(withRegistryOption(r))
}

// WithUniqueTypes enforces only one instance per type.
//
//	Get[T](WithUniqueTypes()) // returns ErrNotUnique if multiple instances of a type are already registered, which can happen if option was not passed in the constructor)
//
//	Set(val, WithUniqueTypes()) // returns `ErrNotUnique` if a type is already registered
//
//	NewRegistry(WithUniqueTypes()) // ensures that only a single instance of a type can be registered in the registry
func (t *optionsBuilder) WithUniqueTypes() *optionsBuilder {
	return t.and(withUniqueTypesOption())
}

// WithUniqueNames enforces that names are unique within their type.
//
//	Get[T](WithUniqueNames()) // returns ErrNotSupported
//
//	Set(val, WithUniqueNames()) // returns ErrNotUniqueName if an instance with the same name is already registered
//
//	NewRegistry(WithUniqueNames()) // ensures that a name can only be registered once per type inside the registry
func (t *optionsBuilder) WithUniqueNames() *optionsBuilder {
	return t.and(withUniqueNamesOption())
}

// WithName specifies a name for the registration or lookup.
//
//	Get[T](WithName("example")) // returns the instance with the name "example" if it exists
//
//	Set(val, WithName("example")) // sets the instance with the name "example" inside the registry
//
//	NewRegistry(WithName("example")) // sets the default name for type instances inside the registry (if not set, default name is an empty string)
func (t *optionsBuilder) WithName(n string) *optionsBuilder {
	return t.and(withNameOption(n))

}

// WithCloneEntries copies all entries from the provided registry into the new registry. It is allowed only inside NewRegistry constructor.
//
//	Get[T](WithCloneEntries(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneEntries(src)) // returns ErrNotSupported
//
//	NewRegistry(WithCloneEntries(src)) // copies all instances from src to the new registry (but not its options)
//
// This option respects other options(such as WithName) and therefore applies last.
func (t *optionsBuilder) WithCloneEntries(src *registry) *optionsBuilder {
	return t.and(withCloneEntriesOption(src))
}

// WithCloneConfig copies configuration from the provided registry
//
//	Get[T](WithCloneConfig(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneConfig(src)) // returns ErrNotSupported
//
//	NewRegistry(WithCloneConfig(src)) // copies config from src to the new registry
//
// This option respects other options(such as WithName which will change the default instance name for all copied values) and therefore applies last.
func (t *optionsBuilder) WithCloneConfig(src *registry) *optionsBuilder {
	return t.and(withCloneConfigOption(src))
}

// WithCloneRegistry copies configuration and entries from the provided registry
//
//	Get[T](WithCloneRegistry(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneRegistry(src)) // returns ErrNotSupported
//
//	NewRegistry(WithCloneRegistry(src)) // copies config from src to the new registry
//
// This option always applies last to check if other incompatible options have been called before it
func (t *optionsBuilder) WithCloneRegistry(src *registry) *optionsBuilder {
	return t.and(withCloneRegistryOption(src))
}
