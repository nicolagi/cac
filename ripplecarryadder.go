package main

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type rippleCarryAdder struct {
	cw     *cloudwatch.CloudWatch
	name   string
	adders [8]*adder
}

func newRippleCarryAdder(cw *cloudwatch.CloudWatch, name string) *rippleCarryAdder {
	rca := &rippleCarryAdder{
		cw:   cw,
		name: name,
	}
	for i := 0; i < 8; i++ {
		rca.adders[i] = newAdder(
			cw,
			rca.adderName(i),
			rca.adderLeftInName(i),
			rca.adderRightInName(i),
			rca.adderCarryInName(i),
		)
	}
	return rca
}

func (rca *rippleCarryAdder) build() error {
	err := pcab(rca.cw, rca.adderCarryInName(0), false)
	if err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		err = pcab(rca.cw, rca.adderLeftInName(i), false)
		if err != nil {
			break
		}
		err = pcab(rca.cw, rca.adderRightInName(i), false)
		if err != nil {
			break
		}
		err = rca.adders[i].build()
		if err != nil {
			break
		}
	}
	return err
}

func (rca *rippleCarryAdder) adderName(i int) string {
	return fmt.Sprintf("adder%d:rca:%s", i, rca.name)
}

func (rca *rippleCarryAdder) adderLeftInName(i int) string {
	return fmt.Sprintf("lin%d:rca:%s", i, rca.name)
}

func (rca *rippleCarryAdder) adderRightInName(i int) string {
	return fmt.Sprintf("rin%d:rca:%s", i, rca.name)
}

func (rca *rippleCarryAdder) adderCarryInName(i int) string {
	if i == 0 {
		return fmt.Sprintf("ground:rca:%s", rca.name)
	}
	return rca.adders[i-1].coutName()
}

func (rca *rippleCarryAdder) soutName(i int) string {
	return rca.adders[i].soutName()
}

func (rca *rippleCarryAdder) overflowName() string {
	return rca.adders[7].coutName()
}

func (rca *rippleCarryAdder) setInputs(leftIn, rightIn register) error {
	var err error
	for i := 0; i < 8; i++ {
		err = rca.adders[i].setMainInputs(leftIn[i], rightIn[i])
		if err != nil {
			break
		}
	}
	return err
}

func (rca *rippleCarryAdder) readOutputs() (sum register, overflow bool, err error) {
	alarmNames := make([]string, 9)
	for i := 0; i < 8; i++ {
		alarmNames[i] = rca.soutName(i)
	}
	alarmNames[8] = rca.overflowName()
	states, err := describeStates(rca.cw, alarmNames)
	if err != nil {
		return sum, false, err
	}
	for i := 0; i < 8; i++ {
		sum[i] = cloudwatch.StateValueAlarm == states[i]
	}
	overflow = cloudwatch.StateValueAlarm == states[8]
	return
}

// saveGraph writes a graph representing the circuit, readable by xdot as a
// diagnostic and demonstration tool.
func (rca *rippleCarryAdder) saveGraph(w io.Writer) error {
	_, _ = fmt.Fprintln(w, "digraph {")
	// Stack of names of alarms of which to get the children, in order to
	// find directed edges. Initially consists of all the 8 output bits and
	// the overflow output bit.
	var stack []string
	for i := 0; i < 8; i++ {
		stack = append(stack, rca.adders[i].soutName())
	}
	stack = append(stack, rca.adders[7].coutName())
	seen := make(map[string]struct{})
	for len(stack) > 0 {
		last := len(stack) - 1
		parentName := stack[last]
		stack = stack[:last]
		childNames, err := children(rca.cw, parentName)
		if err != nil {
			return err
		}
		seen[parentName] = struct{}{}
		for _, cn := range childNames {
			if _, ok := seen[cn]; !ok {
				stack = append(stack, cn)
			}
			_, _ = fmt.Fprintf(w, "\t%q -> %q;\n", cn, parentName)
		}
	}
	_, _ = fmt.Fprintln(w, "}")
	return nil
}

func (rca *rippleCarryAdder) remove() error {
	for i := 0; i < 8; i++ {
		if err := rca.removeFrom(rca.adders[i].leftIn); err != nil {
			return err
		}
		if err := rca.removeFrom(rca.adders[i].rightIn); err != nil {
			return err
		}
		if err := rca.removeFrom(rca.adders[i].carryIn); err != nil {
			return err
		}
	}
	return nil
}

func (rca *rippleCarryAdder) removeFrom(child string) error {
	parentNames, err := parents(rca.cw, child)
	if err != nil {
		return err
	}
	for _, pn := range parentNames {
		if err := rca.removeFrom(pn); err != nil {
			return err
		}
	}
	if err := da(rca.cw, child); err != nil {
		return err
	}
	return nil
}
