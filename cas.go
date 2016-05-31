package cas

import "fmt"
import "bytes"

// Ex stands for Expression. Most structs should implement this
type Ex interface {
	Eval() Ex
	ToString() string
}

// Floating point numbers represented by float64
type Float struct {
	Val float64
}

func (f *Float) Eval() Ex {
	return f
}

func (f *Float) ToString() string {
	return fmt.Sprintf("%g", f.Val)
}

// A sequence of Expressions to be added together
type Add struct {
	addends []Ex
}

func (a *Add) Eval() Ex {
	// Start by evaluating each addend
	for i := range a.addends {
		a.addends[i] = a.addends[i].Eval()
	}

	// If any of the addends are also Adds, merge them with a and remove them
	// We can easily remove an item by replacing it with a zero float.
	for i, e := range a.addends {
		subadd, isadd := e.(*Add)
		if isadd {
			a.addends = append(a.addends, subadd.addends...)
			a.addends[i] = &Float{0}
		}
	}

	// Accumulate floating point values towards the end of the expression
	var lastf *Float = nil
	for _, e := range a.addends {
		f, ok := e.(*Float)
		if ok {
			if lastf != nil {
				f.Val += lastf.Val;
				lastf.Val = 0
			}
			lastf = f
		}
	}

	// Remove zero Floats
	for i := len(a.addends)-1; i >= 0; i-- {
		f, ok := a.addends[i].(*Float)
		if ok && f.Val == 0 {
			a.addends[i] = a.addends[len(a.addends)-1]
			a.addends[len(a.addends)-1] = nil
			a.addends = a.addends[:len(a.addends)-1]
		}
	}

	// If one float remains, replace this Add with the Float
	if len(a.addends) == 1 {
		_, isfloat := a.addends[0].(*Float)
		if isfloat {
			return a.addends[0]
		}
	}

	return a
}

func (a *Add) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i, e := range a.addends {
		buffer.WriteString(e.ToString())
		if i != len(a.addends)-1 {
			buffer.WriteString(" + ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

// Variables are defined by a string-based name
type Variable struct {
	Name string
}

func (v *Variable) Eval() Ex {
	return v
}

func (v *Variable) ToString() string {
	return fmt.Sprintf("%v", v.Name)
}

