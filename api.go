package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	jp "github.com/buger/jsonparser"
)

func crossref(path string) []byte {
	//

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.crossref.org/"+path, nil)
	req.Header.Set("User-Agent", "Hackathon Demo (https://github.com/InspiredGoat; mailto:tomd@airmail.cc)")
	res, _ := client.Do(req)

	buf := new(bytes.Buffer)
	defer res.Body.Close()

	buf.ReadFrom(res.Body)
	bytes := buf.Bytes()

	return bytes
}

func NewStudyFromDOI(doi string) *Study {
	buf := crossref("works/" + doi)
	return NewStudyFromBytes(buf)
}

func NewStudyFromBytes(buf []byte) *Study {
	s := new(Study)

	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		s.Title, _ = jp.GetString(value)
	}, "message", "title")
	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		given, err := jp.GetString(value, "given")
		if err != nil {
			fmt.Println(err.Error())
		}

		family, _ := jp.GetString(value, "family")
		s.Authors = append(s.Authors, family+" "+given)
	}, "message", "author")
	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		s.Subtitle, _ = jp.GetString(value)
	}, "message", "subtitle")
	s.IsReferencedCount, _ = jp.GetInt(buf, "message", "is-referenced-by-count")

	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		var r Reference
		r.Key, _ = jp.GetString(value, "key")
		r.DOI, _ = jp.GetString(value, "DOI")
		r.Unstructured, _ = jp.GetString(value, "unstructured")
		author, _ := jp.GetString(value, "author")
		year, _ := jp.GetString(value, "year")
		title, _ := jp.GetString(value, "volume-title")
		volume, _ := jp.GetString(value, "volume")
		issue, _ := jp.GetString(value, "issue")
		journal, _ := jp.GetString(value, "journal-title")
		page, _ := jp.GetString(value, "first-page")
		r.ArbitraryOrder = author + ". " + title + ". " + journal + ". " + year + ";" + volume + "(" + issue + ");" + page
		s.References = append(s.References, r)
	}, "message", "reference")

	s.Journal, _ = jp.GetString(buf, "message", "publisher")
	s.DOI, _ = jp.GetString(buf, "message", "DOI")

	timestampMilli, _ := jp.GetInt(buf, "message", "created", "timestamp")
	s.PublicationDate = time.UnixMilli(timestampMilli)

	return s
}

func (s *Study) ExpandChildren() {
	for _, r := range s.References {
		if len(r.DOI) > 4 {
			// fetch it and add to children
		}
	}

}

func fuckyoustub() {
	buf := crossref("works?query=petrol&rows=100")

	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		fmt.Println(jp.GetString(value, "abstract"))

		jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
			fmt.Println(jp.GetString(value, "abstract"))
		}, "author")
	}, "message", "items")
}
