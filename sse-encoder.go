// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Server-Sent Events
// W3C Working Draft 29 October 2009
// http://www.w3.org/TR/2009/WD-eventsource-20091029/

const ContentType = "text/event-stream"

var contentType = []string{ContentType}
var noCache = []string{"no-cache"}

var fieldReplacer = strings.NewReplacer(
	"\n", "\\n",
	"\r", "\\r")

var dataReplacer = strings.NewReplacer(
	"\n", "\ndata:",
	"\r", "\\r")

type Event struct {
	Event string
	Id    string
	Retry uint
	Data  interface{}
}

func Encode(writer io.Writer, event Event) (err error) {
	w := checkWriter(writer)
	err = writeId(w, event.Id)
	if err != nil {
		return
	}
	err = writeEvent(w, event.Event)
	if err != nil {
		return
	}
	err = writeRetry(w, event.Retry)
	if err != nil {
		return
	}
	return writeData(w, event.Data)
}

func writeId(w stringWriter, id string) (err error) {
	if len(id) > 0 {
		_, err = w.WriteString("id:")
		if err != nil {
			return
		}
		_, err = fieldReplacer.WriteString(w, id)
		if err != nil {
			return
		}
		_, err = w.WriteString("\n")
	}
	return
}

func writeEvent(w stringWriter, event string) (err error) {
	if len(event) > 0 {
		_, err = w.WriteString("event:")
		if err != nil {
			return
		}
		_, err = fieldReplacer.WriteString(w, event)
		if err != nil {
			return
		}
		_, err = w.WriteString("\n")
	}
	return
}

func writeRetry(w stringWriter, retry uint) (err error) {
	if retry > 0 {
		_, err = w.WriteString("retry:")
		if err != nil {
			return
		}
		_, err = w.WriteString(strconv.FormatUint(uint64(retry), 10))
		if err != nil {
			return
		}
		_, err = w.WriteString("\n")
	}
	return
}

func writeData(w stringWriter, data interface{}) (err error) {
	_, err = w.WriteString("data:")
	if err != nil {
		return
	}
	switch kindOfData(data) {
	case reflect.Struct, reflect.Slice, reflect.Map:
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			return
		}
		_, err = w.WriteString("\n")
	default:
		_, err = dataReplacer.WriteString(w, fmt.Sprint(data))
		if err != nil {
			return
		}
		_, err = w.WriteString("\n\n")
	}
	return
}

func (r Event) Render(w http.ResponseWriter) error {
	header := w.Header()
	header["Content-Type"] = contentType

	if _, exist := header["Cache-Control"]; !exist {
		header["Cache-Control"] = noCache
	}
	return Encode(w, r)
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
