package pantherlog

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/pantherlog/null"
	"github.com/panther-labs/panther/pkg/box"
)

type testStringer struct {
	Foo string
}

func (t *testStringer) String() string {
	return t.Foo
}
func (t *testStringer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Foo)
}

var (
	// Register our own random value kinds
	kindFoo  = FieldID(time.Now().UnixNano())
	kindBar  = kindFoo + 1
	kindBaz  = kindFoo + 2
	kindQux  = kindFoo + 3
	kindQuux = kindFoo + 4
)

func init() {
	MustRegisterIndicator(kindFoo, FieldMeta{
		Name:        "PantherFoo",
		NameJSON:    "p_any_foo",
		Description: "Foo data",
	})
	MustRegisterIndicator(kindBar, FieldMeta{
		Name:        "PantherBar",
		NameJSON:    "p_any_bar",
		Description: "Bar data",
	})
	MustRegisterIndicator(kindBaz, FieldMeta{
		Name:        "PantherBaz",
		NameJSON:    "p_any_baz",
		Description: "Baz data",
	})
	MustRegisterIndicator(kindQux, FieldMeta{
		Name:        "PantherQux",
		NameJSON:    "p_any_qux",
		Description: "Qux data",
	})
	MustRegisterIndicator(kindQuux, FieldMeta{
		Name:        "PantherQuux",
		NameJSON:    "p_any_quux",
		Description: "Quux data",
	})
	MustRegisterScanner("foo", kindFoo, kindFoo)
	MustRegisterScanner("bar", kindBar, kindBar)
	MustRegisterScanner("baz", kindBaz, kindBaz)
	MustRegisterScanner("qux", kindQux, kindQux)
	MustRegisterScanner("quux", kindQuux, kindQuux)
}

func TestPantherExt_DecorateEncoder(t *testing.T) {
	// Check all possible string types
	type T struct {
		Foo  *testStringer `json:"foo" panther:"foo"`
		Bar  testStringer  `json:"bar" panther:"bar"`
		Baz  string        `json:"baz" panther:"baz"`
		Qux  *string       `json:"qux" panther:"qux"`
		Quux null.String   `json:"quux" panther:"quux"`
	}

	v := T{
		Foo: &testStringer{
			Foo: "ok",
		},
		Bar: testStringer{
			Foo: "ok",
		},
		Baz:  "ok",
		Qux:  box.String("ok"),
		Quux: null.FromString("ok"),
	}

	result := Result{
		values: new(ValueBuffer),
	}
	stream := jsoniter.ConfigDefault.BorrowStream(nil)
	stream.Attachment = &result
	stream.WriteVal(&v)
	require.Equal(t, []string{"ok"}, result.values.Get(kindFoo), "foo")
	require.Equal(t, []string{"ok"}, result.values.Get(kindBar), "bar")
	require.Equal(t, []string{"ok"}, result.values.Get(kindBaz), "baz")
	require.Equal(t, []string{"ok"}, result.values.Get(kindQux), "qux")
	require.Equal(t, []string{"ok"}, result.values.Get(kindQuux), "quux")
	actual := string(stream.Buffer())
	require.Equal(t, `{"foo":"ok","bar":"ok","baz":"ok","qux":"ok","quux":"ok"}`, actual)
}

func TestResultEncoder(t *testing.T) {
	now := time.Now()
	tm := now.Add(-1 * time.Minute)
	loc, err := time.LoadLocation(`Europe/Athens`)
	assert := require.New(t)
	assert.NoError(err)
	type T struct {
		Time     time.Time `json:"tm" panther:"event_time"`
		RemoteIP string    `json:"remote_ip" panther:"ip"`
		LocalIP  string    `json:"local_ip" panther:"ip"`
	}
	event := T{
		Time:     tm.In(loc),
		RemoteIP: "2.2.2.2",
		LocalIP:  "1.1.1.1",
	}
	result := Result{
		CoreFields: CoreFields{
			PantherLogType:   "Foo.Bar",
			PantherRowID:     "id",
			PantherParseTime: now.UTC(),
		},
		Event: &event,
	}
	actual, err := jsoniter.MarshalToString(&result)
	assert.NoError(err)
	expect := fmt.Sprintf(`{
		"tm": "%s",
		"remote_ip":"2.2.2.2",
		"local_ip":"1.1.1.1",
		"p_row_id": "id",
		"p_event_time": "%s",
		"p_parse_time": "%s",
		"p_any_ip_addresses": ["1.1.1.1", "2.2.2.2"],
		"p_log_type": "Foo.Bar"
	}`, tm.In(loc).Format(time.RFC3339Nano), tm.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano))
	assert.JSONEq(expect, actual)
}
