# Copystruct

This package is meant to make copying of struct to/from others struct a bit easier.

Original sources are from [ulule/deepcopier](github.com/ulule/deepcopier).

## Usage

```golang
// Deep copy instance1 into instance2
copystruct.Copy(instance1).To(instance2)

// Deep copy instance1 into instance2 and passes the following context (which
// is basically a map[string]interface{}) as first argument
// to methods of instance2 that defined the struct tag "context".
copystruct.Copy(instance1).WithContext(map[string]interface{}{"foo": "bar"}).To(instance2)

// Deep copy instance2 into instance1
copystruct.Copy(instance1).From(instance2)

// Deep copy instance2 into instance1 and passes the following context (which
// is basically a map[string]interface{}) as first argument
// to methods of instance1 that defined the struct tag "context".
copystruct.Copy(instance1).WithContext(map[string]interface{}{"foo": "bar"}).From(instance2)
```

Available options for `deepcopier` struct tag:

| Option    | Description                                                          |
| --------- | -------------------------------------------------------------------- |
| `field`   | Field or method name in source instance                              |
| `skip`    | Ignores the field                                                    |
| `context` | Takes a `map[string]interface{}` as first argument (for methods)     |
| `force`   | Set the value of a `sql.Null*` field (instead of copying the struct) |
| `convert` | convert value types (example between type and its alias)             |

**Options example:**

```golang
type UserName string

type Source struct {
    Name                         UserName
    SkipMe                       string
    SQLNullStringToSQLNullString sql.NullString
    SQLNullStringToString        sql.NullString

}

func (Source) MethodThatTakesContext(c map[string]interface{}) string {
    return "whatever"
}

type Destination struct {
    FieldWithAnotherNameInSource      string         `copystruct:"field:Name;convert"`
    SkipMe                            string         `copystruct:"skip"`
    MethodThatTakesContext            string         `copystruct:"context"`
    SQLNullStringToSQLNullString      sql.NullString 
    SQLNullStringToString             string         `copystruct:"force"`
}

```

Example:

```golang
package main

import (
    "fmt"
 
    "github.com/bingoohuang/goreflect/copystruct"
)

// Model
type User struct {
    // Basic string field
    Name  string
    // copystruct supports https://golang.org/pkg/database/sql/driver/#Valuer
    Email sql.NullString
}

func (u *User) MethodThatTakesContext(ctx map[string]interface{}) string {
    // do whatever you want
    return "hello from this method"
}

// Resource
type UserResource struct {
    DisplayName            string `copystruct:"field:Name"`
    SkipMe                 string `copystruct:"skip"`
    MethodThatTakesContext string `copystruct:"context"`
    Email                  string `copystruct:"force"`

}

func main() {
    user := &User{
        Name: "gilles",
        Email: sql.NullString{
            Valid: true,
            String: "gilles@example.com",
        },
    }

    resource := &UserResource{}

    copystruct.Copy(user).To(resource)

    fmt.Println(resource.DisplayName)
    fmt.Println(resource.Email)
}
```

Looking for more information about the usage?

We wrote [an introduction article](https://github.com/ulule/deepcopier/blob/master/examples/rest-usage/README.rst).
Have a look and feel free to give us your feedback.

