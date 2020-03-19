// nolint gomnd
package copystruct_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/bingoohuang/gor/copystruct"

	"github.com/guregu/null"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	assert "github.com/stretchr/testify/require"
)

type UserName string

// Model
type User struct {
	Name UserName
}

func (u *User) MethodThatTakesContext(ctx map[string]interface{}) string {
	// do whatever you want
	return "xx"
}

type UserResourceContext string

// Resource
type UserResource struct {
	DisplayName            string              `copystruct:"field:Name;convert"`
	SkipMe                 string              `copystruct:"skip"`
	MethodThatTakesContext UserResourceContext `copystruct:"context;convert"`
}

func TestExample1(t *testing.T) {
	user := &User{
		Name: "gilles",
	}

	resource := UserResource{}

	userNameType := reflect.TypeOf(user.Name)
	displayNameType := reflect.TypeOf(resource.DisplayName)

	fmt.Println("ConvertibleTo:", userNameType.ConvertibleTo(displayNameType))
	fmt.Println("ConvertibleTo:", displayNameType.ConvertibleTo(userNameType))
	fmt.Println("AssignableTo:", userNameType.AssignableTo(displayNameType))

	copystruct.Copy(user).To(&resource)
	assert.Equal(t, UserResource{
		DisplayName:            "gilles",
		SkipMe:                 "",
		MethodThatTakesContext: "xx",
	}, resource)
}

type M map[string]interface{}

func TestField(t *testing.T) {
	type (
		Rel struct {
			Int int
		}

		Src struct {
			Int       int
			IntPtr    *int
			Slice     []string
			SlicePtr  *[]string
			Map       M
			MapPtr    *M
			Struct    Rel
			StructPtr *Rel
			Skipped   string `copystruct:"skip"`
		}

		Dst struct {
			Int       int
			IntPtr    *int
			Slice     []string
			SlicePtr  *[]string
			Map       M
			MapPtr    *M
			Struct    Rel
			StructPtr *Rel
			Skipped   string `copystruct:"skip"`
		}

		Renamed struct {
			MyInt       int       `copystruct:"field:Int"`
			MyIntPtr    *int      `copystruct:"field:IntPtr"`
			MySlice     []string  `copystruct:"field:Slice"`
			MySlicePtr  *[]string `copystruct:"field:SlicePtr"`
			MyMap       M         `copystruct:"field:Map"`
			MyMapPtr    *M        `copystruct:"field:MapPtr"`
			MyStruct    Rel       `copystruct:"field:Struct"`
			MyStructPtr *Rel      `copystruct:"field:StructPtr"`
			Skipped     string    `copystruct:"skip"`
		}
	)

	integer := 1
	rel := Rel{Int: 1}
	slc := []string{"one", "two"}
	mp := M{"one": 1}

	src := &Src{
		Int:       integer,
		IntPtr:    &integer,
		Slice:     slc,
		SlicePtr:  &slc,
		Map:       mp,
		MapPtr:    &mp,
		Struct:    rel,
		StructPtr: &rel,
		Skipped:   "I should be skipped",
	}

	srcRenamed := &Renamed{
		MyInt:       integer,
		MyIntPtr:    &integer,
		MySlice:     slc,
		MySlicePtr:  &slc,
		MyMap:       mp,
		MyMapPtr:    &mp,
		MyStruct:    rel,
		MyStructPtr: &rel,
		Skipped:     "I should be skipped",
	}

	//
	// To()
	//

	dst := &Dst{}
	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Equal(t, src.Int, dst.Int)
	assert.Equal(t, src.IntPtr, dst.IntPtr)
	assert.Equal(t, src.Slice, dst.Slice)
	assert.Equal(t, src.SlicePtr, dst.SlicePtr)
	assert.Equal(t, src.Map, dst.Map)
	assert.Equal(t, src.MapPtr, dst.MapPtr)
	assert.Equal(t, src.Struct, dst.Struct)
	assert.Equal(t, src.StructPtr, dst.StructPtr)
	assert.Zero(t, dst.Skipped)

	dstRenamed := &Renamed{}
	assert.Nil(t, copystruct.Copy(src).To(dstRenamed))
	assert.Equal(t, src.Int, dstRenamed.MyInt)
	assert.Equal(t, src.IntPtr, dstRenamed.MyIntPtr)
	assert.Equal(t, src.Slice, dstRenamed.MySlice)
	assert.Equal(t, src.SlicePtr, dstRenamed.MySlicePtr)
	assert.Equal(t, src.Map, dstRenamed.MyMap)
	assert.Equal(t, src.MapPtr, dstRenamed.MyMapPtr)
	assert.Equal(t, src.Struct, dstRenamed.MyStruct)
	assert.Equal(t, src.StructPtr, dstRenamed.MyStructPtr)
	assert.Zero(t, dstRenamed.Skipped)

	//
	// From()
	//

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(src))
	assert.Equal(t, src.Int, dst.Int)
	assert.Equal(t, src.IntPtr, dst.IntPtr)
	assert.Equal(t, src.Slice, dst.Slice)
	assert.Equal(t, src.SlicePtr, dst.SlicePtr)
	assert.Equal(t, src.Map, dst.Map)
	assert.Equal(t, src.MapPtr, dst.MapPtr)
	assert.Equal(t, src.Struct, dst.Struct)
	assert.Equal(t, src.StructPtr, dst.StructPtr)
	assert.Zero(t, dst.Skipped)

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(srcRenamed))
	assert.Equal(t, srcRenamed.MyInt, dst.Int)
	assert.Equal(t, srcRenamed.MyIntPtr, dst.IntPtr)
	assert.Equal(t, srcRenamed.MySlice, dst.Slice)
	assert.Equal(t, srcRenamed.MySlicePtr, dst.SlicePtr)
	assert.Equal(t, srcRenamed.MyMap, dst.Map)
	assert.Equal(t, srcRenamed.MyMapPtr, dst.MapPtr)
	assert.Equal(t, srcRenamed.MyStruct, dst.Struct)
	assert.Equal(t, srcRenamed.MyStructPtr, dst.StructPtr)
	assert.Zero(t, dst.Skipped)
}

