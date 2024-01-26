package scaler

import "errors"

var (
	ErrTargetSmallerThanCurrent = errors.New("target replica num is smaller than the current")
	ErrTargetBiggerThanCurrent  = errors.New("target replica num is bigger than the current")
)
