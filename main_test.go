package main

import (
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"
)

var evaluationLatency = 3 * time.Second

func TestMain(m *testing.M) {
	flag.StringVar(&profile, "profile", "", "the AWS profile to use for test credentials (tests won't run if empty)")
	flag.StringVar(&region, "region", "eu-west-1", "the AWS region to create/use alarms in")
	flag.BoolVar(&verbose, "verbose", false, "log diagnostic messages")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}
