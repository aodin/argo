package argo

import (
	"fmt"
	"strings"
	"unicode"

	sql "github.com/aodin/aspect"
)

// TableElem wraps the given SQL table. The sql.TableElem is not directly
// added to the resource in case we want to provide options / modifiers
// before altering the ResourceSQL (see the Resource() constructor)
type TableElem struct {
	Name    string
	table   *sql.TableElem
	selects Columns
}

// Exclude removes the given field names from the selects
func (elem TableElem) Exclude(names ...string) TableElem {
	for _, name := range names {
		// Remove the column name from the list of selected columns
		if err := elem.selects.Remove(name); err != nil {
			panic(fmt.Sprintf(
				"argo: the column %s cannot be excluded from the table - it either does not exist or has already been excluded",
				name,
			))
		}
	}
	return elem
}

func slugify(name string) string {
	return strings.Map(func(char rune) rune {
		switch {
		case unicode.IsDigit(char):
			return char
		case unicode.IsLetter(char):
			return unicode.ToLower(char)
		case unicode.IsPrint(char):
			return '-'
		default:
			return -1
		}
	}, name)
}

// Slugify makes the table name URL friendly, allowing only lowercase
// characters and dashes.
func (elem TableElem) Slugify() TableElem {
	elem.Name = slugify(elem.Name)
	return elem
}

func FromTable(table *sql.TableElem) (elem TableElem) {
	if table == nil {
		panic("argo: a table cannot be nil")
	}
	if len(table.PrimaryKey()) == 0 {
		panic("argo: tables must have a primary key")
	}
	elem.Name = table.Name
	elem.table = table
	elem.selects = ColumnSet(table.Columns()...)
	return
}