func TestField_PointerToValue(t *testing.T) {
	type (
		Rel struct {
			Int int
		}

		Src struct {
			Int    *int
			Slice  *[]string
			Map    *M
			Struct *Rel
		}

		Dst struct {
			Int    int
			Slice  []string
			Map    M
			Struct Rel
		}

		SrcRenamed struct {
			MyInt    *int      `copystruct:"field:Int"`
			MySlice  *[]string `copystruct:"field:Slice"`
			MyMap    *M        `copystruct:"field:Map"`
			MyStruct *Rel      `copystruct:"field:Struct"`
		}

		DstRenamed struct {
			MyInt    int      `copystruct:"field:Int"`
			MySlice  []string `copystruct:"field:Slice"`
			MyMap    M        `copystruct:"field:Map"`
			MyStruct Rel      `copystruct:"field:Struct"`
		}
	)

	rel := Rel{Int: 1}
	integer := 1
	slc := []string{"one", "two"}
	mp := M{"one": 1}

	src := &Src{
		Int:    &integer,
		Slice:  &slc,
		Map:    &mp,
		Struct: &rel,
	}

	srcRenamed := &SrcRenamed{
		MyInt:    &integer,
		MySlice:  &slc,
		MyMap:    &mp,
		MyStruct: &rel,
	}

	//
	// To()
	//

	dst := &Dst{}
	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Equal(t, *src.Int, dst.Int)
	assert.Equal(t, *src.Slice, dst.Slice)
	assert.Equal(t, *src.Map, dst.Map)
	assert.Equal(t, *src.Struct, dst.Struct)

	dstRenamed := &DstRenamed{}
	assert.Nil(t, copystruct.Copy(src).To(dstRenamed))
	assert.Equal(t, *src.Int, dstRenamed.MyInt)
	assert.Equal(t, *src.Slice, dstRenamed.MySlice)
	assert.Equal(t, *src.Map, dstRenamed.MyMap)
	assert.Equal(t, *src.Struct, dstRenamed.MyStruct)

	//
	// From()
	//

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(src))
	assert.Equal(t, *src.Int, dst.Int)
	assert.Equal(t, *src.Slice, dst.Slice)
	assert.Equal(t, *src.Map, dst.Map)
	assert.Equal(t, *src.Struct, dst.Struct)

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(srcRenamed))
	assert.Equal(t, *srcRenamed.MyInt, dst.Int)
	assert.Equal(t, *srcRenamed.MySlice, dst.Slice)
	assert.Equal(t, *srcRenamed.MyMap, dst.Map)
	assert.Equal(t, *srcRenamed.MyStruct, dst.Struct)
}

