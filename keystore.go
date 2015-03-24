package emissary

// Delivery/Security modules can have properties that implement this interface in order to abstract key retrieval to the developer rather than the module.
type Keystore interface {
	GetKey(id string) ([]byte, error)
}
