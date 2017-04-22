package webui

import (
	"github.com/go-macaron/binding"
	"strings"
)

func init() {
	// UIntegerNotZero
	binding.AddRule(&binding.Rule{
		IsMatch: func(rule string) bool {
			return strings.HasPrefix(rule, "IntegerNotZero")
		},
		IsValid: func(errs binding.Errors, name string, v interface{}) (bool, binding.Errors) {
			num, ok := v.(int)
			if !ok {
				return false, errs
			}
			if num < 0 {
				errs.Add([]string{name}, "IntegerNotZero", "Value should be greater than zero")
				return false, errs
			}
			return true, errs
		},
	})
	//UIntegerAndLessOne
	binding.AddRule(&binding.Rule{
		IsMatch: func(rule string) bool {
			return strings.HasPrefix(rule, "UIntegerAndLessOne")
		},
		IsValid: func(errs binding.Errors, name string, v interface{}) (bool, binding.Errors) {
			num, ok := v.(int)
			if !ok {
				return false, errs
			}
			if num < -1 {
				errs.Add([]string{name}, "UIntegerAndLessOne", "Value should be a positive integer or -1")
				return false, errs
			}
			return true, errs
		},
	})

}
