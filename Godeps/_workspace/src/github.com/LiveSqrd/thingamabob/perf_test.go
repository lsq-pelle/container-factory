package thingamabob

import "testing"

func BenchmarkNaming(b *testing.B) {
	registry := getClient()
	defer func() {
		b.StopTimer()
		registry.Destroy()
	}()
	b.ResetTimer()

	var ctr uint64

	for i := 0; i < b.N; i++ {
		ctr += 1

		if _, err := registry.Id(string(ctr)); err != nil {
			panic(err)
		}
	}
}
