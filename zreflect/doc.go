/*
Package zreflect provides enhanced reflection capabilities for Go programs.

This package extends Go's standard reflect package with additional utilities
and safer access to reflection features. It includes tools for working with
unexported fields, iterating through struct fields, and simplifying common
reflection operations.

Key features:

  - Access to unexported struct fields (with caution)
  - Simplified iteration over struct fields and methods
  - Enhanced type and value handling
  - Utilities for working with struct tags
  - Performance optimizations for reflection operations

Example usage for iterating through struct fields:

	type Person struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Address struct {
			City  string `json:"city"`
			State string `json:"state"`
		} `json:"address"`
	}

	person := Person{Name: "John", Age: 30}
	
	// Iterate through all fields in the struct
	err := zreflect.ForEachValue(reflect.ValueOf(person), func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error {
		fmt.Printf("Field: %s, Tag: %s, Value: %v\n", field.Name, tag, val.Interface())
		return nil
	})

	// Access unexported fields (use with caution)
	val, err := zreflect.GetUnexportedField(reflect.ValueOf(someStruct), "privateField")

	// Set unexported fields (use with caution)
	err = zreflect.SetUnexportedField(reflect.ValueOf(someStruct), "privateField", newValue)
*/
package zreflect
