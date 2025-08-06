package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

//NOTE: Found JSON grammer from https://www.json.org/json-en.html, https://www.crockford.com/mckeeman.html

var debug_flag bool

/*
debugln => takes input same as fmt.Println,functions the same as fmt.Println
except it can be toggled by debug_flag:bool to print
*/
func debugln(a ...any) {
	if !debug_flag {
		return
	}
	fmt.Println(a...)
}

func debugf(s string, a ...any) {
	if !debug_flag {
		return
	}
	fmt.Printf(s, a...)
}

/*
function consume => returns error if the current rune does not match the give rune,
else increments the ip and returns nil
*/
func consume(c rune, ip *int, b []rune, msg string) error {
	// debugln("consume : ", c)
	if len(b) > *ip && b[*ip] == c {
		*ip++
		return nil
	}
	return errors.New(msg)
}

func jsonToObj[T interface{}](ip *int, b []rune, o *T) {
	r := reflect.ValueOf(o).Elem()
	err := elementf(ip, b, r)
	if err != nil {
		debugln(err)
	}
}

func elementf(ip *int, b []rune, r reflect.Value) error {
	wsf(ip, b)
	err := valuef(ip, b, r)
	wsf(ip, b)
	return err
}

func elementsf(ip *int, b []rune, r reflect.Value) error {
	err := elementf(ip, b, r)
	if err != nil {
		return err
	}
	if b[*ip] == ',' {
		*ip++
		err = elementsf(ip, b, r)
		if err != nil {
			return err
		}
	}
	return err
}

/*
function wsf => Consumes whitespaces : ' '
*/
func wsf(ip *int, b []rune) {
	for len(b) > *ip {
		if b[*ip] == ' ' || b[*ip] == '\t' || b[*ip] == '\r' || b[*ip] == '\n' {
			*ip++
			continue
		}
		break
	}
}

func valuef(ip *int, b []rune, r reflect.Value) error {
	if len(b) == *ip {
		return errors.New("No value: EOF")
	}
	current := b[*ip]
	switch current {
	case '{':
		return objectf(ip, b, r)
	case '[':
		return arrayf(ip, b, r)
	case '"':
		value, err := stringf(ip, b)
		if err != nil {
			return err
		}
		if r.Kind() == reflect.String {
			r.SetString(value)
		} else {
			return errors.New(fmt.Sprintf("type mismatch ->  have %s, given %s\n", r.Kind(), "string"))
		}
		debugln("value : ", value)
		return nil
	case 't':
		return boolf(ip, b, r, "true")
	case 'f':
		return boolf(ip, b, r, "false")
		// case 'n':
		// 	nullf(ip,b)
		// 	break
	}
	if isNumber(current) {
		return numberf(ip, b, r)
	}
	return errors.New(fmt.Sprintf("Expected value, got %s\n", string(b[*ip])))
}

func objectf(ip *int, b []rune, r reflect.Value) error {
	// debugln("objectf")
	consume('{', ip, b, "Not not an object, no '{'")
	wsf(ip, b)
	err := membersf(ip, b, r)
	if err != nil {
		return err
	}
	err = consume('}', ip, b, "Not found Ending '}'")
	return err
}

func membersf(ip *int, b []rune, r reflect.Value) error {
	// debugln("membersf")
	err := memberf(ip, b, r)
	if err != nil || len(b) == *ip {
		return err
	}
	if b[*ip] == ',' {
		*ip++
		err = membersf(ip, b, r)
		if err != nil {
			return err
		}
	}
	return err
}

func memberf(ip *int, b []rune, r reflect.Value) error {
	// debugln("memberf")
	wsf(ip, b)
	key, err := stringf(ip, b)
	if err != nil {
		return err
	}
	debugln("key : ", key)
	innerR := r.FieldByName(key)
	if !innerR.IsValid() {
		debugln("Not valid field name")
		return nil
	}
	wsf(ip, b)
	consume(':', ip, b, "Error not found ':' after field")
	err = elementf(ip, b, innerR)
	if err != nil {
		fmt.Printf("Field '%s' : ", key)
		fmt.Println(err)
	}
	return nil
}

func arrayf(ip *int, b []rune, r reflect.Value) error {
	consume('[', ip, b, "Error not found '['")
	var err error
	if r.Type().Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("(%s) is not array/slice type\n", r.Type().Name()))
	}
	for {
		newItem := reflect.New(r.Type().Elem())
		err = elementf(ip, b, newItem.Elem())
		//break on error
		if err != nil {
			break
		}
		r.Set(reflect.Append(r, newItem.Elem()))
		wsf(ip, b)
		//break on not having remaining element
		if b[*ip] != ',' {
			break
		}
		*ip++
	}
	consume(']', ip, b, "Error not found ending ']'")
	return err
}

