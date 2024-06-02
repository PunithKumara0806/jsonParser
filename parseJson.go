package main

import (
	"fmt"
	"reflect"
	"strconv"
)

//NOTE: Found JSON grammer from https://www.json.org/json-en.html, https://www.crockford.com/mckeeman.html

func json[T interface{}](ip *int, b []rune, o *T) {
	// fmt.Println("json")
	r := reflect.ValueOf(o).Elem()
	elementf(ip, b, r)
}

func elementf(ip *int, b []rune, r reflect.Value) {
	// fmt.Println("elementf")
	wsf(ip, b)
	valuef(ip, b, r)
	wsf(ip, b)
}

func elementsf(ip *int, b []rune, r reflect.Value) {
	// fmt.Println("elementsf")
	elementf(ip, b, r)
	if b[*ip] == ',' {
		*ip++
		elementsf(ip, b, r)
	}
}

func wsf(ip *int, b []rune) {
	// fmt.Println("wsf")
	for len(b) > *ip {
		if b[*ip] != ' ' {
			break
		}
		*ip++
	}
	//TODO: skip whitespaces
}

func valuef(ip *int, b []rune, r reflect.Value) {
	if len(b) == *ip {
		return
	}
	// fmt.Println("valuef")
	current := b[*ip]
	switch current {
	case '{':
		objectf(ip, b, r)
		break
	case '[':
		arrayf(ip, b, r)
		break
	case '"':
		value := stringf(ip, b)
		if r.Kind() == reflect.String {
			r.SetString(value)
		} else {
			fmt.Printf("Field  type mismatch : have %s, given %s\n", r.Kind(), "string")
		}
		fmt.Println("value : ", value)
		break
	case 't':
		boolf(ip, b, r, "true")
		break
	case 'f':
		boolf(ip, b, r, "false")
		break
		// case 'n':
		// 	nullf(ip,b)
		// 	break
	}
	if isNumber(current) {
		numberf(ip, b, r)
	}
	synchronize(ip, b)
}

func objectf(ip *int, b []rune, r reflect.Value) {
	// fmt.Println("objectf")
	consume('{', ip, b, "Not not an object, no '{'")
	wsf(ip, b)
	membersf(ip, b, r)
	consume('}', ip, b, "Not found Ending '}'")
}

func membersf(ip *int, b []rune, r reflect.Value) {
	// fmt.Println("membersf")
	memberf(ip, b, r)
	if b[*ip] == ',' {
		*ip++
		membersf(ip, b, r)
	}
}

func memberf(ip *int, b []rune, r reflect.Value) {
	// fmt.Println("memberf")
	wsf(ip, b)
	key := stringf(ip, b)
	fmt.Println("key : ", key)
	innerR := r.FieldByName(key)
	if !innerR.IsValid() {
		fmt.Println("Not valid field name")
		return
	}
	wsf(ip, b)
	consume(':', ip, b, "Error not found ':' after field")
	elementf(ip, b, innerR)
}

func arrayf(ip *int, b []rune, r reflect.Value) {
	consume('[', ip, b, "Error not found '['")
	elementsf(ip, b, r)
	consume(']', ip, b, "Error not found ending ']'")
}

func consume(c rune, ip *int, b []rune, msg string) {
	// fmt.Println("consume : ", c)
	if b[*ip] == c {
		*ip++
	} else {
		fmt.Printf(msg)
	}
}

func stringf(ip *int, b []rune) string {
	// fmt.Println("stringf")
	consume('"', ip, b, "String does not start with '\"'")
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
	consume('"', ip, b, "String does not end with '\"'")
	return string(str)
}

/*
boolf(ip *int, b []rune, r reflect.Value, value string)

A function which is used to parse and set boolean values to the field
It completely consumes the boolean value and points ip to the next rune.
*/
func boolf(ip *int, b []rune, r reflect.Value, value string) {
	if string(b[*ip:*ip+len(value)]) == value {
		if r.Kind() == reflect.Bool {
			if value[0] == 't' {
				r.SetBool(true)
			} else {
				r.SetBool(false)
			}
			*ip += len(value)
		} else {
			fmt.Println("Field not set for bool")
		}
	} else {
		synchronize(ip, b)
	}

}

func synchronize(ip *int, b []rune) {
	for len(b) > *ip {
		if b[*ip] == ',' || b[*ip] == ']' || b[*ip] == '}' {
			break
		}
		*ip++
	}
}

func isNumber(b rune) bool {
	return (b >= '0' && b <= '9') || b == '-'
}

func numberf(ip *int, b []rune, r reflect.Value) {
	isfraction := false

	str := make([]rune, 0, 8)
	if b[*ip] == '-' {
		str = append(str, b[*ip])
		*ip++
	}
	for (b[*ip] <= '9' && b[*ip] >= '0') || b[*ip] == '.' {
		if b[*ip] == '.' {
			isfraction = true
		}
		str = append(str, b[*ip])
		*ip++
	}
	if isfraction {
		i, err := strconv.ParseFloat(string(str), 32)
		if err != nil {
			fmt.Println("Error parsing float")
			return
		}
		r.SetFloat(i)
	} else {
		i, err := strconv.Atoi(string(str))
		if err != nil {
			fmt.Println("Error parsing int")
			return
		}
		r.SetInt(int64(i))
	}
}

// func nullf() { TODO: Implement the below functions
// }

type Message struct {
	Name   innerMessage
	Number int
	Float  float64
}
type innerMessage struct {
	Name2 string
	Flag  bool
	Flag2 bool
}

func main() {
	b := []rune(`{   "Name":{"Name2":"Valu\"e2", "Flag": true, "Flag2":false},"Number":-232,"Float":32.23}`)
	ip := 0
	var m Message
	fmt.Printf("Before : %+v\n", m)
	json(&ip, b, &m)
	fmt.Printf("After : %+v\n", m)
}
