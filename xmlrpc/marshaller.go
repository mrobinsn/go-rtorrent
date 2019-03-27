package xmlrpc

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/// ISO8601 is not very much restrictive, so many combinations exist
const (
	// FullXMLRpcTime is the format of a full XML-RPC time
	FullXMLRpcTime = "2006-01-02T15:04:05-07:00"
	// LocalXMLRpcTime is the XML-RPC time without timezone
	LocalXMLRpcTime = "2006-01-02T15:04:05"
	// DenseXMLRpcTime is a dense-formatted local time
	DenseXMLRpcTime = "20060102T15:04:05"
	// DummyXMLRpcTime is seen in the wild
	DummyXMLRpcTime = "20060102T15:04:05-0700"
)

// ErrUnsupported is the error of "Unsupported type"
var ErrUnsupported = errors.New("Unsupported type")

// Fault is the struct for the fault response
type Fault struct {
	Code    int
	Message string
}

func (f Fault) String() string {
	return fmt.Sprintf("%d: %s", f.Code, f.Message)
}
func (f Fault) Error() string {
	return f.String()
}

// WriteXML writes the XML representation of the fault into the Writer
func (f Fault) WriteXML(w io.Writer) (int, error) {
	return fmt.Fprintf(w, `<fault><value><struct>
			<member><name>faultCode</name><value><int>%d</int></value></member>
			<member><name>faultString</name><value><string>%s</string></value></member>
			</struct></value></fault>`, f.Code, xmlEscape(f.Message))
}

