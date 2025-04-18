package txbarrier

// Operation is a type used for representing the operation
// of distributed transactions branches.
type Operation string

const (
	// Try represents the first phase of tcc.
	Try Operation = "try"
	// Confirm represents the second phase "confirm" of tcc.
	Confirm Operation = "confirm"
	// Cancel represents the second phase "cancel" of tcc.
	Cancel Operation = "cancel"
)