func TestField_Unexported(t *testing.T) {
	type (
		Src struct {
			Exported   int
			unexported string
		}

		Dst struct {
			Exported   int
			unexported string
		}
	)

	src := &Src{Exported: 1, unexported: "unexported"}

	//
	// To()
	//

	dst := &Dst{}
	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Equal(t, "", dst.unexported)

	//
	// From()
	//

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(src))
	assert.Equal(t, "", dst.unexported)
}

func TestField_Unknown(t *testing.T) {
	type Original struct {
		Int int
	}

	type Renamed struct {
		MyInt int `copystruct:"field:Integer"`
	}

	//
	// To()
	//

	src := &Original{Int: 1}
	dstRenamed := &Renamed{}
	assert.Nil(t, copystruct.Copy(src).To(dstRenamed))
	assert.Equal(t, 0, dstRenamed.MyInt)

	//
	// From()
	//

	srcRenamed := &Renamed{MyInt: 1}
	dst := &Original{}
	assert.Nil(t, copystruct.Copy(dst).From(srcRenamed))
	assert.Equal(t, 0, dst.Int)
}

func TestField_EmptyInterface(t *testing.T) {
	type Rel struct {
		Int int
	}

	type Src struct {
		Rel *Rel
	}

	type SrcForce struct {
		Rel *Rel `copystruct:"force"`
	}

	type Dst struct {
		Rel interface{}
	}

	type DstForce struct {
		Rel interface{} `copystruct:"force"`
	}

	rel := &Rel{Int: 1}
	src := &Src{Rel: rel}
	srcForce := &SrcForce{Rel: rel}

	//
	// Without force
	//

	dst := &Dst{}
	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Nil(t, dst.Rel)

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(src))
	assert.Nil(t, dst.Rel)

	//
	// With force
	//

	dstForce := &DstForce{}
	assert.Nil(t, copystruct.Copy(src).To(dstForce))
	assert.Equal(t, src.Rel, dstForce.Rel)

	dstForce = &DstForce{}
	assert.Nil(t, copystruct.Copy(dstForce).From(srcForce))
	assert.Equal(t, srcForce.Rel, dstForce.Rel)
}

