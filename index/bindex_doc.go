// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package index

import (
	"strings"
	"fmt"
	"log"
)

func (in Index) SplitDoc(docId uint32) map[string]string {
	d, exists := in.Docs[docId]
	if !exists {
		return nil
	}
	m := make(map[string]string)

	s := strings.Split(string(d), "\t")

	kn := in.Meta.GetNode("attr", "order")

	for k, v := range s {
		if v == "" {
			// Don't return empty values
			continue
		}
		keyName := kn.GetString(fmt.Sprint(k))
		m[keyName] = v
	}
	return m
}
