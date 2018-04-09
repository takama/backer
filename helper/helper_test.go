package helper

import (
	"testing"
)

func test(t *testing.T, expected bool, messages ...interface{}) {
	if !expected {
		t.Error(messages)
	}
}

func TestRoundPrice(t *testing.T) {
	testData := []struct {
		from float32
		to   float32
	}{
		{
			1.23456,
			1.23,
		},
		{
			1.255,
			1.25,
		},
		{
			1.2555,
			1.26,
		},
		{
			0,
			0,
		},
		{
			-1.23456,
			-1.23,
		},
		{
			-0.99990,
			-1.00,
		},
		{
			0.000000000000000001,
			0,
		},
	}
	for _, item := range testData {
		result := RoundPrice(item.from)
		test(t, result == item.to,
			"Expected result for", item.from, "->", item.to, "got:", result)
	}
}

func TestTruncatePrice(t *testing.T) {
	testData := []struct {
		from float32
		to   float32
	}{
		{
			1.23456,
			1.23,
		},
		{
			1.255,
			1.25,
		},
		{
			1.2555,
			1.25,
		},
		{
			0,
			0,
		},
		{
			-1.23456,
			-1.23,
		},
		{
			-0.99990,
			-0.99,
		},
		{
			0.000000000000000001,
			0,
		},
	}
	for _, item := range testData {
		result := TruncatePrice(item.from)
		test(t, result == item.to,
			"Expected result for", item.from, "->", item.to, "got:", result)
	}
}
