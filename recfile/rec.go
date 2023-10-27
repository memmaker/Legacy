package recfile

import (
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// https://en.wikipedia.org/wiki/Recfiles
// https://www.gnu.org/software/recutils/manual/The-Rec-Format.html#The-Rec-Format
type Field struct {
    Name  string
    Value string
}
type Record []Field

func (f Field) String() string {
    return fmt.Sprintf("%s: %s", f.Name, f.Value)
}

func (f Field) IsEmpty() bool {
    return f.Name == "" && f.Value == ""
}

func (f Field) UnEscapedValue() string {
    return regexp.MustCompile(`\n\+\s`).ReplaceAllString(f.Value, "\n")
}

func (f Field) EscapedValue() string {
    return strings.ReplaceAll(f.Value, "\n", "\n+ ")
}

type DataMap map[string]string

func (d DataMap) GetInt(key string) (int, error) {
    if value, ok := d[key]; ok {
        return strconv.Atoi(value)
    }
    return -1, fmt.Errorf("no value")
}

func (r Record) ToMap() DataMap {
    m := make(map[string]string, len(r))
    for _, field := range r {
        m[field.Name] = field.Value
    }
    return m
}

type RecReader struct {
    records       []Record
    currentRecord []Field
    currentField  Field
    linePart      string
}

func NewReader() *RecReader {
    return &RecReader{
        records:       make([]Record, 0),
        currentRecord: make([]Field, 0),
        currentField:  Field{},
        linePart:      "",
    }
}
func (r *RecReader) ReadLine(line string) {
    //scanner := bufio.NewScanner(file)
    fieldNamePattern := regexp.MustCompile(`^([a-zA-Z%][a-zA-Z0-9_]*):[\t ]?`)
    tryCommitCurrentField := func() {
        if !r.currentField.IsEmpty() {
            r.currentRecord = append(r.currentRecord, r.currentField)
        }
    }
    tryCommitCurrentRecord := func() {
        if len(r.currentRecord) > 0 {
            r.records = append(r.records, r.currentRecord)
        }
    }
    line = r.linePart + line
    r.linePart = ""

    if strings.HasPrefix(line, "#") {
        return
    }
    if strings.HasSuffix(line, "\\") {
        r.linePart = line[:len(line)-1]
        return
    }

    if fieldNamePattern.MatchString(line) {
        tryCommitCurrentField()
        matches := fieldNamePattern.FindStringSubmatch(line)
        r.currentField = Field{
            Name:  matches[1],
            Value: strings.TrimSpace(line[len(matches[0]):]),
        }
    } else if line == "" {
        tryCommitCurrentField()
        r.currentField = Field{}
        tryCommitCurrentRecord()
        r.currentRecord = make([]Field, 0)
    } else {
        r.currentField.Value = strings.TrimSpace(r.currentField.Value + line)
    }
}

func (r *RecReader) End() []Record {
    tryCommitCurrentField := func() {
        if !r.currentField.IsEmpty() {
            r.currentRecord = append(r.currentRecord, r.currentField)
        }
    }
    tryCommitCurrentRecord := func() {
        if len(r.currentRecord) > 0 {
            r.records = append(r.records, r.currentRecord)
        }
    }

    tryCommitCurrentField()
    r.currentField = Field{}
    tryCommitCurrentRecord()
    return r.records
}

func (r *RecReader) ReadLines(data []string) []Record {
    for _, line := range data {
        r.ReadLine(line)
    }
    return r.End()
}

func Write(file *os.File, records []Record) error {
    sanitizeFieldname := func(s string) string {
        saneFieldname := strings.ReplaceAll(s, " ", "_")
        if saneFieldname != s {
            println(fmt.Sprintf("WARNING - Sanitizing fieldname: '%s' -> '%s'", s, saneFieldname))
        }
        return saneFieldname
    }
    for _, record := range records {
        for _, field := range record {
            _, err := file.WriteString(sanitizeFieldname(field.Name) + ": " + field.EscapedValue() + "\n")
            if err != nil {
                return err
            }
        }
        _, err := file.WriteString("\n")
        if err != nil {
            return err
        }
    }
    return nil
}
