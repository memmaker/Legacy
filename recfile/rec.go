package recfile

import (
    "bufio"
    "fmt"
    "io"
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

func (f Field) AsInt() int {
    if value, err := strconv.Atoi(f.Value); err == nil {
        return value
    }
    return 0
}

func (f Field) AsInt32() int32 {
    value, _ := strconv.ParseInt(f.Value, 10, 32)
    return int32(value)
}

func (f Field) AsBool() bool {
    return f.Value == "true"
}

func (f Field) AsFloat() float64 {
    if value, err := strconv.ParseFloat(f.Value, 64); err == nil {
        return value
    }
    return 0.0
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

func (r Record) ToValueList() []string {
    result := make([]string, len(r))
    for i, field := range r {
        result[i] = field.Value
    }
    return result
}

type RecReader struct {
    records           map[string][]Record
    currentRecord     []Field
    currentField      Field
    linePart          string
    currentRecordType string
}

func NewReader() *RecReader {
    return &RecReader{
        records:           make(map[string][]Record),
        currentRecord:     make([]Field, 0),
        currentField:      Field{},
        linePart:          "",
        currentRecordType: "default",
    }
}
func (r *RecReader) ReadLine(line string) {
    //scanner := bufio.NewScanner(file)
    fieldNamePattern := regexp.MustCompile(`^([a-zA-Z%][a-zA-Z0-9_]*):[\t ]?`)
    plusPrefixPattern := regexp.MustCompile(`^\+\s?`)
    // eg. %rec: Article
    recordTypeRegex := regexp.MustCompile(`^%rec: ([a-zA-Z][a-zA-Z0-9_]*)`)
    line = r.linePart + line
    r.linePart = ""

    if strings.HasPrefix(line, "#") {
        return
    }
    if plusPrefixPattern.MatchString(line) {
        line = plusPrefixPattern.ReplaceAllString(line, "\n")
    }
    if matches := recordTypeRegex.FindStringSubmatch(line); matches != nil {
        r.tryCommitCurrentField()
        r.tryCommitCurrentRecord()
        r.currentRecord = make([]Field, 0)
        r.currentField = Field{}
        r.currentRecordType = matches[1]
        r.records[r.currentRecordType] = make([]Record, 0)
        return
    }

    if strings.HasSuffix(line, "\\") {
        r.linePart = line[:len(line)-1]
        return
    }

    if fieldNamePattern.MatchString(line) {
        r.tryCommitCurrentField()
        matches := fieldNamePattern.FindStringSubmatch(line)
        r.currentField = Field{
            Name:  matches[1],
            Value: strings.TrimSpace(line[len(matches[0]):]),
        }
    } else if line == "" {
        r.tryCommitCurrentField()
        r.currentField = Field{}
        r.tryCommitCurrentRecord()
        r.currentRecord = make([]Field, 0)
    } else {
        r.currentField.Value = strings.TrimSpace(r.currentField.Value + line)
    }
}

func (r *RecReader) tryCommitCurrentRecord() {
    if len(r.currentRecord) > 0 {
        r.records[r.currentRecordType] = append(r.records[r.currentRecordType], r.currentRecord)
    }
}

func (r *RecReader) tryCommitCurrentField() {
    if !r.currentField.IsEmpty() {
        r.currentRecord = append(r.currentRecord, r.currentField)
    }
}

func (r *RecReader) End() map[string][]Record {
    r.tryCommitCurrentField()
    r.currentField = Field{}
    r.tryCommitCurrentRecord()
    return r.records
}

func (r *RecReader) ReadLines(data []string) map[string][]Record {
    for _, line := range data {
        r.ReadLine(line)
    }
    return r.End()
}
func defaultOnly(records map[string][]Record) []Record {
    return records["default"]
}
func Read(file io.Reader) []Record {
    return defaultOnly(ReadMulti(file))
}
func ReadMulti(input io.Reader) map[string][]Record {
    scanner := bufio.NewScanner(input)
    reader := NewReader()
    for scanner.Scan() {
        reader.ReadLine(scanner.Text())
    }
    return reader.End()
}
func Write(file io.StringWriter, records []Record) error {
    return WriteMulti(file, map[string][]Record{"default": records})
}

func WriteMulti(file io.StringWriter, recordsInCategories map[string][]Record) error {
    sanitizeFieldname := func(s string) string {
        saneFieldname := strings.ReplaceAll(s, " ", "_")
        if saneFieldname != s {
            println(fmt.Sprintf("WARNING - Sanitizing fieldname: '%s' -> '%s'", s, saneFieldname))
        }
        return saneFieldname
    }
    for recordCategory, records := range recordsInCategories {
        _, catErr := file.WriteString(fmt.Sprintf("%%rec: %s\n", recordCategory))
        if catErr != nil {
            return catErr
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
    }
    return nil
}
