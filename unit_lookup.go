package dateparse

// UnitField identifies which field of a Duration a unit maps to.
type UnitField int

// Exported UnitField constants mapping to internal duration fields.
const (
	FieldYears   UnitField = UnitField(fieldYears)
	FieldMonths  UnitField = UnitField(fieldMonths)
	FieldDays    UnitField = UnitField(fieldDays)
	FieldHours   UnitField = UnitField(fieldHours)
	FieldMinutes UnitField = UnitField(fieldMinutes)
	FieldSeconds UnitField = UnitField(fieldSeconds)
	FieldNanos   UnitField = UnitField(fieldNanos)
)

// UnitInfo holds the exported field and scale for a unit.
type UnitInfo struct {
	Field UnitField
	Scale int
}

// LookupUnit returns the field and scale for a lowercase unit name.
func LookupUnit(name string) (UnitInfo, bool) {
	e, ok := unitTable[name]
	if !ok {
		return UnitInfo{}, false
	}
	return UnitInfo{Field: UnitField(e.field), Scale: e.scale}, ok
}
