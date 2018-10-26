package util

import "fmt"

type ValueSequenceCheckerType int
const (
	ValueSeqType_SEQUENCE       ValueSequenceCheckerType = 0
	ValueSeqType_AND            ValueSequenceCheckerType = 1
	ValueSeqType_OR             ValueSequenceCheckerType = 2
	ValueSeqType_EXACT_SEQUENCE ValueSequenceCheckerType = 3
)

type ValueSequenceChecker struct {
	values          []int32 // used for sequences
	valuesHitCountMap		map[int32]int //used for OR / AND
	valSeqCheckType ValueSequenceCheckerType
	currentPos      int
}

func (sc *ValueSequenceChecker) String() string {
	if sc.valSeqCheckType == ValueSeqType_EXACT_SEQUENCE || sc.valSeqCheckType == ValueSeqType_SEQUENCE {
		res := "("
		for i,v := range sc.values {
			if i == sc.currentPos { res += "[" }

			res += fmt.Sprintf("%d", v)

			if i == sc.currentPos { res += "]" }
			if i != len(sc.values)-1 { res += " "}
		}
		res += ")"
		return res
	}
	if sc.valSeqCheckType == ValueSeqType_AND {
		return fmt.Sprintf("AND values (%v), still needed (%+v)", sc.values, sc.valuesHitCountMap)
	}
	if sc.valSeqCheckType == ValueSeqType_OR {
		return fmt.Sprintf("OR values (%v)", sc.values)
	}
	return ""
}

func (sc *ValueSequenceChecker) Check(val int32) bool {
	switch sc.valSeqCheckType {
	case ValueSeqType_SEQUENCE:
		if len(sc.values) == 0 { return true }

		if val == sc.values[sc.currentPos] {
			if sc.advance() {
				// rewind, to be able to reuse the sequence checker
				sc.currentPos = 0
				return true
			}
		}
	case ValueSeqType_EXACT_SEQUENCE:
		if len(sc.values) == 0 { return true }
		if val == sc.values[sc.currentPos] {
			if sc.advance() {
				// rewind, to be able to reuse the sequence checker
				sc.currentPos = 0
				return true
			}
		} else {
			sc.currentPos = 0 // rewind if no out-of-order allowed
		}
	case ValueSeqType_AND:
		if remaining,exists := sc.valuesHitCountMap[val]; exists {
			// lower count as we have a hit
			newRemaining := remaining-1
			sc.valuesHitCountMap[val] = newRemaining

			// if count == 0, remove from map
			if newRemaining < 1 {
				delete(sc.valuesHitCountMap, val)
			}
		}

		// if the map is empty, we succeeded
		if len(sc.valuesHitCountMap) == 0 {
			// ToDo: make this thread save
			// rebuild map to be able to reuse
			valuesHitCountMap := make(map[int32]int)
			for _,val := range sc.values {
				if currentCount,exists := valuesHitCountMap[val]; exists {
					valuesHitCountMap[val] = currentCount+1
				} else {
					valuesHitCountMap[val] = 1
				}
			}
			sc.valuesHitCountMap = valuesHitCountMap


			return true
		}
	case ValueSeqType_OR:
		if _,exists := sc.valuesHitCountMap[val]; exists {
			return true //every value from the map would produce a hit
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

func NewValueSequenceChecker(values []int32, valSeqCheckType ValueSequenceCheckerType) *ValueSequenceChecker {
	res := &ValueSequenceChecker{
		values:          values,
		valSeqCheckType: valSeqCheckType,
		currentPos:      0,
	}
	if valSeqCheckType == ValueSeqType_AND || valSeqCheckType == ValueSeqType_OR {
		// convert values to map for easy lookup (and marking for AND)
		res.valuesHitCountMap = make(map[int32]int)
		for _,val := range values {
			if currentCount,exists := res.valuesHitCountMap[val]; exists {
				res.valuesHitCountMap[val] = currentCount+1
			} else {
				res.valuesHitCountMap[val] = 1
			}
		}
	}
	return res
}
