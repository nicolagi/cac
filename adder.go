package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// adder uses leftIn, rightIn, and carryIn as inputs and uses output alarms
// whose names are constructed from the adder's name. The input alarms must
// exist already.
type adder struct {
	cw   *cloudwatch.CloudWatch
	name string
	ha1  *halfAdder
	ha2  *halfAdder

	leftIn  string
	rightIn string
	carryIn string
}

func newAdder(cw *cloudwatch.CloudWatch, name, leftIn, rightIn, carryIn string) *adder {
	a := &adder{
		cw:      cw,
		name:    name,
		leftIn:  leftIn,
		rightIn: rightIn,
		carryIn: carryIn,
	}
	a.ha1 = &halfAdder{
		cw:      a.cw,
		name:    a.ha1Name(),
		leftIn:  a.leftIn,
		rightIn: a.rightIn,
	}
	a.ha2 = &halfAdder{
		cw:      a.cw,
		name:    a.ha2Name(),
		leftIn:  a.ha1.soutName(),
		rightIn: a.carryIn,
	}
	return a
}

func (a *adder) build() error {
	if err := a.ha1.build(); err != nil {
		return err
	}
	if err := a.ha2.build(); err != nil {
		return err
	}
	rule := fmt.Sprintf("ALARM(%q) OR ALARM(%q)", a.ha1.coutName(), a.ha2.coutName())
	return pca(a.cw, a.coutName(), rule)
}

func (a *adder) ha1Name() string {
	return fmt.Sprintf("ha1:fa:%s", a.name)
}

func (a *adder) ha2Name() string {
	return fmt.Sprintf("ha2:fa:%s", a.name)
}

func (a *adder) coutName() string {
	return fmt.Sprintf("cout:fa:%s", a.name)
}

func (a *adder) soutName() string {
	return a.ha2.soutName()
}

func (a *adder) setMainInputs(leftIn bool, rightIn bool) error {
	err := sas(a.cw, a.leftIn, leftIn)
	if err != nil {
		return err
	}
	return sas(a.cw, a.rightIn, rightIn)
}

func (a *adder) setInputs(leftIn bool, rightIn bool, carryIn bool) error {
	err := a.setMainInputs(leftIn, rightIn)
	if err != nil {
		return err
	}
	return sas(a.cw, a.carryIn, carryIn)
}

func (a *adder) readOutputs() (carry bool, sum bool, err error) {
	states, err := describeStates(a.cw, []string{
		a.coutName(),
		a.soutName(),
	})
	if err != nil {
		return false, false, err
	}
	carry = cloudwatch.StateValueAlarm == states[0]
	sum = cloudwatch.StateValueAlarm == states[1]
	return
}
