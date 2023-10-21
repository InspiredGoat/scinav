package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	jp "github.com/buger/jsonparser"
)

func crossref(path string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.crossref.org/"+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Hackathon Demo (https://github.com/InspiredGoat; mailto:tomd@airmail.cc)")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	defer res.Body.Close()

	buf.ReadFrom(res.Body)
	bytes := buf.Bytes()

	return bytes, nil
}

func searchStudy(query string) {
}

func NewStudyFromDOI(doi string) (*Study, error) {
	buf, err := crossref("works/" + doi)
	if err != nil {
		return nil, err
	} else {
		return NewStudyFromBytes(buf), nil
	}
}

func NewStudyFromBytes(buf []byte) *Study {
	s := new(Study)

	s.FilteredIn = true
	s.Title, _ = jp.GetString(buf, "message", "title", "[0]")
	s.Subtitle, _ = jp.GetString(buf, "message", "subtitle", "[0]")

	jp.ArrayEach(buf, func(value []byte, dataType jp.ValueType, offset int, err error) {
		given, _ := jp.GetString(value, "given")
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
	if s.Expanded {
		return
	}

	wg := new(sync.WaitGroup)

	for _, r := range s.References {
		wg.Add(1)
		go func(s *Study, r Reference, wg *sync.WaitGroup) {
			defer wg.Done()
			if len(r.DOI) > 4 {
				// fetch it and add to children
				s_child, err := NewStudyFromDOI(r.DOI)

				if err != nil {
					fmt.Println("Couldn't expand child, because of error in fetching new study")
					fmt.Println(err.Error())
				} else {
					s_child.Parent = s
					s.Children = append(s.Children, s_child)
				}
			}
		}(s, r, wg)
	}

	wg.Wait()
	s.Expanded = true
}
