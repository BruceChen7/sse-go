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

// 协议的content-type
const ContentType = "text/event-stream"

var contentType = []string{ContentType}
var noCache = []string{"no-cache"}

var fieldReplacer = strings.NewReplacer(
	"\n", "\\n",
	"\r", "\\r")

// strings.NewReplacer表示将后面的字符将会取代前面的字符
// 返回的是Replacer
var dataReplacer = strings.NewReplacer(
	"\n", "\ndata:",
	"\r", "\\r")

type Event struct {
	Event string
	Id    string
	Retry uint
	Data  interface{}
}

func Encode(writer io.Writer, event Event) error {
	w := checkWriter(writer)
	writeId(w, event.Id)
	writeEvent(w, event.Event)
	writeRetry(w, event.Retry)
	return writeData(w, event.Data)
}

func writeId(w stringWriter, id string) {
	// 写成id:1\n
	if len(id) > 0 {
		w.WriteString("id:")
		fieldReplacer.WriteString(w, id)
		w.WriteString("\n")
	}
}

func writeEvent(w stringWriter, event string) {
	// 直接写数据
	if len(event) > 0 {
		w.WriteString("event:")
		// 直接将数据中的转
		fieldReplacer.WriteString(w, event)
		w.WriteString("\n")
	}
}

func writeRetry(w stringWriter, retry uint) {
	if retry > 0 {
		w.WriteString("retry:")
		w.WriteString(strconv.FormatUint(uint64(retry), 10))
		w.WriteString("\n")
	}
}

func writeData(w stringWriter, data interface{}) error {
	w.WriteString("data:")
	switch kindOfData(data) {
	// 根据go中的数据结构，采用json来序列化
	case reflect.Struct, reflect.Slice, reflect.Map:
		// 对于slice，结构体，map，使用json来encode
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			return err
		}
		w.WriteString("\n")
		// 一定会走到default分支
	default:
		// 将替换后的字符串写到w中。
		dataReplacer.WriteString(w, fmt.Sprint(data))
		// 数据以\n\n结尾
		w.WriteString("\n\n")
	}
	return nil
}

func (r Event) Render(w http.ResponseWriter) error {
	header := w.Header()
	// 设置头
	header["Content-Type"] = contentType

	if _, exist := header["Cache-Control"]; !exist {
		header["Cache-Control"] = noCache
	}
	// 编码
	return Encode(w, r)
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	// 如果是指针类型，那么取相关乐行
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
