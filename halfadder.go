package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// halfAdder uses leftIn and rightIn as inputs and uses output alarms whose names
// are constructed from the half-adder's name. The input alarms must exist
// already.
type halfAdder struct {
	cw   *cloudwatch.CloudWatch
	name string

	leftIn  string
	rightIn string
}

func (ha *halfAdder) build() error {
	rule := fmt.Sprintf("ALARM(%q) AND ALARM(%q)", ha.leftIn, ha.rightIn)
	err := pca(ha.cw, ha.coutName(), rule)
	if err != nil {
		return fmt.Errorf("could not create %q: %w", ha.coutName(), err)
	}
	rule = fmt.Sprintf("(ALARM(%q) OR ALARM(%q)) AND NOT (ALARM(%q) AND ALARM(%q))", ha.leftIn, ha.rightIn, ha.leftIn, ha.rightIn)
	err = pca(ha.cw, ha.soutName(), rule)
	if err != nil {
		return fmt.Errorf("could not create %q: %w", ha.soutName(), err)
	}
	return nil
}

func (ha *halfAdder) coutName() string {
	return fmt.Sprintf("cout:ha:%s", ha.name)
}

func (ha *halfAdder) soutName() string {
	return fmt.Sprintf("sout:ha:%s", ha.name)
}

func (ha *halfAdder) setInputs(leftIn bool, rightIn bool) error {
	err := sas(ha.cw, ha.leftIn, leftIn)
	if err != nil {
		return err
	}
	return sas(ha.cw, ha.rightIn, rightIn)
}

func (ha *halfAdder) readOutputs() (carry bool, sum bool, err error) {
	states, err := describeStates(ha.cw, []string{
		ha.coutName(),
		ha.soutName(),
	})
	if err != nil {
		return false, false, err
	}
	carry = cloudwatch.StateValueAlarm == states[0]
	sum = cloudwatch.StateValueAlarm == states[1]
	return
}
