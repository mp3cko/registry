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

// WithUniqueInstance enforces only one instance per type.
//
//	Get[T](WithUniqueInstance()) // returns ErrNotUnique if multiple instances are already registered, which can happen if option was not passed in the constructor)
//
//	Set(val, WithUniqueInstance()) // returns `ErrNotUnique` if an instance is already registered
//
//	NewRegistry(WithUniqueInstance()) // ensures that only a single instance can be registered per type inside the registry
func (t *optionsBuilder) WithUniqueInstance() *optionsBuilder {
	return t.and(withUniqueInstanceOption())
}

// WithName specifies a name for the registration or lookup.
//
//	Get[T](WithName("example")) // returns the instance with the name "example" if it exists
//
//	Set(val, WithName("example")) // sets the instance with the name "example" inside the registry
//
//	NewRegistry(WithName("example")) // sets the default name for instances inside the registry (if not set, default name is an empty string)
func (t *optionsBuilder) WithName(n string) *optionsBuilder {
	return t.and(withNameOption(n))

}

// WithCloneStore copies all entries from the provided registry into the new registry. It is allowed only inside NewRegistry constructor.
//
//	Get[T](WithCloneStore(src)) // returns ErrNotSupported
//
//	Set(val, WithCloneStore(src)) // returns ErrNotSupported
//
//	NewRegistry(WithCloneStore(src)) // copies all instances from src to the new registry (but not its options)
//
// This option respects other options(such as WithName) and therefore applies last.
func (t *optionsBuilder) WithCloneStore(src *registry) *optionsBuilder {
	return t.and(withCloneStoreOption(src))
}

func (t *optionsBuilder) and(opts ...*option) *optionsBuilder {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.o = append(t.o, opts...)

	return t
}
