package util

import "strings"

// PrintTable is a table for printing with fixed column width and padding
type PrintTable struct {
	Columns []PrintColumn
	NumRows int
}

// PrintColumn is a column in PrintTable
type PrintColumn struct {
	Items []string
	Width int
}

//------------//
// PrintTable //
//------------//

// AddRow adds a row of headers/values to the table
func (t *PrintTable) AddRow(items ...string) *PrintTable {
	if t.Columns == nil {
		t.Columns = make([]PrintColumn, len(items))
		for i := range t.Columns {
			c := &t.Columns[i]
			c.Items = []string{}
			c.Width = 0
		}
	} else if len(items) != len(t.Columns) {
		panic("PrintTable columns mismatch")
	}
	for i := range t.Columns {
		c := &t.Columns[i]
		s := items[i]
		c.Items = append(c.Items, s)
		c.Width = max(len(s), c.Width)
	}
	t.NumRows++
	return t
}

// RecalcWidth can be called to recalculate the columns width after manual updates are made
func (t *PrintTable) RecalcWidth() *PrintTable {
	for i := range t.Columns {
		t.Columns[i].RecalcWidth()
	}
	return t
}

func (t *PrintTable) String() string {
	res := ""
	for r := 0; r < t.NumRows; r++ {
		for i := range t.Columns {
			c := &t.Columns[i]
			res += c.Pad(c.Items[r])
		}
		res += "\n"
	}
	return res
}

//-------------//
// PrintColumn //
//-------------//

// Pad adds padding the provided string to make it print nicely as a table cell
func (c *PrintColumn) Pad(s string) string {
	width := min(len(s), c.Width)
	pad := strings.Repeat(" ", c.Width-width+3)
	return s[0:width] + pad
}

// RecalcWidth can be called to recalculate the column width after manual updates are made
func (c *PrintColumn) RecalcWidth() {
	c.Width = maxLen(c.Items...)
}

func min(items ...int) int {
	r := items[0]
	for _, x := range items {
		if r > x {
			r = x
		}
	}
	return r
}

func max(items ...int) int {
	r := items[0]
	for _, x := range items {
		if r < x {
			r = x
		}
	}
	return r
}

func maxLen(items ...string) int {
	r := len(items[0])
	for _, x := range items {
		if r < len(x) {
			r = len(x)
		}
	}
	return r
}
