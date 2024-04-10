package memory

import (
	"context"
	"fmt"
)

func Example() {
	s := NewStorage()

	ctx := context.Background()

	metricName := "foo"
	err := s.GaugeSet(ctx, metricName, 19.17)
	if err != nil {
		fmt.Printf("We got error: %s\n", err)
	} else {
		val, _ := s.GaugeGet(ctx, metricName)
		fmt.Printf("Gauge value is: %f\n", val)
	}

	// Output:
	// Gauge value is: 19.170000
}