func TestField_NullTypes(t *testing.T) {
	type Src struct {
		PQNullTimeValid      pq.NullTime
		PQNullTimeValidPtr   pq.NullTime
		PQNullTimeInvalid    pq.NullTime
		PQNullTimeInvalidPtr pq.NullTime

		NullStringValid      null.String
		NullStringValidPtr   null.String
		NullStringInvalid    null.String
		NullStringInvalidPtr null.String

		SQLNullStringValid      sql.NullString
		SQLNullStringValidPtr   sql.NullString
		SQLNullStringInvalid    sql.NullString
		SQLNullStringInvalidPtr sql.NullString

		SQLNullInt64Valid      sql.NullInt64
		SQLNullInt64ValidPtr   sql.NullInt64
		SQLNullInt64Invalid    sql.NullInt64
		SQLNullInt64InvalidPtr sql.NullInt64

		SQLNullBoolValid      sql.NullBool
		SQLNullBoolValidPtr   sql.NullBool
		SQLNullBoolInvalid    sql.NullBool
		SQLNullBoolInvalidPtr sql.NullBool
	}

	type SrcForce struct {
		PQNullTimeValid      pq.NullTime `copystruct:"force"`
		PQNullTimeValidPtr   pq.NullTime `copystruct:"force"`
		PQNullTimeInvalid    pq.NullTime `copystruct:"force"`
		PQNullTimeInvalidPtr pq.NullTime `copystruct:"force"`

		NullStringValid      null.String `copystruct:"force"`
		NullStringValidPtr   null.String `copystruct:"force"`
		NullStringInvalid    null.String `copystruct:"force"`
		NullStringInvalidPtr null.String `copystruct:"force"`

		SQLNullStringValid      sql.NullString `copystruct:"force"`
		SQLNullStringValidPtr   sql.NullString `copystruct:"force"`
		SQLNullStringInvalid    sql.NullString `copystruct:"force"`
		SQLNullStringInvalidPtr sql.NullString `copystruct:"force"`

		SQLNullInt64Valid      sql.NullInt64 `copystruct:"force"`
		SQLNullInt64ValidPtr   sql.NullInt64 `copystruct:"force"`
		SQLNullInt64Invalid    sql.NullInt64 `copystruct:"force"`
		SQLNullInt64InvalidPtr sql.NullInt64 `copystruct:"force"`

		SQLNullBoolValid      sql.NullBool `copystruct:"force"`
		SQLNullBoolValidPtr   sql.NullBool `copystruct:"force"`
		SQLNullBoolInvalid    sql.NullBool `copystruct:"force"`
		SQLNullBoolInvalidPtr sql.NullBool `copystruct:"force"`
	}

	type Dst struct {
		PQNullTimeValid      time.Time
		PQNullTimeValidPtr   *time.Time
		PQNullTimeInvalid    time.Time
		PQNullTimeInvalidPtr *time.Time

		NullStringValid      string
		NullStringValidPtr   *string
		NullStringInvalid    string
		NullStringInvalidPtr *string

		SQLNullStringValid      string
		SQLNullStringValidPtr   *string
		SQLNullStringInvalid    string
		SQLNullStringInvalidPtr *string

		SQLNullInt64Valid      int64
		SQLNullInt64ValidPtr   *int64
		SQLNullInt64Invalid    int64
		SQLNullInt64InvalidPtr *int64

		SQLNullBoolValid      bool
		SQLNullBoolValidPtr   *bool
		SQLNullBoolInvalid    bool
		SQLNullBoolInvalidPtr *bool
	}

	type DstForce struct {
		PQNullTimeValid      time.Time  `copystruct:"force"`
		PQNullTimeValidPtr   *time.Time `copystruct:"force"`
		PQNullTimeInvalid    time.Time  `copystruct:"force"`
		PQNullTimeInvalidPtr *time.Time `copystruct:"force"`

		NullStringValid      string  `copystruct:"force"`
		NullStringValidPtr   *string `copystruct:"force"`
		NullStringInvalid    string  `copystruct:"force"`
		NullStringInvalidPtr *string `copystruct:"force"`

		SQLNullStringValid      string  `copystruct:"force"`
		SQLNullStringValidPtr   *string `copystruct:"force"`
		SQLNullStringInvalid    string  `copystruct:"force"`
		SQLNullStringInvalidPtr *string `copystruct:"force"`

		SQLNullInt64Valid      int64  `copystruct:"force"`
		SQLNullInt64ValidPtr   *int64 `copystruct:"force"`
		SQLNullInt64Invalid    int64  `copystruct:"force"`
		SQLNullInt64InvalidPtr *int64 `copystruct:"force"`

		SQLNullBoolValid      bool  `copystruct:"force"`
		SQLNullBoolValidPtr   *bool `copystruct:"force"`
		SQLNullBoolInvalid    bool  `copystruct:"force"`
		SQLNullBoolInvalidPtr *bool `copystruct:"force"`
	}

	src := &Src{
		PQNullTimeValid:      pq.NullTime{Valid: true, Time: time.Now()},
		PQNullTimeValidPtr:   pq.NullTime{Valid: true, Time: time.Now()},
		PQNullTimeInvalid:    pq.NullTime{Valid: false, Time: time.Now()},
		PQNullTimeInvalidPtr: pq.NullTime{Valid: false, Time: time.Now()},

		NullStringValid:      null.NewString("hello", true),
		NullStringValidPtr:   null.NewString("hello", true),
		NullStringInvalid:    null.NewString("hello", false),
		NullStringInvalidPtr: null.NewString("hello", false),

		SQLNullStringValid:      sql.NullString{Valid: true, String: "hello"},
		SQLNullStringValidPtr:   sql.NullString{Valid: true, String: "hello"},
		SQLNullStringInvalid:    sql.NullString{Valid: false, String: "hello"},
		SQLNullStringInvalidPtr: sql.NullString{Valid: false, String: "hello"},

		SQLNullInt64Valid:      sql.NullInt64{Valid: true, Int64: 1},
		SQLNullInt64ValidPtr:   sql.NullInt64{Valid: true, Int64: 1},
		SQLNullInt64Invalid:    sql.NullInt64{Valid: false, Int64: 1},
		SQLNullInt64InvalidPtr: sql.NullInt64{Valid: false, Int64: 1},

		SQLNullBoolValid:      sql.NullBool{Valid: true, Bool: true},
		SQLNullBoolValidPtr:   sql.NullBool{Valid: true, Bool: true},
		SQLNullBoolInvalid:    sql.NullBool{Valid: false, Bool: true},
		SQLNullBoolInvalidPtr: sql.NullBool{Valid: false, Bool: true},
	}

	srcForce := &SrcForce{
		PQNullTimeValid:      pq.NullTime{Valid: true, Time: time.Now()},
		PQNullTimeValidPtr:   pq.NullTime{Valid: true, Time: time.Now()},
		PQNullTimeInvalid:    pq.NullTime{Valid: false, Time: time.Now()},
		PQNullTimeInvalidPtr: pq.NullTime{Valid: false, Time: time.Now()},

		NullStringValid:      null.NewString("hello", true),
		NullStringValidPtr:   null.NewString("hello", true),
		NullStringInvalid:    null.NewString("hello", false),
		NullStringInvalidPtr: null.NewString("hello", false),

		SQLNullStringValid:      sql.NullString{Valid: true, String: "hello"},
		SQLNullStringValidPtr:   sql.NullString{Valid: true, String: "hello"},
		SQLNullStringInvalid:    sql.NullString{Valid: false, String: "hello"},
		SQLNullStringInvalidPtr: sql.NullString{Valid: false, String: "hello"},

		SQLNullInt64Valid:      sql.NullInt64{Valid: true, Int64: 1},
		SQLNullInt64ValidPtr:   sql.NullInt64{Valid: true, Int64: 1},
		SQLNullInt64Invalid:    sql.NullInt64{Valid: false, Int64: 1},
		SQLNullInt64InvalidPtr: sql.NullInt64{Valid: false, Int64: 1},

		SQLNullBoolValid:      sql.NullBool{Valid: true, Bool: true},
		SQLNullBoolValidPtr:   sql.NullBool{Valid: true, Bool: true},
		SQLNullBoolInvalid:    sql.NullBool{Valid: false, Bool: true},
		SQLNullBoolInvalidPtr: sql.NullBool{Valid: false, Bool: true},
	}

	//
	// Without force
	//

	dst := &Dst{}

	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Zero(t, dst.PQNullTimeValid)
	assert.Nil(t, dst.PQNullTimeValidPtr)
	assert.Zero(t, dst.PQNullTimeInvalid)
	assert.Nil(t, dst.PQNullTimeInvalidPtr)

	assert.Zero(t, dst.NullStringValid)
	assert.Nil(t, dst.NullStringValidPtr)
	assert.Zero(t, dst.NullStringInvalid)
	assert.Nil(t, dst.NullStringInvalidPtr)

	assert.Zero(t, dst.SQLNullStringValid)
	assert.Nil(t, dst.SQLNullStringValidPtr)
	assert.Zero(t, dst.SQLNullStringInvalid)
	assert.Nil(t, dst.SQLNullStringInvalidPtr)

	assert.Zero(t, dst.SQLNullInt64Valid)
	assert.Nil(t, dst.SQLNullInt64ValidPtr)
	assert.Zero(t, dst.SQLNullInt64Invalid)
	assert.Nil(t, dst.SQLNullInt64InvalidPtr)

	assert.Zero(t, dst.SQLNullBoolValid)
	assert.Nil(t, dst.SQLNullBoolValidPtr)
	assert.Zero(t, dst.SQLNullBoolInvalid)
	assert.Nil(t, dst.SQLNullBoolInvalidPtr)

	//
	// With force
	//

	dstForce := &DstForce{}
	assert.Nil(t, copystruct.Copy(srcForce).To(dstForce))

	assert.Equal(t, srcForce.PQNullTimeValid.Time, dstForce.PQNullTimeValid)
	assert.NotNil(t, dstForce.PQNullTimeValidPtr)
	assert.Equal(t, srcForce.PQNullTimeValidPtr.Time, *dstForce.PQNullTimeValidPtr)
	assert.Zero(t, dstForce.PQNullTimeInvalid)
	assert.Nil(t, dstForce.PQNullTimeInvalidPtr)

	assert.Equal(t, srcForce.NullStringValid.String, dstForce.NullStringValid)
	assert.NotNil(t, dstForce.NullStringValidPtr)
	assert.Equal(t, srcForce.NullStringValidPtr.String, *dstForce.NullStringValidPtr)
	assert.Zero(t, dstForce.NullStringInvalid)
	assert.Nil(t, dstForce.NullStringInvalidPtr)

	assert.Equal(t, srcForce.SQLNullStringValid.String, dstForce.SQLNullStringValid)
	assert.NotNil(t, dstForce.SQLNullStringValidPtr)
	assert.Equal(t, srcForce.SQLNullStringValidPtr.String, *dstForce.SQLNullStringValidPtr)
	assert.Zero(t, dstForce.SQLNullStringInvalid)
	assert.Nil(t, dstForce.SQLNullStringInvalidPtr)

	assert.Equal(t, srcForce.SQLNullInt64Valid.Int64, dstForce.SQLNullInt64Valid)
	assert.NotNil(t, dstForce.SQLNullInt64ValidPtr)
	assert.Equal(t, srcForce.SQLNullInt64ValidPtr.Int64, *dstForce.SQLNullInt64ValidPtr)
	assert.Zero(t, dstForce.SQLNullInt64Invalid)
	assert.Nil(t, dstForce.SQLNullInt64InvalidPtr)

	assert.Equal(t, srcForce.SQLNullBoolValid.Bool, dstForce.SQLNullBoolValid)
	assert.NotNil(t, dstForce.SQLNullBoolValidPtr)
	assert.Equal(t, srcForce.SQLNullBoolValidPtr.Bool, *dstForce.SQLNullBoolValidPtr)
	assert.Zero(t, dstForce.SQLNullBoolInvalid)
	assert.Nil(t, dstForce.SQLNullBoolInvalidPtr)
}