var xmlSpecial = map[byte]string{
	'<':  "&lt;",
	'>':  "&gt;",
	'"':  "&quot;",
	'\'': "&apos;",
	'&':  "&amp;",
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if s, ok := xmlSpecial[c]; ok {
			b.WriteString(s)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

type valueNode struct {
	Type string `xml:",attr"`
	Body string `xml:",chardata"`
}

type state struct {
	p         *xml.Decoder
	level     int
	remainder *interface{}
	last      *xml.Token
}

func newParser(p *xml.Decoder) *state {
	return &state{p, 0, nil, nil}
}

const (
	tokStart = iota
	tokText
	tokStop
)

var (
	errNotStartElement = errors.New("not start element")
	errNameMismatch    = errors.New("not the required token")
	errNotEndElement   = errors.New("not end element")
)

func (st *state) parseValue() (nv interface{}, e error) {
	var se xml.StartElement
	if se, e = st.getStart(""); e != nil {
		if ErrEq(e, errNotStartElement) {
			e = nil
		}
		return
	}

	var vn valueNode
	switch se.Name.Local {
	case "value":
		if nv, e = st.parseValue(); e == nil {
			e = st.checkLast("value")
		}
		return
	case "boolean", "string", "int", "i1", "i2", "i4", "i8", "double", "dateTime.iso8601", "base64": //simple
		st.last = nil
		if e = st.p.DecodeElement(&vn, &se); e != nil {
			return
		}

		switch se.Name.Local {
		case "boolean":
			nv, e = strconv.ParseBool(vn.Body)
		case "string":
			nv = vn.Body
		case "int", "i1", "i2", "i4":
			var i64 int64
			i64, e = strconv.ParseInt(vn.Body, 10, 32)
			nv = int(i64)
		case "i8":
			var i64 int64
			i64, e = strconv.ParseInt(vn.Body, 10, 64)
			nv = int(i64)
		case "double":
			nv, e = strconv.ParseFloat(vn.Body, 64)
		case "dateTime.iso8601":
			for _, format := range []string{FullXMLRpcTime, LocalXMLRpcTime, DenseXMLRpcTime, DummyXMLRpcTime} {
				nv, e = time.Parse(format, vn.Body)
				if e == nil {
					break
				}
			}
		case "base64":
			nv, e = base64.StdEncoding.DecodeString(vn.Body)
		}
		return

	case "struct":
		var name string
		values := make(map[string]interface{}, 4)
		nv = values
		for {
			if se, e = st.getStart("member"); e != nil {
				if ErrEq(e, errNotStartElement) {
					e = st.checkLast("struct")
					break
				}
				return
			}
			if name, e = st.getText("name"); e != nil {
				return
			}
			if se, e = st.getStart("value"); e != nil {
				return
			}
			if values[name], e = st.parseValue(); e != nil {
				return
			}
			if e = st.checkLast("value"); e != nil {
				return
			}
			if e = st.checkLast("member"); e != nil {
				return
			}
		}
		return

	case "array":
		values := make([]interface{}, 0, 4)
		var val interface{}
		if _, e = st.getStart("data"); e != nil {
			return
		}
		for {
			if se, e = st.getStart("value"); e != nil {
				if ErrEq(e, errNotStartElement) {
					e = nil //st.checkLast("data")
					break
				}
				return
			}
			if val, e = st.parseValue(); e != nil {
				return
			}
			values = append(values, val)
			if e = st.checkLast("value"); e != nil {
				return
			}
		}
		if e = st.checkLast("data"); e == nil {
			e = st.checkLast("array")
		}
		nv = values
		return
	default:
		e = fmt.Errorf("cannot parse unknown tag %s", se)
	}
	return
}

func (st *state) token(typ int, name string) (t xml.Token, body string, e error) {
	// var ok bool
	if st.last != nil {
		t = *st.last
		st.last = nil
	}
Reading:
	for {
		if t != nil {
			switch t.(type) {
			case xml.StartElement:
				se := t.(xml.StartElement)
				if se.Name.Local != "" {
					break Reading
				}
			case xml.EndElement:
				ee := t.(xml.EndElement)
				if ee.Name.Local != "" {
					break Reading
				}
			default:
			}
		}
		if t, e = st.p.Token(); e != nil {
			return
		}
		if t == nil {
			e = errors.New("nil token")
			return
		}
	}
	switch typ {
	case tokStart, tokText:
		se, ok := t.(xml.StartElement)
		if !ok {
			st.last = &t
			e = Errorf2(errNotStartElement, "required startelement(%s), found %s %T", name, t, t)
			return
		}
		switch typ {
		case tokStart:
			if name != "" && se.Name.Local != name {
				e = Errorf2(errNameMismatch, "required <%s>, found <%s>", name, se.Name.Local)
				return
			}
		default:
			var vn valueNode
			if e = st.p.DecodeElement(&vn, &se); e != nil {
				return
			}
			body = vn.Body
		}
	default:
		ee, ok := t.(xml.EndElement)
		if !ok {
			st.last = &t
			e = Errorf2(errNotEndElement, "required endelement(%s), found %s %T", name, t, t)
			return
		}
		if name != "" && ee.Name.Local != name {
			e = Errorf2(errNameMismatch, "required </%s>, found </%s>", name, ee.Name.Local)
			return
		}
	}
	return
}

func (st *state) getStart(name string) (se xml.StartElement, e error) {
	var t xml.Token
	t, _, e = st.token(tokStart, name)
	se, _ = t.(xml.StartElement)
	if e != nil {
		return
	}
	se = t.(xml.StartElement)
	return
}

func (st *state) getText(name string) (text string, e error) {
	_, text, e = st.token(tokText, name)
	return
}

func (st *state) checkLast(name string) (e error) {
	_, _, e = st.token(tokStop, name)
	return
}

func toXML(v interface{}, typ bool) (s string) {
	r := reflect.ValueOf(v)
	t := r.Type()
	k := t.Kind()

	if b, ok := v.([]byte); ok {
		return "<base64>" + base64.StdEncoding.EncodeToString(b) + "</base64>"
	}

	switch k {
	case reflect.Invalid:
		panic("Unsupported type")
	case reflect.Bool:
		return fmt.Sprintf("<boolean>%v</boolean>", v)
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typ {
			return fmt.Sprintf("<int>%v</int>", v)
		}
		return fmt.Sprintf("%v", v)
	case reflect.Uintptr:
		panic("Unsupported type")
	case reflect.Float32, reflect.Float64:
		if typ {
			return fmt.Sprintf("<double>%v</double>", v)
		}
		return fmt.Sprintf("%v", v)
	case reflect.Complex64, reflect.Complex128:
		panic("Unsupported type")
	case reflect.Array, reflect.Slice:
		s = "<array><data>"
		for n := 0; n < r.Len(); n++ {
			s += "<value>"
			s += toXML(r.Index(n).Interface(), typ)
			s += "</value>"
		}
		s += "</data></array>"
		return s
	case reflect.Chan:
		panic("Unsupported type")
	case reflect.Func:
		panic("Unsupported type")
	case reflect.Interface:
		return toXML(r.Elem(), typ)
	case reflect.Map:
		s = "<struct>"
		for _, key := range r.MapKeys() {
			s += "<member>"
			s += "<name>" + xmlEscape(key.Interface().(string)) + "</name>"
			s += "<value>" + toXML(r.MapIndex(key).Interface(), typ) + "</value>"
			s += "</member>"
		}
		return s + "</struct>"
	case reflect.Ptr:
		panic("Unsupported type")
	case reflect.String:
		if typ {
			return fmt.Sprintf("<string>%v</string>", xmlEscape(v.(string)))
		}
		return xmlEscape(v.(string))
	case reflect.Struct:
		s = "<struct>"
		for n := 0; n < r.NumField(); n++ {
			s += "<member>"
			s += "<name>" + t.Field(n).Name + "</name>"
			s += "<value>" + toXML(r.FieldByIndex([]int{n}).Interface(), true) + "</value>"
			s += "</member>"
		}
		return s + "</struct>"
	case reflect.UnsafePointer:
		return toXML(r.Elem(), typ)
	}
	return
}

// WriteXML writes v, typed if typ is true, into w Writer
func WriteXML(w io.Writer, v interface{}, typ bool) (err error) {
	var (
		r  reflect.Value
		ok bool
	)
	// go back from reflect.Value, if needed.
	if r, ok = v.(reflect.Value); !ok {
		r = reflect.ValueOf(v)
	} else {
		v = r.Interface()
	}
	if fp, ok := getFault(v); ok {
		_, err = fp.WriteXML(w)
		return
	}
	if b, ok := v.([]byte); ok {
		length := base64.StdEncoding.EncodedLen(len(b))
		dst := make([]byte, length)
		base64.StdEncoding.Encode(dst, b)
		_, err = taggedWrite(w, []byte("base64"), dst)
		return
	}
	if tim, ok := v.(time.Time); ok {
		_, err = taggedWriteString(w, "dateTime.iso8601", tim.Format(FullXMLRpcTime))
		return
	}
	t := r.Type()
	k := t.Kind()

	switch k {
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		return Errorf2(ErrUnsupported, "v=%#v t=%v k=%s", v, t, k)
	case reflect.Bool:
		_, err = fmt.Fprintf(w, "<boolean>%v</boolean>", v)
		return err
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typ {
			_, err = fmt.Fprintf(w, "<int>%v</int>", v)
			return err
		}
		_, err = fmt.Fprintf(w, "%v", v)
		return err
	case reflect.Float32, reflect.Float64:
		if typ {
			_, err = fmt.Fprintf(w, "<double>%v</double>", v)
			return err
		}
		_, err = fmt.Fprintf(w, "%v", v)
		return err
	case reflect.Array, reflect.Slice:
		if _, err = io.WriteString(w, "<array><data>\n"); err != nil {
			return
		}
		n := r.Len()
		for i := 0; i < n; i++ {
			if _, err = io.WriteString(w, "  <value>"); err != nil {
				return
			}
			if err = WriteXML(w, r.Index(i).Interface(), typ); err != nil {
				return
			}
			if _, err = io.WriteString(w, "</value>\n"); err != nil {
				return
			}
		}
		if _, err = io.WriteString(w, "</data></array>\n"); err != nil {
			return
		}
	case reflect.Interface:
		return WriteXML(w, r.Elem(), typ)
	case reflect.Map:
		if _, err = io.WriteString(w, "<struct>\n"); err != nil {
			return
		}
		for _, key := range r.MapKeys() {
			if _, err = io.WriteString(w, "  <member><name>"); err != nil {
				return
			}
			if _, err = io.WriteString(w, xmlEscape(key.Interface().(string))); err != nil {
				return
			}
			if _, err = io.WriteString(w, "</name><value>"); err != nil {
				return
			}
			if err = WriteXML(w, r.MapIndex(key).Interface(), typ); err != nil {
				return
			}
			if _, err = io.WriteString(w, "</value></member>\n"); err != nil {
				return
			}
		}
		_, err = io.WriteString(w, "</struct>")
		return
	case reflect.Ptr:
		return WriteXML(w, reflect.Indirect(r), typ)
	case reflect.String:
		if typ {
			_, err = fmt.Fprintf(w, "<string>%v</string>", xmlEscape(v.(string)))
			return
		}
		_, err = io.WriteString(w, xmlEscape(v.(string)))
		return
	case reflect.Struct:
		if _, err = io.WriteString(w, "<struct>"); err != nil {
			return
		}
		n := r.NumField()
		for i := 0; i < n; i++ {
			c := t.Field(i).Name[:1]
			if strings.ToLower(c) == c { //have to skip unexported fields
				continue
			}
			if _, err = io.WriteString(w, "<member><name>"); err != nil {
				return
			}
			if _, err = io.WriteString(w, xmlEscape(getStructFieldName(t.Field(i)))); err != nil {
				return
			}
			if _, err = io.WriteString(w, "</name><value>"); err != nil {
				return
			}
			if err = WriteXML(w, r.Field(i).Interface(), true); err != nil {
				return
			}
			if _, err = io.WriteString(w, "</value></member>"); err != nil {
				return err
			}
		}
		_, err = io.WriteString(w, "</struct>")
		return
	case reflect.UnsafePointer:
		return WriteXML(w, r.Elem(), typ)
	}
	return
}

func taggedWrite(w io.Writer, tag, inner []byte) (n int, err error) {
	var j int
	for _, elt := range [][]byte{[]byte("<"), tag, []byte(">"), inner,
		[]byte("</"), tag, []byte(">")} {
		j, err = w.Write(elt)
		n += j
		if err != nil {
			return
		}
	}
	return
}
func taggedWriteString(w io.Writer, tag, inner string) (n int, err error) {
	if n, err = io.WriteString(w, "<"+tag+">"); err != nil {
		return
	}
	var j int
	j, err = io.WriteString(w, inner)
	n += j
	if err != nil {
		return
	}
	j, err = io.WriteString(w, "</"+tag+">")
	n += j
	return
}

func getStructFieldName(sf reflect.StructField) string {
	if sf.Tag.Get("xml") == "" {
		return sf.Name
	}
	return sf.Tag.Get("xml")
}

// Marshal marshals the named thing (methodResponse if name == "", otherwise a methodCall)
// into the w Writer
func Marshal(w io.Writer, name string, args ...interface{}) (err error) {
	if name == "" {
		if _, err = io.WriteString(w, "<methodResponse>"); err != nil {
			return
		}
		if len(args) > 0 {
			fp, ok := getFault(args[0])
			if ok {
				_, err = fp.WriteXML(w)
				if err == nil {
					_, err = io.WriteString(w, "\n</methodResponse>")
				}
				return
			}
		}
	} else {
		if _, err = io.WriteString(w, "<methodCall><methodName>"); err != nil {
			return
		}
		if _, err = io.WriteString(w, xmlEscape(name)); err != nil {
			return
		}
		if _, err = io.WriteString(w, "</methodName>\n"); err != nil {
			return
		}
	}
	if _, err = io.WriteString(w, "<params>\n"); err != nil {
		return
	}
	for _, arg := range args {
		if _, err = io.WriteString(w, "  <param><value>"); err != nil {
			return
		}
		if err = WriteXML(w, arg, true); err != nil {
			return
		}
		if _, err = io.WriteString(w, "</value></param>\n"); err != nil {
			return
		}
	}
	if name == "" {
		_, err = io.WriteString(w, "</params></methodResponse>")
	} else {
		_, err = io.WriteString(w, "</params></methodCall>")
	}
	return err
}

func getFault(v interface{}) (*Fault, bool) {
	if f, ok := v.(Fault); ok {
		return &f, true
	}
	if f, ok := v.(*Fault); ok {
		if f != nil {
			return f, true
		}
	} else {
		if e, ok := v.(error); ok {
			return &Fault{Code: -1, Message: e.Error()}, true
		}
	}
	return nil, false
}

// Unmarshal unmarshals the thing (methodResponse, methodCall or fault),
// returns the name of the method call in the first return argument;
// the params of the call or the response
// or the Fault if this is a Fault
func Unmarshal(r io.Reader) (name string, params []interface{}, fault *Fault, e error) {
	p := xml.NewDecoder(r)
	st := newParser(p)
	typ := "methodResponse"
	if _, e = st.getStart(typ); ErrEq(e, errNameMismatch) { // methodResponse or methodCall
		typ = "methodCall"
		if name, e = st.getText("methodName"); e != nil {
			return
		}
	}
	var se xml.StartElement
	if se, e = st.getStart("params"); e != nil {
		if ErrEq(e, errNameMismatch) && se.Name.Local == "fault" {
			var v interface{}
			if v, e = st.parseValue(); e != nil {
				return
			}
			fmap, ok := v.(map[string]interface{})
			if !ok {
				e = fmt.Errorf("fault not fault: %+v", v)
				return
			}
			fault = &Fault{Code: -1, Message: ""}
			code, ok := fmap["faultCode"]
			if !ok {
				e = fmt.Errorf("no faultCode in fault: %v", fmap)
				return
			}
			fcode, ok := code.(int)
			if !ok {
				e = fmt.Errorf("faultCode not int? %v", code)
				return
			}
			fault.Code = int(fcode)
			msg, ok := fmap["faultString"]
			if !ok {
				e = fmt.Errorf("no faultString in fault: %v", fmap)
				return
			}
			if fault.Message, ok = msg.(string); !ok {
				e = fmt.Errorf("faultString not strin? %v", msg)
				return
			}
			e = st.checkLast("fault")
		}
		return
	}
	params = make([]interface{}, 0, 8)
	var v interface{}
	for {
		if _, e = st.getStart("param"); e != nil {
			if ErrEq(e, errNotStartElement) {
				e = nil
				break
			}
			return
		}
		if v, e = st.parseValue(); e != nil {
			break
		}
		params = append(params, v)
		if e = st.checkLast("param"); e != nil {
			return
		}
	}
	if e = st.checkLast("params"); e == nil {
		e = st.checkLast(typ)
	}
	return
}

type errorStruct struct {
	main    error
	message string
}

func (es errorStruct) Error() string {
	return es.main.Error() + " [" + es.message + "]"
}

// Errorf2 returns an error embedding the main error with the formatted message
func Errorf2(err error, format string, a ...interface{}) error {
	return &errorStruct{main: err, message: fmt.Sprintf(format, a...)}
}

// ErrEq checks equality of the errorStructs (equality of the embedded main errors
func ErrEq(a, b error) bool {
	var maina, mainb error = a, b
	if esa, ok := a.(errorStruct); ok {
		maina = esa.main
	} else if esa, ok := a.(*errorStruct); ok {
		maina = esa.main
	}
	if esb, ok := b.(errorStruct); ok {
		mainb = esb.main
	} else if esb, ok := b.(*errorStruct); ok {
		mainb = esb.main
	}
	return maina == mainb
}
