package util

import (
    "fmt"
    "strings"
)

func rightPad(s string, pLen int) string {
    return s + strings.Repeat(" ", pLen-len(s))
}

func rightAlignColumns(cols []string, width int) string {
    for i, col := range cols {
        if len(col) < width {
            cols[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", width-len(col)), col)
        }
    }
    return strings.Join(cols, "")
}

type TableRow struct {
    Label   string
    Columns []string
}

func TableLine(labelWidth, colWidth int, label string, columns ...string) string {
    return fmt.Sprintf("%s%s", rightPad(label, labelWidth+1), rightAlignColumns(columns, colWidth))
}
func TableLayout(tableData []TableRow) []string {
    var maxLabelLen int
    var maxColLen int

    for _, row := range tableData {
        if len(row.Label) > maxLabelLen {
            maxLabelLen = len(row.Label)
        }
        for _, col := range row.Columns {
            if len(col) > maxColLen {
                maxColLen = len(col)
            }
        }
    }

    var result []string
    for _, row := range tableData {
        result = append(result, fmt.Sprintf("%s%s", rightPad(row.Label, maxLabelLen+1), rightAlignColumns(row.Columns, maxColLen)))
    }
    return result
}
