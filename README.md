<p align="center">
<img 
    src="img/logo.png"
    width="500" alt="Golang reflect package examples">
<br/><br/>
</p>

This repository contains a bunch of examples for dealing with the `reflect` package.
Mainly, for decoding/encoding stuff, and calling functions dynamically.  
Most of the examples were taken from projects I worked on in the past, and some from projects I am currently working on.  

You will also find informative comments in the examples, that will help you to understand the code.
some of them are mine, and some of them were taken from the godoc website.

If you want to contribute to this repository, don't hesitate to create a PR.

The awesome gopher in the logo was taken from [@egonelbre/gophers](https://github.com/egonelbre/gophers).


### Table Of Content
- [Read struct tags](read_struct_tags_test.go)
- [Get and set struct fields](get_set_struct_fields_test.go)
- [Fill slice with values](fill_slice_string_test.go)
- [Set a value of a number](set_num_test.go)
- [Decode key-value pairs into map](kvstring2map_test.go)
- [Decode key-value pairs into struct](kvstring2struct_test.go)
- [Encode struct into key-value pairs](struct2kvstring_test.go)
- [Check if the underlying type implements an interface](type_impl_interface_test.go)
- [Wrap a `reflect.Value` with pointer (`T` => `*T`)](wrap_with_pointer_test.go)
- Function calls
  - [Call to a method without prameters, and without return value](function_call_test.go)
  - [Call to a function with list of arguments, and validate return values](function_call_args_test.go)
  - [Call to a function dynamically. similar to the template/text package](function_call_dynamic_test.go)
  - [Call to a function with variadic parameter](function_call_varargs_test.go)
  - [Create function at runtime](function_create_test.go)
- [Deep copy struct](deepcopy_test.go)
- [Fore export private](forceexport_test.go)
- [Getting and setting fields, Listing fields/methods](reflector_test.go)
- [reflect walk](reflectwalk_test.go)
- [remove read-only restrictions](sudo_test.go)
- [reflect2 set](set_test.go)

