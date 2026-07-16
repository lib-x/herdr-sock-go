package herdrsock

import (
	"encoding/json"
	"fmt"
)

type popupSizeKind uint8

const (
	popupSizeCells popupSizeKind = iota
	popupSizePercent
)

// PopupSize is an outer popup dimension encoded as terminal cells or a
// percentage of the available terminal area.
type PopupSize struct {
	kind  popupSizeKind
	value uint16
}

// PopupCells returns a popup size measured in terminal cells.
func PopupCells(cells uint16) PopupSize {
	return PopupSize{kind: popupSizeCells, value: cells}
}

// PopupPercent returns a popup size between 1% and 100%.
func PopupPercent(percent uint8) (PopupSize, error) {
	if percent < 1 || percent > 100 {
		return PopupSize{}, fmt.Errorf("herdrsock: popup percent must be between 1 and 100, got %d", percent)
	}
	return PopupSize{kind: popupSizePercent, value: uint16(percent)}, nil
}

func (s PopupSize) MarshalJSON() ([]byte, error) {
	if s.kind == popupSizePercent {
		return json.Marshal(fmt.Sprintf("%d%%", s.value))
	}
	return json.Marshal(s.value)
}
