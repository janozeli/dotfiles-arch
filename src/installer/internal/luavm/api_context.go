package luavm

// contextStore holds inter-task data for a single unit.
type contextStore struct {
	data map[string]any
}

func newContextStore() *contextStore {
	return &contextStore{data: make(map[string]any)}
}

// GetData returns the value for a key from the context store.
func (c *contextStore) GetData(key string) (any, bool) {
	val, ok := c.data[key]
	return val, ok
}

// SetData stores a value in the context store.
func (c *contextStore) SetData(key string, value any) {
	c.data[key] = value
}
