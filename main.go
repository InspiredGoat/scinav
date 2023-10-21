package main

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/gen2brain/raylib-go/raylib"
	"net/http"
)

func sendreq(req string) *http.Response {
	res, _ := http.Get(req + "&mailto=tomd@airmail.cc")
	return res
}

func main() {
	res := sendreq("https://api.crossref.org/works?query=petrol&rows=100")

	buf := new(bytes.Buffer)

	buf.ReadFrom(res.Body)
	defer res.Body.Close()

	jsonparser.ArrayEach(buf.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(jsonparser.GetString(value, "abstract"))

		jsonparser.ArrayEach(buf.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			fmt.Println(jsonparser.GetString(value, "abstract"))
		}, "author")
	}, "message", "items")

	// fmt.Println(string(val))
}
