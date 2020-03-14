package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

var (
	profile  string
	region   string
	endpoint string
	headers  string
	verbose  bool

	headerMap = make(map[string]string)
)

func addHeaders(r *request.Request) {
	for k, v := range headerMap {
		r.HTTPRequest.Header.Set(k, v)
	}
}

func defaultClient() (*cloudwatch.CloudWatch, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		return nil, err
	}
	return cloudwatch.New(sess), nil
}

func pca(cw *cloudwatch.CloudWatch, name, rule string) error {
	if verbose {
		log.Printf("Setting %s to %s", name, rule)
	}
	req, _ := cw.PutCompositeAlarmRequest(&cloudwatch.PutCompositeAlarmInput{
		AlarmName: aws.String(name),
		AlarmRule: aws.String(rule),
	})
	addHeaders(req)
	err := req.Send()
	if err != nil {
		return fmt.Errorf("error putting %q with rule %q: %v", name, rule, err)
	}
	return nil
}

func pcab(cw *cloudwatch.CloudWatch, name string, constant bool) error {
	if constant {
		return pca(cw, name, "TRUE")
	}
	return pca(cw, name, "FALSE")
}

func sas(cw *cloudwatch.CloudWatch, name string, constant bool) error {
	if verbose {
		log.Printf("Setting state of %q to %t", name, constant)
	}
	var stateValue string
	if constant {
		stateValue = cloudwatch.StateValueAlarm
	} else {
		stateValue = cloudwatch.StateValueOk
	}
	req, _ := cw.SetAlarmStateRequest(&cloudwatch.SetAlarmStateInput{
		AlarmName:       &name,
		StateValue:      &stateValue,
		StateReason:     aws.String("8-bit adder test"),
		StateReasonData: aws.String("{}"),
	})
	addHeaders(req)
	if err := req.Send(); err != nil {
		return fmt.Errorf("setting alarm state %q to %q: %w", name, stateValue, err)
	}
	return nil
}

func children(cw *cloudwatch.CloudWatch, parentName string) (childNames []string, err error) {
	if verbose {
		log.Printf("Finding children of: %s", parentName)
	}
	req, output := cw.DescribeAlarmsRequest(&cloudwatch.DescribeAlarmsInput{
		ChildrenOfAlarmName: aws.String(parentName),
	})
	addHeaders(req)
	if err := req.Send(); err != nil {
		return nil, err
	}
	// We know the circuits are constructed from composite alarms only, no
	// metric alarms, so we iterate only on the former.
	for _, child := range output.CompositeAlarms {
		childNames = append(childNames, *child.AlarmName)
	}
	return
}

func parents(cw *cloudwatch.CloudWatch, childName string) (parentNames []string, err error) {
	if verbose {
		log.Printf("Finding parents of: %s", childName)
	}
	req, output := cw.DescribeAlarmsRequest(&cloudwatch.DescribeAlarmsInput{
		ParentsOfAlarmName: aws.String(childName),
	})
	addHeaders(req)
	if err := req.Send(); err != nil {
		return nil, err
	}
	// We know the circuits are constructed from composite alarms only, no
	// metric alarms, so we iterate only on the former.
	for _, parent := range output.CompositeAlarms {
		parentNames = append(parentNames, *parent.AlarmName)
	}
	return
}

func da(cw *cloudwatch.CloudWatch, name string) error {
	if verbose {
		log.Printf("Deleting %s", name)
	}
	req, _ := cw.DeleteAlarmsRequest(&cloudwatch.DeleteAlarmsInput{
		AlarmNames: []*string{aws.String(name)},
	})
	addHeaders(req)
	return req.Send()
}

func daRecursive(cw *cloudwatch.CloudWatch, name string) error {
	parentNames, err := parents(cw, name)
	if err != nil {
		return err
	}
	for _, pn := range parentNames {
		if err := daRecursive(cw, pn); err != nil {
			return err
		}
	}
	if err := da(cw, name); err != nil {
		return err
	}
	return nil
}

