package main

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func Test_getStringENV(t *testing.T) {
	os.Setenv("TEST_PORT", "8081")
	type args struct {
		key string
		def string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "String ENV key - exists and loads",
			args: args{
				key: "TEST_PORT",
				def: "8080",
			},
			want: "8081",
		},
		{
			name: "String ENV key - does not exist",
			args: args{
				key: "TEST_PORT_",
				def: "8080",
			},
			want: "8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringENV(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("getStringENV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIntENV(t *testing.T) {
	os.Setenv("TEST_TIMEOUT", "30")
	type args struct {
		key string
		def int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Int ENV key - exists and loads",
			args: args{
				key: "TEST_TIMEOUT",
				def: 60,
			},
			want: 30,
		},
		{
			name: "Int ENV key - does not exist",
			args: args{
				key: "TEST_TIMEOUT_",
				def: 60,
			},
			want: 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIntENV(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("getIntENV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricsDB_SumMetric(t *testing.T) {
	type args struct {
		key string
		to  int
	}
	tests := []struct {
		name string
		mdb  MetricsDB
		args args
		want *Metric
	}{
		{
			name: "Metrics in 0",
			mdb:  make(MetricsDB, 0),
			args: args{
				key: "active_visitors",
				to:  60,
			},
			want: &Metric{
				Value: 0,
			},
		},
		{
			name: "Got 3 active visitors, records not timed out",
			mdb: MetricsDB{
				"active_visitors": []Metric{
					Metric{Value: 2, Timestamp: time.Now()},
					Metric{Value: 1, Timestamp: time.Now()},
				},
			},
			args: args{
				key: "active_visitors",
				to:  60,
			},
			want: &Metric{
				Value: 3,
			},
		},
		{
			name: "Got 12 active visitors, one record timed out",
			mdb: MetricsDB{
				"active_visitors": []Metric{
					Metric{Value: 4, Timestamp: time.Now().Add(-2 * time.Hour)},
					Metric{Value: 7, Timestamp: time.Now().Add(-10 * time.Minute)},
					Metric{Value: 2, Timestamp: time.Now().Add(-15 * time.Minute)},
					Metric{Value: 3, Timestamp: time.Now().Add(-20 * time.Minute)},
				},
			},
			args: args{
				key: "active_visitors",
				to:  60,
			},
			want: &Metric{
				Value: 12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mdb.SumMetric(tt.args.key, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricsDB.SumMetric() = %v, want %v", got.Value, tt.want.Value)
			} else {
				t.Logf("MetricsDB.SumMetric() = %v, want %v", got.Value, tt.want.Value)
			}
		})
	}
}
