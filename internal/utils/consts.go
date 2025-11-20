package utils

type BuilderState int

const (
	StateInitial BuilderState = iota
	StateBodyBuilt
	StateHeadersSet
	StateRequestBuilt
	StateExecuted
)