func stringf(ip *int, b []rune) (string, error) {
	// debugln("stringf")
	err := consume('"', ip, b, "String does not start with '\"'")
	if err != nil {
		return "", err
	}

	str := make([]rune, 0, 8)
	for len(b) > *ip {
		if b[*ip] == '\\' {
			//TODO: implement such that only certain prescribed characters are escapable,
			str = append(str, b[*ip+1])
			*ip += 2
			continue
		}
		if b[*ip] == '"' {
			break
		}
		str = append(str, b[*ip])
		*ip++
	}

	err = consume('"', ip, b, "String does not end with '\"'")
	if err != nil {
		return "", err
	}
	return string(str), nil
}

/*
boolf(ip *int, b []rune, r reflect.Value, value string)

A function which is used to parse and set boolean values to the field
It completely consumes the boolean value and points ip to the next rune.
*/
func boolf(ip *int, b []rune, r reflect.Value, value string) error {
	var err error
	if string(b[*ip:*ip+len(value)]) == value {
		if r.Kind() == reflect.Bool {
			if value[0] == 't' {
				r.SetBool(true)
			} else {
				r.SetBool(false)
			}
			*ip += len(value)
		} else {
			err = errors.New("Field not set for bool")
		}
	}
	return err

}

/*
It felt neccessary but :( fount no use
*/
// func synchronize(ip *int, b []rune) {
// 	for len(b) > *ip {
// 		if b[*ip] == ',' || b[*ip] == ']' || b[*ip] == '}' {
// 			break
// 		}
// 		*ip++
// 	}
// }

/*
checks if the rune is within 0 - 9 or '-'
*/
func isNumber(b rune) bool {
	return (b >= '0' && b <= '9') || b == '-'
}

/*
function numberf => Parses the source runes to number ( int64 , float32 ) depending on the decimal
else returns error
*/
func numberf(ip *int, b []rune, r reflect.Value) error {
	hasfraction := false

	str := make([]rune, 0, 8)

	if b[*ip] == '-' {
		str = append(str, b[*ip])
		*ip++
	}

	for !(b[*ip] == ',' || b[*ip] == ']' || b[*ip] == '}' || b[*ip] == ' ') {
		if b[*ip] == '.' {
			hasfraction = true
		}
		str = append(str, b[*ip])
		*ip++
	}
	if hasfraction {
		i, err := strconv.ParseFloat(string(str), 32)
		if err != nil {
			return err
		}
		r.SetFloat(i)
	} else {
		i, err := strconv.Atoi(string(str))
		if err != nil {
			return err
		}
		r.SetInt(int64(i))
	}
	return nil
}

func encodeString(b *bytes.Buffer, name string) {
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteByte('"')
}

func encodeSlice(b *bytes.Buffer, r reflect.Value) {
	b.WriteByte('[')
	for i := 0; i < r.Len(); i++ {
		encodeValue(b, r.Index(i))
		if i != r.Len()-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte(']')

}

func encodeValue(b *bytes.Buffer, r reflect.Value) {
	switch r.Kind() {
	case reflect.String:
		encodeString(b, r.String())
	case reflect.Int:
		b.WriteString(fmt.Sprint(r.Int()))
	case reflect.Float64:
		b.WriteString(fmt.Sprint(r.Float()))
	case reflect.Slice:
		encodeSlice(b, r)
	case reflect.Bool:
		b.WriteString(fmt.Sprint(r.Bool()))
	default:
		buildJson(r, b)
	}

}

func buildJson(r reflect.Value, b *bytes.Buffer) error {
	b.WriteByte('{')
	for i := 0; i < r.NumField(); i++ {
		//Key
		encodeString(b, r.Type().Field(i).Name)
		b.WriteByte(':')
		//Value
		encodeValue(b, r.Field(i))
		if i != r.NumField()-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte('}')
	return nil
}

/*
Function that parses json and populates to the object give
*/
func Decode[T interface{}](b []byte, o *T) {
	ip := 0
	jsonToObj(&ip, []rune(string(b)), o)
}

/*
Function that converts object to its equivalent json
*/
func Encode[T interface{}](a T) ([]byte, error) {
	r := reflect.ValueOf(a)
	b := bytes.NewBuffer(nil)
	err := buildJson(r, b)
	return b.Bytes(), err
}

type Message struct {
	Name   innerMessage
	Number int
	Float  float64
	String string
	Flag3  []int
}
type innerMessage struct {
	Name2 string
	Flag  bool
	Flag2 []int
}

func main() {
	b := `{	 "Flag3":[1,2] ,"String" : "hello","Name":{"Name2":"Valu\"e2", "Flag": true, "Flag2":[1,2]},"Number":-232,"Float":}`
	var m Message
	fmt.Printf("Before : %+v\n", m)
	Decode([]byte(b), &m)
	fmt.Printf("After : %+v\n", m)
	result, _ := Encode(m)
	fmt.Println(" Encoded : ", string(result))
}
