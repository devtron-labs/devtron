package pp

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	indentWidth = 2
)

var (
	// If the length of array or slice is larger than this,
	// the buffer will be shorten as {...}.
	BufferFoldThreshold = 1024
)

func format(object interface{}) string {
	return newPrinter(object).String()
}

func newPrinter(object interface{}) *printer {
	buffer := bytes.NewBufferString("")
	tw := new(tabwriter.Writer)
	tw.Init(buffer, indentWidth, 0, 1, ' ', 0)

	return &printer{
		Buffer:  buffer,
		tw:      tw,
		depth:   0,
		value:   reflect.ValueOf(object),
		visited: map[uintptr]bool{},
	}
}

type printer struct {
	*bytes.Buffer
	tw      *tabwriter.Writer
	depth   int
	value   reflect.Value
	visited map[uintptr]bool
}

func (p *printer) String() string {
	switch p.value.Kind() {
	case reflect.Bool:
		p.colorPrint(p.raw(), currentScheme.Bool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Complex64, reflect.Complex128:
		p.colorPrint(p.raw(), currentScheme.Integer)
	case reflect.Float32, reflect.Float64:
		p.colorPrint(p.raw(), currentScheme.Float)
	case reflect.String:
		p.printString()
	case reflect.Map:
		p.printMap()
	case reflect.Struct:
		p.printStruct()
	case reflect.Array, reflect.Slice:
		p.printSlice()
	case reflect.Chan:
		p.printf("(%s)(%s)", p.typeString(), p.pointerAddr())
	case reflect.Interface:
		p.printInterface()
	case reflect.Ptr:
		p.printPtr()
	case reflect.Func:
		p.printf("%s {...}", p.typeString())
	case reflect.UnsafePointer:
		p.printf("%s(%s)", p.typeString(), p.pointerAddr())
	case reflect.Invalid:
		p.print(p.nil())
	default:
		p.print(p.raw())
	}

	p.tw.Flush()
	return p.Buffer.String()
}

func (p *printer) print(text string) {
	fmt.Fprint(p.tw, text)
}

func (p *printer) printf(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	p.print(text)
}

func (p *printer) println(text string) {
	p.print(text + "\n")
}

func (p *printer) indentPrint(text string) {
	p.print(p.indent() + text)
}

func (p *printer) indentPrintf(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	p.indentPrint(text)
}

func (p *printer) colorPrint(text string, color uint16) {
	p.print(colorize(text, color))
}

func (p *printer) printString() {
	quoted := strconv.Quote(p.value.String())
	quoted = quoted[1 : len(quoted)-1]

	p.colorPrint(`"`, currentScheme.StringQuotation)
	for len(quoted) > 0 {
		pos := strings.IndexByte(quoted, '\\')
		if pos == -1 {
			p.colorPrint(quoted, currentScheme.String)
			break
		}
		if pos != 0 {
			p.colorPrint(quoted[0:pos], currentScheme.String)
		}

		n := 1
		switch quoted[pos+1] {
		case 'x': // "\x00"
			n = 3
		case 'u': // "\u0000"
			n = 5
		case 'U': // "\U00000000"
			n = 9
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // "\000"
			n = 3
		}
		p.colorPrint(quoted[pos:pos+n+1], currentScheme.EscapedChar)
		quoted = quoted[pos+n+1:]
	}
	p.colorPrint(`"`, currentScheme.StringQuotation)
}

func (p *printer) printMap() {
	if p.value.Len() == 0 {
		p.printf("%s{}", p.typeString())
		return
	}

	if p.visited[p.value.Pointer()] {
		p.printf("%s{...}", p.typeString())
		return
	}
	p.visited[p.value.Pointer()] = true

	p.println("{")
	p.indented(func() {
		keys := p.value.MapKeys()
		for i := 0; i < p.value.Len(); i++ {
			value := p.value.MapIndex(keys[i])
			p.indentPrintf("%s:\t%s,\n", p.format(keys[i]), p.format(value))
		}
	})
	p.indentPrint("}")
}

func (p *printer) printStruct() {
	if p.value.Type().String() == "time.Time" {
		p.printTime()
		return
	}

	if p.value.NumField() == 0 {
		p.print(p.typeString() + "{}")
		return
	}

	p.println(p.typeString() + "{")
	p.indented(func() {
		for i := 0; i < p.value.NumField(); i++ {
			field := colorize(p.value.Type().Field(i).Name, currentScheme.FieldName)
			value := p.value.Field(i)
			p.indentPrintf("%s:\t%s,\n", field, p.format(value))
		}
	})
	p.indentPrint("}")
}

func (p *printer) printTime() {
	if !p.value.CanInterface() {
		p.printf("(unexported time.Time)")
		return
	}

	tm := p.value.Interface().(time.Time)
	p.printf(
		"%s-%s-%s %s:%s:%s %s",
		colorize(strconv.Itoa(tm.Year()), currentScheme.Time),
		colorize(fmt.Sprintf("%02d", tm.Month()), currentScheme.Time),
		colorize(fmt.Sprintf("%02d", tm.Day()), currentScheme.Time),
		colorize(fmt.Sprintf("%02d", tm.Hour()), currentScheme.Time),
		colorize(fmt.Sprintf("%02d", tm.Minute()), currentScheme.Time),
		colorize(fmt.Sprintf("%02d", tm.Second()), currentScheme.Time),
		colorize(tm.Location().String(), currentScheme.Time),
	)
}

func (p *printer) printSlice() {
	if p.value.Len() == 0 {
		p.printf("%s{}", p.typeString())
		return
	}

	if p.value.Kind() == reflect.Slice {
		if p.visited[p.value.Pointer()] {
			// Stop travarsing cyclic reference
			p.printf("%s{...}", p.typeString())
			return
		}
		p.visited[p.value.Pointer()] = true
	}

	// Fold a large buffer
	if p.value.Len() > BufferFoldThreshold {
		p.printf("%s{...}", p.typeString())
		return
	}

	p.println(p.typeString() + "{")
	p.indented(func() {
		groupsize := 0
		switch p.value.Type().Elem().Kind() {
		case reflect.Uint8:
			groupsize = 16
		case reflect.Uint16:
			groupsize = 8
		case reflect.Uint32:
			groupsize = 8
		case reflect.Uint64:
			groupsize = 4
		}

		if groupsize > 0 {
			for i := 0; i < p.value.Len(); i++ {
				// indent for new group
				if i%groupsize == 0 {
					p.print(p.indent())
				}
				// slice element
				p.printf("%s,", p.format(p.value.Index(i)))
				// space or newline
				if (i+1)%groupsize == 0 || i+1 == p.value.Len() {
					p.print("\n")
				} else {
					p.print(" ")
				}
			}
		} else {
			for i := 0; i < p.value.Len(); i++ {
				p.indentPrintf("%s,\n", p.format(p.value.Index(i)))
			}
		}
	})
	p.indentPrint("}")
}

func (p *printer) printInterface() {
	e := p.value.Elem()
	if e.Kind() == reflect.Invalid {
		p.print(p.nil())
	} else if e.IsValid() {
		p.print(p.format(e))
	} else {
		p.printf("%s(%s)", p.typeString(), p.nil())
	}
}

func (p *printer) printPtr() {
	if p.visited[p.value.Pointer()] {
		p.printf("&%s{...}", p.elemTypeString())
		return
	}
	if p.value.Pointer() != 0 {
		p.visited[p.value.Pointer()] = true
	}

	if p.value.Elem().IsValid() {
		p.printf("&%s", p.format(p.value.Elem()))
	} else {
		p.printf("(%s)(%s)", p.typeString(), p.nil())
	}
}

func (p *printer) pointerAddr() string {
	return colorize(fmt.Sprintf("%#v", p.value.Pointer()), currentScheme.PointerAdress)
}

func (p *printer) typeString() string {
	return p.colorizeType(p.value.Type().String())
}

func (p *printer) elemTypeString() string {
	return p.colorizeType(p.value.Elem().Type().String())
}

func (p *printer) colorizeType(t string) string {
	prefix := ""

	if p.matchRegexp(t, `^\[\].+$`) {
		prefix = "[]"
		t = t[2:]
	}

	if p.matchRegexp(t, `^\[\d+\].+$`) {
		num := regexp.MustCompile(`\d+`).FindString(t)
		prefix = fmt.Sprintf("[%s]", colorize(num, currentScheme.ObjectLength))
		t = t[2+len(num):]
	}

	if p.matchRegexp(t, `^[^\.]+\.[^\.]+$`) {
		ts := strings.Split(t, ".")
		t = fmt.Sprintf("%s.%s", ts[0], colorize(ts[1], currentScheme.StructName))
	} else {
		t = colorize(t, currentScheme.StructName)
	}
	return prefix + t
}

func (p *printer) matchRegexp(text, exp string) bool {
	return regexp.MustCompile(exp).MatchString(text)
}

func (p *printer) indented(proc func()) {
	p.depth++
	proc()
	p.depth--
}

func (p *printer) raw() string {
	// Some value causes panic when Interface() is called.
	switch p.value.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%#v", p.value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%#v", p.value.Int())
	case reflect.Uint, reflect.Uintptr:
		return fmt.Sprintf("%#v", p.value.Uint())
	case reflect.Uint8:
		return fmt.Sprintf("0x%02x", p.value.Uint())
	case reflect.Uint16:
		return fmt.Sprintf("0x%04x", p.value.Uint())
	case reflect.Uint32:
		return fmt.Sprintf("0x%08x", p.value.Uint())
	case reflect.Uint64:
		return fmt.Sprintf("0x%016x", p.value.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", p.value.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%#v", p.value.Complex())
	default:
		return fmt.Sprintf("%#v", p.value.Interface())
	}
}

func (p *printer) nil() string {
	return colorize("nil", currentScheme.Nil)
}

func (p *printer) format(object interface{}) string {
	pp := newPrinter(object)
	pp.depth = p.depth
	pp.visited = p.visited
	if value, ok := object.(reflect.Value); ok {
		pp.value = value
	}
	return pp.String()
}

func (p *printer) indent() string {
	return strings.Repeat("\t", p.depth)
}