func TestField_SameNameWithDifferentType(t *testing.T) {
	type FooInt struct {
		Foo int
	}

	type FooStr struct {
		Foo string
	}

	//
	// To()
	//

	srcInt := &FooInt{Foo: 1}
	dstStr := &FooStr{}

	assert.Nil(t, copystruct.Copy(dstStr).From(srcInt))
	bytes, _ := json.Marshal(dstStr)
	fmt.Println(string(bytes))
	assert.Empty(t, dstStr.Foo)

	//
	// From()
	//

	dstStr = &FooStr{}
	assert.Nil(t, copystruct.Copy(dstStr).From(srcInt))
	assert.Empty(t, dstStr.Foo)
}

func TestMethod(t *testing.T) {
	c := M{"message": "hello"}
	src := &MethodTesterFoo{TagFirst: "field-value"}
	dst := &MethodTesterBar{}

	//
	// To()
	//

	assert.Nil(t, copystruct.Copy(src).WithContext(c).To(dst))
	assert.Equal(t, c, dst.FooContext)
	assert.Equal(t, MethodTesterFoo{}.FooInteger(), dst.FooInteger)
	assert.Empty(t, dst.FooSkipped)
	assert.Equal(t, "method-value", dst.TagFirst)

	assert.Equal(t, MethodTesterFoo{}.FooSliceToSlicePtr(), *dst.FooSliceToSlicePtr)
	assert.Equal(t, *MethodTesterFoo{}.FooSlicePtrToSlice(), dst.FooSlicePtrToSlice)

	assert.Equal(t, MethodTesterFoo{}.FooStringToStringPtr(), *dst.FooStringToStringPtr)
	assert.Equal(t, *MethodTesterFoo{}.FooStringPtrToString(), dst.FooStringPtrToString)

	assert.Equal(t, MethodTesterFoo{}.FooMapToMapPtr(), *dst.FooMapToMapPtr)
	assert.Equal(t, *MethodTesterFoo{}.FooMapPtrToMap(), dst.FooMapPtrToMap)

	//
	// From()
	//

	dst = &MethodTesterBar{}
	assert.Nil(t, copystruct.Copy(dst).WithContext(c).From(src))
	assert.Equal(t, c, dst.FooContext)
	assert.Equal(t, MethodTesterFoo{}.FooInteger(), dst.FooInteger)
	assert.Empty(t, dst.FooSkipped)
	assert.Equal(t, "method-value", dst.TagFirst)

	assert.Equal(t, MethodTesterFoo{}.FooSliceToSlicePtr(), *dst.FooSliceToSlicePtr)
	assert.Equal(t, *MethodTesterFoo{}.FooSlicePtrToSlice(), dst.FooSlicePtrToSlice)

	assert.Equal(t, MethodTesterFoo{}.FooStringToStringPtr(), *dst.FooStringToStringPtr)
	assert.Equal(t, *MethodTesterFoo{}.FooStringPtrToString(), dst.FooStringPtrToString)

	assert.Equal(t, MethodTesterFoo{}.FooMapToMapPtr(), *dst.FooMapToMapPtr)
	assert.Equal(t, *MethodTesterFoo{}.FooMapPtrToMap(), dst.FooMapPtrToMap)
}

