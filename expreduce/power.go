package expreduce

import (
	"github.com/corywalker/mathbigext"
	"math/big"
)

func GetPowerDefinitions() (defs []Definition) {
	defs = append(defs, Definition{
		Name: "Power",
		Attributes: []string{"Listable", "NumericFunction", "OneIdentity"},
		Rules: []Rule{
			// Simplify nested exponents
			{"Power[Power[a_,b_Integer],c_Integer]", "a^(b*c)"},
			{"Power[Power[a_,b_Real],c_Integer]", "a^(b*c)"},

			// Power definitions
			{"Power[Times[Except[_Symbol, first_], inner___], pow_]", "first^pow*Power[Times[inner],pow]"},
			{"Power[Times[first_, inner___], Except[_Symbol, pow_]]", "first^pow*Power[Times[inner],pow]"},
		},
		toString: func(this *Expression, form string) (bool, string) {
			return ToStringInfixAdvanced(this.Parts[1:], "^", false, "", "", form)
		},
		legacyEvalFn: func(this *Expression, es *EvalState) Ex {
			if len(this.Parts) != 3 {
				return this
			}

			// TODO: Handle cases like float raised to the float and things raised to
			// zero and 1

			baseInt, baseIsInt := this.Parts[1].(*Integer)
			powerInt, powerIsInt := this.Parts[2].(*Integer)
			baseFlt, baseIsFlt := this.Parts[1].(*Flt)
			powerFlt, powerIsFlt := this.Parts[2].(*Flt)
			// Anything raised to the 1st power is itself
			if powerIsFlt {
				if powerFlt.Val.Cmp(big.NewFloat(1)) == 0 {
					return this.Parts[1]
				}
			} else if powerIsInt {
				if powerInt.Val.Cmp(big.NewInt(1)) == 0 {
					return this.Parts[1]
				}
			}
			// Anything raised to the 0th power is 1, with a small exception
			isZerothPower := false
			if powerIsFlt {
				if powerFlt.Val.Cmp(big.NewFloat(0)) == 0 {
					isZerothPower = true
				}
			} else if powerIsInt {
				if powerInt.Val.Cmp(big.NewInt(0)) == 0 {
					isZerothPower = true
				}
			}
			isZeroBase := false
			if baseIsFlt {
				if baseFlt.Val.Cmp(big.NewFloat(0)) == 0 {
					isZeroBase = true
				}
			} else if baseIsInt {
				if baseInt.Val.Cmp(big.NewInt(0)) == 0 {
					isZeroBase = true
				}
			}
			if isZerothPower {
				if isZeroBase {
					return &Symbol{"Indeterminate"}
				}
				return &Integer{big.NewInt(1)}
			}

			//es.Debugf("Power eval. baseIsInt=%v, powerIsInt=%v", baseIsInt, powerIsInt)
			// Fully integer Power expression
			if baseIsInt && powerIsInt {
				cmpres := powerInt.Val.Cmp(big.NewInt(0))
				//es.Debugf("Cmpres: %v", cmpres)
				if cmpres == 1 {
					res := big.NewInt(0)
					res.Exp(baseInt.Val, powerInt.Val, nil)
					return &Integer{res}
				} else if cmpres == -1 {
					newbase := big.NewInt(0)
					absPower := big.NewInt(0)
					absPower.Abs(powerInt.Val)
					newbase.Exp(baseInt.Val, absPower, nil)
					if newbase.Cmp(big.NewInt(1)) == 0 {
						return &Integer{big.NewInt(1)}
					}
					if newbase.Cmp(big.NewInt(-1)) == 0 {
						return &Integer{big.NewInt(-1)}
					}
					//return &Expression{[]Ex{&Symbol{"Power"}, &Integer{newbase}, &Integer{big.NewInt(-1)}}}
					return &Expression{[]Ex{&Symbol{"Rational"}, &Integer{big.NewInt(1)}, &Integer{newbase}}}
				} else {
					return &Expression{[]Ex{&Symbol{"Error"}, &String{"Unexpected zero power in Power evaluation."}}}
				}
			}

			if baseIsFlt && powerIsInt {
				return &Flt{mathbigext.Pow(baseFlt.Val, big.NewFloat(0).SetInt(powerInt.Val))}
			}
			if baseIsInt && powerIsFlt {
				return &Flt{mathbigext.Pow(big.NewFloat(0).SetInt(baseInt.Val), powerFlt.Val)}
			}
			if baseIsFlt && powerIsFlt {
				return &Flt{mathbigext.Pow(baseFlt.Val, powerFlt.Val)}
			}

			powerRat, powerIsRat := this.Parts[2].(*Rational)
			if powerIsRat {
				if powerRat.Num.Cmp(big.NewInt(1)) == 0 && powerRat.Den.Cmp(big.NewInt(2)) == 0 {
					return &Expression{[]Ex{
						&Symbol{"Sqrt"},
						this.Parts[1],
					}}
				}
			}

			return this
		},
		Tests: []TestInstruction{
			// Test raising expressions to the first power
			&StringTest{"(1 + x)", "(x+1)^1"},
			&StringTest{"0", "0^1"},
			&StringTest{"0.", "0.^1"},
			&StringTest{"-5", "-5^1"},
			&StringTest{"-5.5", "-5.5^1"},
			&StringTest{"(1 + x)", "(x+1)^1."},
			&StringTest{"0", "0^1."},
			&StringTest{"0.", "0.^1."},
			&StringTest{"-5", "(-5)^1."},
			&StringTest{"-5.5", "-5.5^1."},

			// Test raising expressions to the zero power
			&StringTest{"1", "(x+1)^0"},
			&StringTest{"Indeterminate", "0^0"},
			&StringTest{"Indeterminate", "0.^0"},
			&StringTest{"-1", "-5^0"},
			&StringTest{"1", "(-5)^0"},
			&StringTest{"1", "(-5.5)^0"},
			&StringTest{"1", "(x+1)^0."},
			&StringTest{"Indeterminate", "0^0."},
			&StringTest{"Indeterminate", "0.^0."},
			&StringTest{"-1", "-5^0."},
			&StringTest{"1", "(-5.5)^0."},
			&StringTest{"-1", "-5^0"},
			&StringTest{"1", "99^0"},

			&StringTest{"125", "5^3"},
			&StringTest{"1/125", "5^-3"},
			&StringTest{"-125", "(-5)^3"},
			&StringTest{"-1/125", "(-5)^-3"},

			//&StringTest{"2.975379863266329e+1589", "39^999."},
			//&StringTest{"3.360915398890324e-1590", "39^-999."},
			//&StringTest{"1.9950631168791027e+3010", ".5^-10000."},
			//&StringTest{"1.9950631168791027e+3010", ".5^-10000"},
			&StringTest{"2.97538e+1589", "39^999."},
			&StringTest{"3.36092e-1590", "39^-999."},
			&StringTest{"1.99506e+3010", ".5^-10000."},
			&StringTest{"1.99506e+3010", ".5^-10000"},

			&StringTest{"1", "1^1"},
			&StringTest{"1", "1^2"},
			&StringTest{"1", "1^0"},
			&StringTest{"1", "1^-1"},
			&StringTest{"1", "1^-2"},
			&StringTest{"1", "1^99999992"},
			&StringTest{"1.", "1^2."},
			&StringTest{"1.", "1^99999992."},
			&StringTest{"1.", "1.^30"},
			&StringTest{"4.", "(1.*2*1.)^2"},
			&StringTest{"-1", "(-1)^1"},
			&StringTest{"1", "(-1)^2"},
			&StringTest{"1", "(-1)^0"},
			&StringTest{"1", "(-1)^0"},
			&StringTest{"-1", "(-1)^-1"},
			&StringTest{"1", "(-1)^-2"},
			&StringTest{"1", "(-1)^99999992"},
			&StringTest{"1.", "(-1.)^30"},
			&StringTest{"4.", "(1.*2*-1.)^2"},
			&StringTest{"-0.5", "(1.*2*-1.)^(-1)"},

			&SameTest{"Rational", "Power[2, -1] // Head"},
			&SameTest{"Integer", "Power[1, -1] // Head"},
			&SameTest{"Integer", "Power[2, 2] // Head"},
			&SameTest{"Rational", "Power[-2, -1] // Head"},
			&SameTest{"Rational", "Power[2, -2] // Head"},

			// Exponent simplifications
			&SameTest{"m^4", "m^2*m^2"},
			&SameTest{"m^4", "(m^2)^2"},
			&SameTest{"(m^2)^2.", "(m^2)^2."},
			&SameTest{"(m^2.)^2.", "(m^2.)^2."},
			&SameTest{"m^4.", "(m^2.)^2"},
		},
	})
	defs = append(defs, Definition{
		Name: "PowerExpand",
		Rules: []Rule{
			{"PowerExpand[exp_]", "exp //. {Log[x_ y_]:>Log[x]+Log[y],Log[x_^k_]:>k Log[x]}"},
		},
	})
	return
}