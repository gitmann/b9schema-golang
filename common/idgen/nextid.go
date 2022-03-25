package idgen

var lastID int

// NextID is used internally to generate the next element ID
func NextID() int {
	lastID++
	return lastID
}

// Reset resets the ID counter to zero.
func Reset() {
	lastID = 0
}