func TestAnonymousStruct(t *testing.T) {
	type Embedded struct{ Int int }
	type EmbeddedRenamedField struct {
		MyInt int `copystruct:"field:Int"`
	}

	type Src struct{ Embedded }
	type SrcRenamedField struct{ EmbeddedRenamedField }

	type Dst struct{ Int int }
	type DstRenamedField struct {
		MyInt int `copystruct:"field:Int"`
	}

	embedded := Embedded{Int: 1}
	embeddedRenamedField := EmbeddedRenamedField{MyInt: 1}
	src := &Src{Embedded: embedded}
	srcRenamedField := &SrcRenamedField{EmbeddedRenamedField: embeddedRenamedField}

	//
	// To()
	//

	dst := &Dst{}
	assert.Nil(t, copystruct.Copy(src).To(dst))
	assert.Equal(t, src.Int, dst.Int)

	dstRenamedField := &DstRenamedField{}
	assert.Nil(t, copystruct.Copy(src).To(dstRenamedField))
	assert.Equal(t, src.Int, dstRenamedField.MyInt)

	//
	// From()
	//

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(src))
	assert.Equal(t, src.Int, dst.Int)

	dst = &Dst{}
	assert.Nil(t, copystruct.Copy(dst).From(srcRenamedField))
	assert.Equal(t, srcRenamedField.MyInt, dst.Int)
}