// describeStates fetches the state value for each composite alarm in the
// input.  It uses the same order in the output as specified in the input
// (something which DescribeAlarms does not do, and I was expect to).
func describeStates(cw *cloudwatch.CloudWatch, alarmNames []string) (states []string, err error) {
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: []*string{
			aws.String(cloudwatch.AlarmTypeCompositeAlarm),
		},
	}
	for _, an := range alarmNames {
		input.AlarmNames = append(input.AlarmNames, aws.String(an))
	}
	req, output := cw.DescribeAlarmsRequest(input)
	addHeaders(req)
	if err := req.Send(); err != nil {
		return nil, err
	}
	if got, want := len(output.CompositeAlarms), len(alarmNames); got != want {
		return nil, fmt.Errorf("got %d composite alarms, want %d", got, want)
	}
	m := make(map[string]string)
	for _, a := range output.CompositeAlarms {
		m[*a.AlarmName] = *a.StateValue
	}
	states = make([]string, len(alarmNames))
	for i, an := range alarmNames {
		states[i] = m[an]
	}
	return states, nil
}

func main() {
	flag.StringVar(&profile, "profile", "computer", "the AWS profile to use for credentials")
	flag.StringVar(&region, "region", "eu-west-1", "the AWS region to create/use alarms in")
	flag.StringVar(&endpoint, "endpoint", "", "the custom `endpoint` if you need to override the default")
	flag.StringVar(&headers, "headers", "", "additional `headers` if required, in the form k1=v1,k2=v2")
	name := flag.String("name", "computer", "the `name` of the circuit")
	build := flag.Bool("build", false, "whether the circuit must be built")
	visualize := flag.Bool("visualize", false, "whether the circuit should be printed in dot format")
	exercise := flag.Bool("exercise", false, "exercise the circuit with one random addition")
	remove := flag.Bool("remove", false, "remove the circuit (all alarms that are part of the circuit)")
	flag.BoolVar(&verbose, "verbose", false, "log diagnostic messages")
	seed := flag.Int64("seed", time.Now().Unix(), "seed for random numbers for exercise - only for reproducibility")
	listRegions := flag.Bool("list-regions", false, "lists the available regions")
	flag.Parse()
	if headers != "" {
		for _, pair := range strings.Split(headers, ",") {
			parts := strings.SplitN(pair, "=", 2)
			key := parts[0]
			value := parts[1]
			if verbose {
				log.Printf("Will use header %q set to %q for all requests", key, value)
			}
			headerMap[key] = value
		}
	}
	cw, err := defaultClient()
	if err != nil {
		log.Fatal(err)
	}
	rca := newRippleCarryAdder(cw, *name)
	if *build {
		err = rca.build()
		if err != nil {
			log.Fatal(err)
		}
	}
	if *visualize {
		err = rca.saveGraph(os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *exercise {
		log.Printf("Using seed %d.", *seed)
		rand.Seed(*seed)
		a := uint8(rand.Intn(256))
		b := uint8(rand.Intn(256))
		log.Printf("Want to add %d and %d and get %d", a, b, a+b)
		log.Print("Setting inputs.")
		aRegister := toRegister(a)
		bRegister := toRegister(b)
		err := rca.setInputs(aRegister, bRegister)
		if err != nil {
			log.Fatalf("Could not set inputs: %v", err)
		}
		log.Print("Sleeping.")
		time.Sleep(time.Second)
		log.Print("Reading outputs.")
		attempts := 1
		log.Printf("  a = %s (%d)", aRegister, a)
		log.Printf("  b = %s (%d)", bRegister, b)
	retry:
		sumRegister, overflow, err := rca.readOutputs()
		if err != nil {
			log.Fatalf("Could not read outputs: %v", err)
		}
		if overflow {
			log.Printf("WARNING: The computation overflowed.")
		}
		sum := fromRegister(sumRegister)
		log.Printf("sum = %s (%d)", sumRegister, sum)
		if sum == a+b {
			log.Printf("STATUS: Success at attempt %d.", attempts)
		} else if attempts == 10 {
			log.Printf("STATUS: Failed!!! Because %d != %d.", sum, a+b)
		} else {
			time.Sleep(time.Second)
			attempts++
			goto retry
		}
	}
	if *remove {
		if err := rca.remove(); err != nil {
			log.Fatal(err)
		}
	}
	if *listRegions {
		var regions []string
		for _, r := range endpoints.AwsPartition().Services()[cloudwatch.EndpointsID].Regions() {
			regions = append(regions, r.ID())
		}
		for _, r := range endpoints.AwsCnPartition().Services()[cloudwatch.EndpointsID].Regions() {
			regions = append(regions, r.ID())
		}
		sort.Strings(regions)
		for _, r := range regions {
			fmt.Println(r)
		}
	}
}
