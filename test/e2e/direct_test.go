// +build e2e

/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"knative.dev/eventing-rabbitmq/test/e2e/config/direct"
	"knative.dev/eventing-rabbitmq/test/e2e/config/recorder"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"

	"knative.dev/eventing/pkg/test/observer"
	recorder_collector "knative.dev/eventing/pkg/test/observer/recorder-collector"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "knative.dev/pkg/system/testing"
)

//
// producer ---> broker --[trigger]--> recorder
//

// DirectTestBrokerImpl makes sure an RabbitMQ Broker delivers events to a single consumer.
func DirectTestBroker() *feature.Feature {
	f := new(feature.Feature)

	f.Setup("install test resources", direct.Install())
	f.Setup("install recorder", recorder.Install())
	f.Alpha("RabbitMQ broker").Must("goes ready", AllGoReady)
	f.Alpha("RabbitMQ broker").Must("goes ready", CheckDirectEvents)
	return f
}

func CheckDirectEvents(ctx context.Context, t *testing.T) {
	env := environment.FromContext(ctx)

	sendCount := 5
	// TODO: we want a wait for events for x time in the future.
	time.Sleep(1 * time.Minute)

	c := recorder_collector.New(ctx)

	from := duckv1.KReference{
		Kind:       "Namespace",
		Name:       "default",
		APIVersion: "v1",
	}

	obsName := "recorder-" + env.Namespace()
	events, err := c.List(ctx, from, func(ob observer.Observed) bool {
		return ob.Observer == obsName
	})
	if err != nil {
		t.Fatal("failed to list observed events, ", err)
	}

	for i, e := range events {
		fmt.Printf("[%d]: seen by %q\n%s\n", i, e.Observer, e.Event)
	}

	got := len(events)
	want := sendCount
	if want != got {
		t.Errorf("failed to observe the correct number of events, want: %d, got: %d", want, got)
	}

	// Pass!
}
