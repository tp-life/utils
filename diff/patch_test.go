package diff

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	cases := []struct {
		Name      string
		A, B      interface{}
		Changelog Changelog
	}{
		{
			"uint-slice-insert", &[]uint{1, 2, 3}, &[]uint{1, 2, 3, 4},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: uint(4)},
			},
		},
		{
			"int-slice-insert", &[]int{1, 2, 3}, &[]int{1, 2, 3, 4},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: 4},
			},
		},
		{
			"uint-slice-delete", &[]uint{1, 2, 3}, &[]uint{1, 3},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: uint(2)},
			},
		},
		{
			"int-slice-delete", &[]int{1, 2, 3}, &[]int{1, 3},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
			},
		},
		{
			"uint-slice-insert-delete", &[]uint{1, 2, 3}, &[]uint{1, 3, 4},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: uint(2)},
				Change{Type: CREATE, Path: []string{"2"}, To: uint(4)},
			},
		},
		{
			"int-slice-insert-delete", &[]int{1, 2, 3}, &[]int{1, 3, 4},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
				Change{Type: CREATE, Path: []string{"2"}, To: 4},
			},
		},
		{
			"string-slice-insert", &[]string{"1", "2", "3"}, &[]string{"1", "2", "3", "4"},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: "4"},
			},
		},
		{
			"string-slice-delete", &[]string{"1", "2", "3"}, &[]string{"1", "3"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
			},
		},
		{
			"string-slice-insert-delete", &[]string{"1", "2", "3"}, &[]string{"1", "3", "4"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
				Change{Type: CREATE, Path: []string{"2"}, To: "4"},
			},
		},
		{
			"comparable-slice-update", &[]tistruct{{"one", 1}, {"two", 2}}, &[]tistruct{{"one", 1}, {"two", 50}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"two", "value"}, From: 1, To: 50},
			},
		},
		{
			"struct-string-update", &tstruct{Name: "one"}, &tstruct{Name: "two"},
			Changelog{
				Change{Type: UPDATE, Path: []string{"name"}, From: "one", To: "two"},
			},
		},
		{
			"struct-int-update", &tstruct{Value: 1}, &tstruct{Value: 50},
			Changelog{
				Change{Type: UPDATE, Path: []string{"value"}, From: 1, To: 50},
			},
		},
		{
			"struct-bool-update", &tstruct{Bool: true}, &tstruct{Bool: false},
			Changelog{
				Change{Type: UPDATE, Path: []string{"bool"}, From: true, To: false},
			},
		},
		{
			"struct-time-update", &tstruct{}, &tstruct{Time: currentTime},
			Changelog{
				Change{Type: UPDATE, Path: []string{"time"}, From: time.Time{}, To: currentTime},
			},
		},
		{
			"struct-nil-string-pointer-update", &tstruct{Pointer: nil}, &tstruct{Pointer: sptr("test")},
			Changelog{
				Change{Type: UPDATE, Path: []string{"pointer"}, From: nil, To: sptr("test")},
			},
		},
		{
			"struct-string-pointer-update-to-nil", &tstruct{Pointer: sptr("test")}, &tstruct{Pointer: nil},
			Changelog{
				Change{Type: UPDATE, Path: []string{"pointer"}, From: sptr("test"), To: nil},
			},
		}, {
			"struct-generic-slice-insert", &tstruct{Values: []string{"one"}}, &tstruct{Values: []string{"one", "two"}},
			Changelog{
				Change{Type: CREATE, Path: []string{"values", "1"}, From: nil, To: "two"},
			},
		},
		{
			"struct-generic-slice-delete", &tstruct{Values: []string{"one", "two"}}, &tstruct{Values: []string{"one"}},
			Changelog{
				Change{Type: DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
			},
		},
		{
			"struct-unidentifiable-slice-insert-delete", &tstruct{Unidentifiables: []tuistruct{{1}, {2}, {3}}}, &tstruct{Unidentifiables: []tuistruct{{5}, {2}, {3}, {4}}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"unidentifiables", "0", "value"}, From: 1, To: 5},
				Change{Type: CREATE, Path: []string{"unidentifiables", "3", "value"}, From: nil, To: 4},
			},
		},
		{
			"slice", &tstruct{}, &tstruct{Nested: tnstruct{Slice: []tmstruct{{"one", 1}, {"two", 2}}}},
			Changelog{
				Change{Type: CREATE, Path: []string{"nested", "slice", "0", "foo"}, From: nil, To: "one"},
				Change{Type: CREATE, Path: []string{"nested", "slice", "0", "bar"}, From: nil, To: 1},
				Change{Type: CREATE, Path: []string{"nested", "slice", "1", "foo"}, From: nil, To: "two"},
				Change{Type: CREATE, Path: []string{"nested", "slice", "1", "bar"}, From: nil, To: 2},
			},
		},
		{
			"slice-duplicate-items", &[]int{1}, &[]int{1, 1},
			Changelog{
				Change{Type: CREATE, Path: []string{"1"}, From: nil, To: 1},
			},
		},
		{
			"embedded-struct-field",
			&embedstruct{Embedded{Foo: "a", Bar: 2}, true},
			&embedstruct{Embedded{Foo: "b", Bar: 3}, false},
			Changelog{
				Change{Type: UPDATE, Path: []string{"foo"}, From: "a", To: "b"},
				Change{Type: UPDATE, Path: []string{"bar"}, From: 2, To: 3},
				Change{Type: UPDATE, Path: []string{"baz"}, From: true, To: false},
			},
		},
		{
			"custom-tags",
			&customTagStruct{Foo: "abc", Bar: 3},
			&customTagStruct{Foo: "def", Bar: 4},
			Changelog{
				Change{Type: UPDATE, Path: []string{"foo"}, From: "abc", To: "def"},
				Change{Type: UPDATE, Path: []string{"bar"}, From: 3, To: 4},
			},
		},
		{
			"custom-types",
			&customTypeStruct{Foo: "a", Bar: 1},
			&customTypeStruct{Foo: "b", Bar: 2},
			Changelog{
				Change{Type: UPDATE, Path: []string{"foo"}, From: CustomStringType("a"), To: CustomStringType("b")},
				Change{Type: UPDATE, Path: []string{"bar"}, From: CustomIntType(1), To: CustomIntType(2)},
			},
		},
		{
			"map",
			map[string]interface{}{"1": "one", "3": "three"},
			map[string]interface{}{"2": "two", "3": "tres"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "one", To: nil},
				Change{Type: CREATE, Path: []string{"2"}, From: nil, To: "two"},
				Change{Type: UPDATE, Path: []string{"3"}, From: "three", To: "tres"},
			},
		},
		{
			"map-nested-create",
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active", "attributes": map[string]interface{}{"attrA": "A", "attrB": "B"}}},
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active", "attributes": map[string]interface{}{"attrA": "A", "attrB": "B"}, "secondary-attributes": map[string]interface{}{"attrA": "A", "attrB": "B"}}},
			Changelog{
				Change{Type: "create", Path: []string{"details", "secondary-attributes"}, From: nil, To: map[string]interface{}{"attrA": "A", "attrB": "B"}},
			},
		},
		{
			"map-nested-update",
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active", "attributes": map[string]interface{}{"attrA": "A", "attrB": "B"}}},
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active", "attributes": map[string]interface{}{"attrA": "C", "attrD": "X"}}},
			Changelog{
				Change{Type: "update", Path: []string{"details", "attributes"}, From: map[string]interface{}{"attrA": "A", "attrB": "B"}, To: map[string]interface{}{"attrA": "C", "attrD": "X"}},
			},
		},
		{
			"map-nested-delete",
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active", "attributes": map[string]interface{}{"attrA": "A", "attrB": "B"}}},
			map[string]interface{}{"firstName": "John", "lastName": "Michael", "createdBy": "TS", "details": map[string]interface{}{"status": "active"}},
			Changelog{
				Change{Type: "delete", Path: []string{"details", "attributes"}, From: map[string]interface{}{"attrA": "A", "attrB": "B"}, To: nil},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			var options []func(d *Differ) error
			switch tc.Name {
			case "mixed-slice-map", "nil-map", "map-nil":
				options = append(options, StructMapKeySupport())
			case "embedded-struct-field":
				options = append(options, FlattenEmbeddedStructs())
			case "custom-tags":
				options = append(options, TagName("json"))
			}
			d, err := NewDiffer(options...)
			if err != nil {
				panic(err)
			}
			pl := d.Patch(tc.Changelog, tc.A)

			assert.Equal(t, tc.B, tc.A)
			require.Equal(t, len(tc.Changelog), len(pl))
			assert.False(t, pl.HasErrors())
		})
	}

	t.Run("convert-types", func(t *testing.T) {
		a := &tmstruct{Foo: "a", Bar: 1}
		b := &customTypeStruct{Foo: "b", Bar: 2}
		cl := Changelog{
			Change{Type: UPDATE, Path: []string{"foo"}, From: CustomStringType("a"), To: CustomStringType("b")},
			Change{Type: UPDATE, Path: []string{"bar"}, From: CustomIntType(1), To: CustomIntType(2)},
		}

		d, err := NewDiffer()
		if err != nil {
			panic(err)
		}
		pl := d.Patch(cl, a)

		assert.True(t, pl.HasErrors())

		d, err = NewDiffer(ConvertCompatibleTypes())
		if err != nil {
			panic(err)
		}
		pl = d.Patch(cl, a)

		assert.False(t, pl.HasErrors())
		assert.Equal(t, string(b.Foo), a.Foo)
		assert.Equal(t, int(b.Bar), a.Bar)
		require.Equal(t, len(cl), len(pl))
	})

	t.Run("pointer", func(t *testing.T) {
		type tps struct {
			S *string
		}

		str1 := "before"
		str2 := "after"

		t1 := tps{S: &str1}
		t2 := tps{S: &str2}

		changelog, err := Diff(t1, t2)
		assert.NoError(t, err)

		patchLog := Patch(changelog, &t1)
		assert.False(t, patchLog.HasErrors())
	})

	t.Run("map-to-pointer", func(t *testing.T) {
		t1 := make(map[string]*tmstruct)
		t2 := map[string]*tmstruct{"after": {Foo: "val"}}

		changelog, err := Diff(t1, t2)
		assert.NoError(t, err)

		patchLog := Patch(changelog, &t1)
		assert.False(t, patchLog.HasErrors())

		assert.True(t, len(t2) == len(t1))
		assert.Equal(t, t1["after"], &tmstruct{Foo: "val"})
	})

	t.Run("map-to-nil-pointer", func(t *testing.T) {
		t1 := make(map[string]*tmstruct)
		t2 := map[string]*tmstruct{"after": nil}

		changelog, err := Diff(t1, t2)
		assert.NoError(t, err)

		patchLog := Patch(changelog, &t1)
		assert.False(t, patchLog.HasErrors())

		assert.Equal(t, len(t2), len(t1))
		assert.Nil(t, t1["after"])
	})

	t.Run("pointer-with-converted-type", func(t *testing.T) {
		type tps struct {
			S *int
		}

		val1 := 1
		val2 := 2

		t1 := tps{S: &val1}
		t2 := tps{S: &val2}

		changelog, err := Diff(t1, t2)
		assert.NoError(t, err)

		js, err := json.Marshal(changelog)
		assert.NoError(t, err)

		assert.NoError(t, json.Unmarshal(js, &changelog))

		d, err := NewDiffer(ConvertCompatibleTypes())
		assert.NoError(t, err)

		assert.Equal(t, 1, *t1.S)

		patchLog := d.Patch(changelog, &t1)
		assert.False(t, patchLog.HasErrors())
		assert.Equal(t, 2, *t1.S)

		// test nil pointer
		t1 = tps{S: &val1}
		t2 = tps{S: nil}

		changelog, err = Diff(t1, t2)
		assert.NoError(t, err)

		patchLog = d.Patch(changelog, &t1)
		assert.False(t, patchLog.HasErrors())
	})
}
