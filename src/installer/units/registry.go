package units

var registry []Unit

// Register adds a unit to the global registry.
func Register(u Unit) {
	registry = append(registry, u)
}

// All returns all registered units.
func All() []Unit {
	return registry
}
