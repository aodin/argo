package relational

import (
	"github.com/aodin/argo"
	"github.com/aodin/aspect"
	"github.com/aodin/aspect/postgis"
	"net/url"
	"strings"
)

type StringSet []string

func (set StringSet) Has(s string) bool {
	for _, elem := range set {
		if elem == s {
			return true
		}
	}
	return false
}

type LocationField struct {
	latitude  string
	longitude string
	location  string
}

// Resource is a relational resource generated from an aspect Table
// TODO defaults:
type Resource struct {
	Name           string
	table          *aspect.TableElem
	sortable       StringSet // acceptable columns for sorting, TODO use map?
	pk             []string
	intFields      []string
	stringFields   []string
	floatFields    []string
	timeFields     []string
	boolFields     []string
	locationFields []LocationField
	Limit          int // 0 indications no limit
	Radius         int // TODO This should be location field specific
}

// Select will construct an aspect SELECT statement from the given parameters.
// TODO perhaps (Values, *SelectStmt, *Meta) error ?
func (t *Resource) Select(p url.Values) (s aspect.SelectStmt, meta argo.Meta) {
	// Initialize the meta
	meta = argo.Meta{}

	// Does this resource need to handle locations?
	if len(t.locationFields) > 0 {
		// Location fields are dropped by default
		// TODO geojson output
		var exceptions []aspect.ColumnElem
		for _, lf := range t.locationFields {
			exceptions = append(exceptions, t.table.C[lf.location])
		}
		s = t.table.SelectExcept(exceptions...)

		// Was a location requested?
		// TODO Parse multiple location fields
		loc := t.locationFields[0]
		lat, lng, err := ParseLatLng(p.Get(loc.latitude), p.Get(loc.longitude))
		if err == LatLngUnparsed {
			// Do nothing
		} else if err != nil {
			// TODO What to do on error?
		} else {
			// TODO radius field should not be hard-coded
			radius := ParseIntOrDefault(p.Get("radius"), t.Radius)

			// Set the location meta information
			meta["radius"] = radius
			meta["latitude"] = lat
			meta["longitude"] = lng

			// Lat and Lng were parsed successfully
			// Create a where statement
			s = s.Where(postgis.DWithin(
				t.table.C[loc.location],
				postgis.LatLong{lat, lng},
				radius,
			))
		}
	} else {
		s = t.table.Select()
	}

	// Determine ordering
	// TODO order should not be hardcoded
	ordering := t.Ordering(p.Get("order"))
	if len(ordering) > 0 {
		s = s.OrderBy(ordering...)
	}

	// Is a limit specified?
	if t.Limit != 0 {
		s = s.Limit(t.Limit)
		meta["limit"] = t.Limit

		// Check for an offset
		// TODO offset should not be hardcoded
		offset := ParseIntOrDefault(p.Get("offset"), 0)
		if offset != 0 {
			s = s.Offset(offset)
			meta["offset"] = offset
		}
	}
	return s, meta
}

// Ordering parses the order parameter. The default order will be the primary
// key(s) in ascending order.
func (t *Resource) Ordering(raw string) (ordering []aspect.Orderable) {
	// TODO Hopefully fields don't have commas
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		var inverted bool
		if part != "" && part[0] == '-' {
			part = part[1:]
			inverted = true
		}

		// Confirm that the column is sortable
		if t.sortable.Has(part) {
			if inverted {
				ordering = append(ordering, t.table.C[part].Desc())
			} else {
				ordering = append(ordering, t.table.C[part].Asc())
			}
		}
	}

	// If no ordering was determined from the given parameter, use default
	if len(ordering) == 0 {
		// TODO Iterate through the ColumnElems?
		for _, column := range t.pk {
			ordering = append(ordering, t.table.C[column].Asc())
		}
	}
	return
}

// ResourceFromTable builds an argo resource from a aspect table element.
func ResourceFromTable(t *aspect.TableElem) (r *Resource) {
	r = &Resource{}
	// Set the table properties
	r.Name = t.Name
	r.table = t
	r.pk = t.PrimaryKey()


	// Every column should be sortable by default
	columns := t.Columns()
	columnNames := make([]string, len(columns))
	for i, column := range columns {
		columnNames[i] = column.Name()
	}
	r.sortable = columnNames

	// Set defaults
	r.Limit = 100
	r.Radius = 200

	// TODO Check the actual columns
	if r.sortable.Has("location") && r.sortable.Has("latitude") && r.sortable.Has("longitude") {
		r.locationFields = []LocationField{
			{
				latitude: "latitude",
				longitude: "longitude",
				location: "location",
			},
		}
	}

	// TODO Remove the location columns so they aren't parsed as floats
	return r
}
