package validator

import "errors"

var (
	// UnexpectedType is error that can func
	// 	func TypeIs[T any]()
	UnexpectedType = errors.New("Unexpected type")
)