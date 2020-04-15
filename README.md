# gor

[![Travis CI](https://img.shields.io/travis/bingoohuang/gor/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/gor)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/gor/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/gor)
[![Coverage Status](http://codecov.io/github/bingoohuang/gor/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/gor?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/gor)](https://www.goreportcard.com/report/github.com/bingoohuang/gor)

go reflect collections.

This repository contains a bunch of examples for dealing with the `reflect` package.
Mainly, for decoding/encoding stuff, and calling functions dynamically.  
Most of the examples were taken from projects I worked on in the past, and some from projects I am currently working on.  

You will also find informative comments in the examples, that will help you to understand the code.
some of them are mine, and some of them were taken from the godoc website.

If you want to contribute to this repository, don't hesitate to create a PR.

The awesome gopher in the logo was taken from [@egonelbre/gophers](https://github.com/egonelbre/gophers).


### Table Of Content

- [Read struct tags](examples/read_struct_tags_test.go)
- [Get and set struct fields](examples/get_set_struct_fields_test.go)
- [Fill slice with values](examples/fill_slice_string_test.go)
- [Set a value of a number](examples/set_num_test.go)
- [Decode key-value pairs into map](examples/kvstring2map_test.go)
- [Decode key-value pairs into struct](examples/kvstring2struct_test.go)
- [Encode struct into key-value pairs](examples/struct2kvstring_test.go)
- [Check if the underlying type implements an interface](examples/type_impl_interface_test.go)
- [Wrap a `reflect.Value` with pointer (`T` => `*T`)](examples/wrap_with_pointer_test.go)
- Function calls

  - [Call to a method without prameters, and without return value](examples/function_call_test.go)
  - [Call to a function with list of arguments, and validate return values](examples/function_call_args_test.go)
  - [Call to a function dynamically. similar to the template/text package](examples/function_call_dynamic_test.go)
  - [Call to a function with variadic parameter](examples/function_call_varargs_test.go)
  - [Create function at runtime](examples/function_create_test.go)

- [Deep copy struct](examples/deepcopy_test.go)
- [Fore export private](examples/forceexport_test.go)
- [Getting and setting fields, Listing fields/methods](examples/reflector_test.go)
- [reflect walk](examples/reflectwalk_test.go)
- [remove read-only restrictions](examples/sudo_test.go)
- [reflect2 set](examples/set_test.go)
- [editing struct's fields during runtime and mapping structs to other structs](https://github.com/Ompluscator/dynamic-struct)

    ```go
    instance := dynamicstruct.NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	data := []byte(`
    {
        "int": 123,
        "someText": "example",
        "double": 123.45,
        "Boolean": true,
        "Slice": [1, 2, 3],
        "Anonymous": "avoid to read"
    }
    `)
    ```

### Resources

1. [Golang reflector](https://github.com/tkrajina/go-reflector)
1. [Copier for golang, copy value from struct to struct and more](https://github.com/jinzhu/copier)
1. [DeepCopy makes deep copies of things: unexported field values are not copied](https://github.com/mohae/deepcopy)
1. [Golang package for editing struct's fields during runtime and mapping structs to other structs.](https://github.com/Ompluscator/dynamic-struct)
1. [monkey.Patch\(\<target function\>, \<replacement function\>\) to replace a function.](https://github.com/bouk/monkey)

