package cas

import "bytes"
import "math/big"

// A sequence of Expressions to be multiplied together
type Times struct {
	Multiplicands []Ex
}

func (this *Times) ContainsFloat() bool {
	res := false
	for _, e := range this.Multiplicands {
		_, isfloat := e.(*Flt)
		res = res || isfloat
	}
	return res
}

func (m *Times) Eval(es *EvalState) Ex {
	// Start by evaluating each multiplicand
	for i := range m.Multiplicands {
		m.Multiplicands[i] = m.Multiplicands[i].Eval(es)
	}

	// If any of the multiplicands are also Times, merge them with m and remove them
	// We can easily remove an item by replacing it with a one int.
	for i, e := range m.Multiplicands {
		submul, ismul := e.(*Times)
		if ismul {
			m.Multiplicands = append(m.Multiplicands, submul.Multiplicands...)
			m.Multiplicands[i] = &Integer{big.NewInt(1)}
		}
	}

	// If this expression contains any floats, convert everything possible to
	// a float
	if m.ContainsFloat() {
		for i, e := range m.Multiplicands {
			subint, isint := e.(*Integer)
			if isint {
				newfloat := big.NewFloat(0)
				newfloat.SetInt(subint.Val)
				m.Multiplicands[i] = &Flt{newfloat}
			}
		}
	}

	// If there is a zero in the expression, return zero
	for _, e := range m.Multiplicands {
		f, ok := e.(*Flt)
		if ok {
			if f.Val.Cmp(big.NewFloat(0)) == 0 {
				return &Flt{big.NewFloat(0)}
			}
		}
	}

	// Geometrically accumulate floating point values towards the end of the expression
	var lastf *Flt = nil
	for _, e := range m.Multiplicands {
		f, ok := e.(*Flt)
		if ok {
			if lastf != nil {
				f.Val.Mul(f.Val, lastf.Val)
				lastf.Val = big.NewFloat(1)
			}
			lastf = f
		}
	}

	// Remove one Floats
	for i := len(m.Multiplicands)-1; i >= 0; i-- {
		f, ok := m.Multiplicands[i].(*Flt)
		if ok && f.Val.Cmp(big.NewFloat(1)) == 0 {
			m.Multiplicands[i] = m.Multiplicands[len(m.Multiplicands)-1]
			m.Multiplicands[len(m.Multiplicands)-1] = nil
			m.Multiplicands = m.Multiplicands[:len(m.Multiplicands)-1]
		}
	}

	// Geometrically accumulate integer values towards the end of the expression
	var lasti *Integer = nil
	for _, e := range m.Multiplicands {
		theint, ok := e.(*Integer)
		if ok {
			if lasti != nil {
				theint.Val.Mul(theint.Val, lasti.Val)
				lasti.Val = big.NewInt(1)
			}
			lasti = theint
		}
	}

	// Remove one Integers
	for i := len(m.Multiplicands)-1; i >= 0; i-- {
		theint, ok := m.Multiplicands[i].(*Integer)
		if ok && theint.Val.Cmp(big.NewInt(1)) == 0 {
			m.Multiplicands[i] = m.Multiplicands[len(m.Multiplicands)-1]
			m.Multiplicands[len(m.Multiplicands)-1] = nil
			m.Multiplicands = m.Multiplicands[:len(m.Multiplicands)-1]
		}
	}

	// If one expression remains, replace this Times with the expression
	if len(m.Multiplicands) == 1 {
		return m.Multiplicands[0]
	}

	return m
}

func (m *Times) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i, e := range m.Multiplicands {
		buffer.WriteString(e.ToString())
		if i != len(m.Multiplicands)-1 {
			buffer.WriteString(" * ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

func (this *Times) IsEqual(otherEx Ex, es *EvalState) string {
	thisEx := this.Eval(es)
	otherEx = otherEx.Eval(es)
	this, ok := thisEx.(*Times)
	if !ok {
		return thisEx.IsEqual(otherEx, es)
	}
	other, ok := otherEx.(*Times)
	if !ok {
		return "EQUAL_FALSE"
	}
	return CommutativeIsEqual(this.Multiplicands, other.Multiplicands, es)
}

func (this *Times) DeepCopy() Ex {
	var thiscopy = &Times{}
	for i := range this.Multiplicands {
		thiscopy.Multiplicands = append(thiscopy.Multiplicands, this.Multiplicands[i].DeepCopy())
	}
	return thiscopy
}