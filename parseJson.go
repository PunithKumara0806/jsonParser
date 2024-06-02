package main

import (
	"fmt"
	"reflect"
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
		}
		fmt.Println("value : ", value)
		break
		// case 't':
		// 	truef(ip,b)
		// 	break
		// case 'f':
		// 	falsef(ip,b)
		// 	break
		// case 'n':
		// 	nullf(ip,b)
		// 	break
	}
	// if isNumber(current) {
	// 	numberf(ip,b)
	// }
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

//
// func numberf() { TODO: Implement the below functions
// }
// func truef() {
// }
// func falsef() {
// }
// func nullf() {
// }

type Message struct {
	Name innerMessage
}
type innerMessage struct {
	Name2 string
}

func main() {
	b := []rune(`{"Name":{"Name2":"Valu\"e2"}}`)
	ip := 0
	var m Message
	fmt.Printf("Before : %+v\n", m)
	json(&ip, b, &m)
	fmt.Printf("After : %+v\n", m)
}