func TestNullableType(t *testing.T) {
	type Value struct {
		UUID uuid.UUID
	}

	type Ptr struct {
		UUID *uuid.UUID
	}

	type ToString struct {
		UUID uuid.UUID `copystruct:"force"`
	}

	type PtrToString struct {
		UUID *uuid.UUID `copystruct:"force"`
	}

	type FromNullable struct {
		UUID string `copystruct:"force"`
	}

	type PtrFromNullable struct {
		UUID *string `copystruct:"force"`
	}

	// Same type: value -- copy to
	{
		src := &Value{UUID: uuid.NewV4()}
		dst := &Value{}
		assert.Nil(t, copystruct.Copy(src).To(dst))
		assert.Equal(t, src.UUID, dst.UUID)
	}

	// Same type: value -- copy from
	{
		src := &Value{}
		from := &Value{UUID: uuid.NewV4()}
		assert.Nil(t, copystruct.Copy(src).From(from))
		assert.Equal(t, from.UUID, src.UUID)
	}

	// Same type: pointer -- copy to
	{
		uid := uuid.NewV4()
		src := &Ptr{UUID: &uid}
		dst := &Ptr{}
		assert.Nil(t, copystruct.Copy(src).To(dst))
		assert.Equal(t, src.UUID, dst.UUID)
	}

	// Same type: pointer -- copy from
	{
		uid := uuid.NewV4()
		src := &Ptr{}
		from := &Ptr{UUID: &uid}
		assert.Nil(t, copystruct.Copy(src).From(from))
		assert.Equal(t, from.UUID, src.UUID)
	}

	// Value to value -- copy to
	{
		src := &Value{UUID: uuid.NewV4()}
		dst := &FromNullable{}
		assert.Nil(t, copystruct.Copy(src).To(dst))
		assert.Equal(t, src.UUID.String(), dst.UUID)
	}

	// Value to value -- copy from
	{
		src := &FromNullable{}
		from := &ToString{UUID: uuid.NewV4()}
		assert.Nil(t, copystruct.Copy(src).From(from))
		assert.Equal(t, from.UUID.String(), src.UUID)
	}

	// Value to pointer -- copy to
	{
		src := &ToString{UUID: uuid.NewV4()}
		dst := &PtrFromNullable{}
		assert.Nil(t, copystruct.Copy(src).To(dst))
		assert.Equal(t, src.UUID.String(), *dst.UUID)
	}

	// Value to pointer -- copy from
	{
		src := &PtrFromNullable{}
		from := &ToString{UUID: uuid.NewV4()}
		assert.Nil(t, copystruct.Copy(src).From(from))
		assert.Equal(t, from.UUID.String(), *src.UUID)
	}

	// Pointer to value -- copy to
	{
		uid := uuid.NewV4()
		src := &PtrToString{UUID: &uid}
		dst := &FromNullable{}
		assert.Nil(t, copystruct.Copy(src).To(dst))
		assert.Equal(t, src.UUID.String(), dst.UUID)
	}

	// Pointer to value -- copy from
	{
		uid := uuid.NewV4()
		src := &FromNullable{}
		from := &PtrToString{UUID: &uid}
		assert.Nil(t, copystruct.Copy(src).From(from))
		assert.Equal(t, from.UUID.String(), src.UUID)
	}
}

