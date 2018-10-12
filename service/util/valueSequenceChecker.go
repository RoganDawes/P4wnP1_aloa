package util

type ValueSequenceChecker struct {
	values          []int32
	allowOutOfOrder bool
	currentPos      int
}

func (sc *ValueSequenceChecker) Check(val int32) bool {
	if len(sc.values) == 0 { return true }

	if val == sc.values[sc.currentPos] {
		if sc.advance() {
			// rewind, to be able to reuse the sequence checker
			sc.currentPos = 0
			return true
		}
	} else {
		if !sc.allowOutOfOrder {
			sc.currentPos = 0 // rewind if no out-of-order allowed
		}
	}

	return false
}

func (sc *ValueSequenceChecker) advance() bool {
	if sc.currentPos++; sc.currentPos >= len(sc.values) {
		return true
	}
	return false
}

func NewValueSequenceChecker(values []int32, allowOutOfOrder bool) *ValueSequenceChecker {
	res := &ValueSequenceChecker{
		values:          values,
		allowOutOfOrder: allowOutOfOrder,
		currentPos:      0,
	}
	return res
}
