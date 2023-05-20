package core

type Filter struct {
	// Only broadcast messages with kind in list.
	Kinds []Kind `json:"kinds,omitempty"`
}