// ----------------------------------------------------------------------------
// Method testers
// ----------------------------------------------------------------------------

type MethodTesterFoo struct {
	BarInteger int
	BarContext M      `copystruct:"context"`
	BarSkipped string `copystruct:"skip"`
	TagFirst   string `copystruct:"field:GetTagFirst"`
}

func (MethodTesterFoo) FooInteger() int { return 1 }

func (MethodTesterFoo) FooContext(c M) M { return c }

func (MethodTesterFoo) FooSkipped() string { return "skipped" }

func (MethodTesterFoo) GetTagFirst() string { return "method-value" }

func (MethodTesterFoo) FooSliceToSlicePtr() []string { return []string{"hello"} }

func (MethodTesterFoo) FooSlicePtrToSlice() *[]string { return &[]string{"hello"} }

func (MethodTesterFoo) FooStringToStringPtr() string { return "hello" }

func (MethodTesterFoo) FooStringPtrToString() *string { s := "hello"; return &s }

func (MethodTesterFoo) FooMapToMapPtr() M {
	return M{"one": 1}
}

func (MethodTesterFoo) FooMapPtrToMap() *M {
	return &M{"one": 1}
}

type MethodTesterBar struct {
	FooInteger           int
	FooContext           M         `copystruct:"context"`
	FooSkipped           string    `copystruct:"skip"`
	TagFirst             string    `copystruct:"field:GetTagFirst"`
	FooSliceToSlicePtr   *[]string `copystruct:"force"`
	FooSlicePtrToSlice   []string  `copystruct:"force"`
	FooStringToStringPtr *string   `copystruct:"force"`
	FooStringPtrToString string    `copystruct:"force"`
	FooMapToMapPtr       *M        `copystruct:"force"`
	FooMapPtrToMap       M         `copystruct:"force"`
}

func (MethodTesterBar) BarInteger() int { return 1 }

func (MethodTesterBar) BarContext(c M) M { return c }

func (MethodTesterBar) BarSkipped() string { return "skipped" }

func (MethodTesterBar) GetTagFirst() string { return "method-value" }
