// Package table provides a convenient way to generate tabular output of any
// data, primarily useful for CLI tools.
//
// Columns are left-aligned and padded to accomodate the largest cell in that
// column.
//
// Source: https://github.com/rodaine/table
//
//	table.DefaultHeaderFormatter = func(format string, vals ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(format, vals...))
//	}
//
//	tbl := table.New("ID", "Name", "Cost ($)")
//
//	for _, widget := range Widgets {
//	  tbl.AddRow(widget.ID, widget.Name, widget.Cost)
//	}
//
//	tbl.Print()
//
//	// Output:
//	// ID  NAME      COST ($)
//	// 1   Foobar    1.23
//	// 2   Fizzbuzz  4.56
//	// 3   Gizmo     78.90
package table

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// These are the default properties for all Tables created from this package
// and can be modified.
var (
	// DefaultPadding specifies the number of spaces between columns in a table.
	DefaultPadding = 2

	// DefaultWriter specifies the output io.Writer for the Table.Print method.
	DefaultWriter io.Writer = os.Stdout

	// DefaultHeaderFormatter specifies the default Formatter for the table header.
	DefaultHeaderFormatter Formatter

	// DefaultFirstColumnFormatter specifies the default Formatter for the first column cells.
	DefaultFirstColumnFormatter Formatter

	// DefaultWidthFunc specifies the default WidthFunc for calculating column widths
	DefaultWidthFunc WidthFunc = utf8.RuneCountInString
)

// Formatter functions expose a fmt.Sprintf signature that can be used to modify
// the display of the text in either the header or first column of a Table.
// The formatter should not change the width of original text as printed since
// column widths are calculated pre-formatting (though this issue can be mitigated
// with increased padding).
//
//	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(format, vals...))
//	})
//
// A good use case for formatters is to use ANSI escape codes to color the cells
// for a nicer interface. The package color (https://github.com/fatih/color) makes
// it easy to generate these automatically: http://godoc.org/github.com/fatih/color#Color.SprintfFunc
type Formatter func(string, ...interface{}) string

// A WidthFunc calculates the width of a string. By default, the number of runes
// is used but this may not be appropriate for certain character sets. The
// package runewidth (https://github.com/mattn/go-runewidth) could be used to
// accomodate multi-cell characters (such as emoji or CJK characters).
type WidthFunc func(string) int

// Table describes the interface for building up a tabular representation of data.
// It exposes fluent/chainable methods for convenient table building.
//
// WithHeaderFormatter and WithFirstColumnFormatter sets the Formatter for the
// header and first column, respectively. If nil is passed in (the default), no
// formatting will be applied.
//
//	New("foo", "bar").WithFirstColumnFormatter(func(f string, v ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(f, v...))
//	})
//
// WithPadding specifies the minimum padding between cells in a row and defaults
// to DefaultPadding. Padding values less than or equal to zero apply no extra
// padding between the columns.
//
//	New("foo", "bar").WithPadding(3)
//
// WithWriter modifies the writer which Print outputs to, defaulting to DefaultWriter
// when instantiated. If nil is passed, os.Stdout will be used.
//
//	New("foo", "bar").WithWriter(os.Stderr)
//
// WithWidthFunc sets the function used to calculate the width of the string in
// a column. By default, the number of utf8 runes in the string is used.
//
// AddRow adds another row of data to the table. Any values can be passed in and
// will be output as its string representation as described in the fmt standard
// package. Rows can have less cells than the total number of columns in the table;
// subsequent cells will be rendered empty. Rows with more cells than the total
// number of columns will be truncated. References to the data are not held, so
// the passed in values can be modified without affecting the table's output.
//
//	New("foo", "bar").AddRow("fizz", "buzz").AddRow(time.Now()).AddRow(1, 2, 3).Print()
//	// Output:
//	// foo                              bar
//	// fizz                             buzz
//	// 2006-01-02 15:04:05.0 -0700 MST
//	// 1                                2
//
// Print writes the string representation of the table to the provided writer.
// Print can be called multiple times, even after subsequent mutations of the
// provided data. The output is always preceded and followed by a new line.
type Table interface {
	WithHeaderFormatter(f Formatter) Table
	WithFirstColumnFormatter(f Formatter) Table
	WithPadding(p int) Table
	WithWriter(w io.Writer) Table
	WithWidthFunc(f WidthFunc) Table
	WithHeaderSeparatorRow(r rune) Table

	AddRow(vals ...interface{}) Table
	SetRows(rows [][]string) Table
	Print()
}

// New creates a Table instance with the specified header(s) provided. The number
// of columns is fixed at this point to len(columnHeaders) and the defined defaults
// are set on the instance.
func New(columnHeaders ...interface{}) Table {
	t := table{header: make([]string, len(columnHeaders))}

	t.WithPadding(DefaultPadding)
	t.WithWriter(DefaultWriter)
	t.WithHeaderFormatter(DefaultHeaderFormatter)
	t.WithFirstColumnFormatter(DefaultFirstColumnFormatter)
	t.WithWidthFunc(DefaultWidthFunc)

	for i, col := range columnHeaders {
		t.header[i] = fmt.Sprint(col)
	}

	return &t
}

