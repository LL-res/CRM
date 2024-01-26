package errs

import (
	"errors"
	"fmt"
)

var (
	UNDEFINED_COLLECTOR   = errors.New("undefined collector")
	NO_SUFFICENT_DATA_RAW = errors.New("no sufficient data")
	UNREADY_TO_PREDICT    = errors.New("the model is not ready to predict")
	TRAINING              = errors.New("the model is training")
)
var NO_SUFFICENT_DATA = new(NoSufficentErr)

type NoSufficentErr struct {
	dataCap int
}

func (e *NoSufficentErr) Error() string {
	return fmt.Sprintf("no sufficient data,data cap : %d", e.dataCap)
}
func (e *NoSufficentErr) SetCap(cap int) {
	e.dataCap = cap
}
func (e *NoSufficentErr) GetCap() int {
	return e.dataCap
}
