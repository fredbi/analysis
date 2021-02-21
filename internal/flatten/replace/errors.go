package replace

import "fmt"

// ErrReplace is a class of errors returned by the replace package
type ErrReplace string

func (e ErrReplace) Error() string {
	return string(e)
}

func asErr(format string, args ...interface{}) ErrReplace {
	return ErrReplace(fmt.Sprintf(format, args...))
}

func fromErr(err error) ErrReplace {
	return ErrReplace(err.Error())
}

// ErrNoSchema expected a schema under $ref
func ErrNoSchema(key string, value interface{}) ErrReplace {
	return asErr("no schema with ref found at %s for %T", key, value)
}

// ErrNotANumber expected a number in pointer to reference an item in array
func ErrNotANumber(key string, err error) ErrReplace {
	return fromErr(fmt.Errorf("%s not a number: %w", key, err))
}

// ErrUnhandledType unexpected type found
func ErrUnhandledType(key string, value interface{}) ErrReplace {
	return asErr("unhandled parent schema rewrite %s (%T)", key, value)
}

// ErrCyclicalPointer cannot resolve a cyclic chain of pointers
func ErrCyclicalPointer(key string) ErrReplace {
	return asErr("cannot resolve cyclic chain of pointers under %s", key)
}

// ErrParentResolve pointer to parent could not be resolved
func ErrParentResolve(key string, err error) ErrReplace {
	return fromErr(fmt.Errorf("can't get parent for %s: %w", key, err))
}