type table struct {
	FirstColumnFormatter Formatter
	HeaderFormatter      Formatter
	Padding              int
	Writer               io.Writer
	Width                WidthFunc
	HeaderSeparatorRune  rune

	header []string
	rows   [][]string
	widths []int
}

func (t *table) WithHeaderFormatter(f Formatter) Table {
	t.HeaderFormatter = f
	return t
}

func (t *table) WithHeaderSeparatorRow(r rune) Table {
	t.HeaderSeparatorRune = r
	return t
}

func (t *table) WithFirstColumnFormatter(f Formatter) Table {
	t.FirstColumnFormatter = f
	return t
}

func (t *table) WithPadding(p int) Table {
	if p < 0 {
		p = 0
	}

	t.Padding = p
	return t
}

func (t *table) WithWriter(w io.Writer) Table {
	if w == nil {
		w = os.Stdout
	}

	t.Writer = w
	return t
}

func (t *table) WithWidthFunc(f WidthFunc) Table {
	t.Width = f
	return t
}

func (t *table) AddRow(vals ...interface{}) Table {
	maxNumNewlines := 0
	for _, val := range vals {
		maxNumNewlines = max(strings.Count(fmt.Sprint(val), "\n"), maxNumNewlines)
	}
	for i := 0; i <= maxNumNewlines; i++ {
		row := make([]string, len(t.header))
		for j, val := range vals {
			if j >= len(t.header) {
				break
			}
			v := strings.Split(fmt.Sprint(val), "\n")
			row[j] = safeOffset(v, i)
		}
		t.rows = append(t.rows, row)
	}

	return t
}

func (t *table) SetRows(rows [][]string) Table {
	t.rows = [][]string{}
	headerLength := len(t.header)

	for _, row := range rows {
		if len(row) > headerLength {
			t.rows = append(t.rows, row[:headerLength])
		} else {
			t.rows = append(t.rows, row)
		}
	}

	return t
}

func (t *table) Print() {
	format := strings.Repeat("%s", len(t.header)) + "\n"
	t.calculateWidths()

	t.printHeader(format)
	if t.HeaderSeparatorRune != 0 {
		t.printHeaderSeparator(format)
	}
	for _, row := range t.rows {
		t.printRow(format, row)
	}
}

func (t *table) printHeaderSeparator(format string) {
	separators := make([]string, len(t.header))

	// The separator could be any unicode char. Since some chars take up more
	// than one cell in a monospace context, we can get a number higher than 1
	// here. Am example would be this emoji 🤣.
	separatorCellWidth := t.Width(string([]rune{t.HeaderSeparatorRune}))
	for index, headerName := range t.header {
		headerCellWidth := t.Width(headerName)
		// Note that this might not be evenly divisble. In this case we'll get a
		// separator that is at least 1 cell shorter than the header. This was
		// an intentional design decision in order to prevent widening the cell
		// or overstepping the column bounds.
		repeatCharTimes := headerCellWidth / separatorCellWidth
		separator := make([]rune, repeatCharTimes)
		for i := 0; i < repeatCharTimes; i++ {
			separator[i] = t.HeaderSeparatorRune
		}
		separators[index] = string(separator)
	}

	vals := t.applyWidths(separators, t.widths)
	if t.HeaderFormatter != nil {
		txt := t.HeaderFormatter(format, vals...)
		fmt.Fprint(t.Writer, txt)
	} else {
		fmt.Fprintf(t.Writer, format, vals...)
	}
}

func (t *table) printHeader(format string) {
	vals := t.applyWidths(t.header, t.widths)
	if t.HeaderFormatter != nil {
		txt := t.HeaderFormatter(format, vals...)
		fmt.Fprint(t.Writer, txt)
	} else {
		fmt.Fprintf(t.Writer, format, vals...)
	}
}

func (t *table) printRow(format string, row []string) {
	vals := t.applyWidths(row, t.widths)

	if t.FirstColumnFormatter != nil {
		vals[0] = t.FirstColumnFormatter("%s", vals[0])
	}

	fmt.Fprintf(t.Writer, format, vals...)
}

func (t *table) calculateWidths() {
	t.widths = make([]int, len(t.header))
	for _, row := range t.rows {
		for i, v := range row {
			if w := t.Width(v) + t.Padding; w > t.widths[i] {
				t.widths[i] = w
			}
		}
	}

	for i, v := range t.header {
		if w := t.Width(v) + t.Padding; w > t.widths[i] {
			t.widths[i] = w
		}
	}
}

func (t *table) applyWidths(row []string, widths []int) []interface{} {
	out := make([]interface{}, len(row))
	for i, s := range row {
		out[i] = s + t.lenOffset(s, widths[i])
	}
	return out
}

func (t *table) lenOffset(s string, w int) string {
	l := w - t.Width(s)
	if l <= 0 {
		return ""
	}
	return strings.Repeat(" ", l)
}

func max(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func safeOffset(sarr []string, idx int) string {
	if idx >= len(sarr) {
		return ""
	}
	return sarr[idx]
}
