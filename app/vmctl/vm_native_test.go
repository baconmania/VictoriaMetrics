package main

import (
	"context"
	"testing"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmctl/native"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmctl/stepper"
)

// If you want to run this test:
// 1. run two instances of victoriametrics and define -httpListenAddr for both or just for second instance
// 2. define srcAddr and dstAddr const with your victoriametrics addresses
// 3. define matchFilter const with your importing data
// 4. define timeStartFilter
// 5. run each test one by one

const (
	matchFilter     = `{job="avalanche"}`
	timeStartFilter = "2020-01-01T20:07:00Z"
	timeEndFilter   = "2020-08-01T20:07:00Z"
	srcAddr         = "http://127.0.0.1:8428"
	dstAddr         = "http://127.0.0.1:8528"
)

// This test simulates close process if user abort it
func Test_vmNativeProcessor_run(t *testing.T) {
	t.Skip()
	type fields struct {
		filter    native.Filter
		rateLimit int64
		dst       *native.Client
		src       *native.Client
	}
	tests := []struct {
		name    string
		fields  fields
		closer  func(cancelFunc context.CancelFunc)
		wantErr bool
	}{
		{
			name: "simulate syscall.SIGINT",
			fields: fields{
				filter: native.Filter{
					Match:     matchFilter,
					TimeStart: timeStartFilter,
				},
				rateLimit: 0,
				dst: &native.Client{
					Addr: dstAddr,
				},
				src: &native.Client{
					Addr: srcAddr,
				},
			},
			closer: func(cancelFunc context.CancelFunc) {
				time.Sleep(time.Second * 5)
				cancelFunc()
			},
			wantErr: true,
		},
		{
			name: "simulate correct work",
			fields: fields{
				filter: native.Filter{
					Match:     matchFilter,
					TimeStart: timeStartFilter,
				},
				rateLimit: 0,
				dst: &native.Client{
					Addr: dstAddr,
				},
				src: &native.Client{
					Addr: srcAddr,
				},
			},
			closer:  func(cancelFunc context.CancelFunc) {},
			wantErr: false,
		},
		{
			name: "simulate correct work with chunking",
			fields: fields{
				filter: native.Filter{
					Match:     matchFilter,
					TimeStart: timeStartFilter,
					TimeEnd:   timeEndFilter,
					Chunk:     stepper.StepMonth,
				},
				rateLimit: 0,
				dst: &native.Client{
					Addr: dstAddr,
				},
				src: &native.Client{
					Addr: srcAddr,
				},
			},
			closer:  func(cancelFunc context.CancelFunc) {},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelFn := context.WithCancel(context.Background())
			p := &vmNativeProcessor{
				filter:    tt.fields.filter,
				rateLimit: tt.fields.rateLimit,
				dst:       tt.fields.dst,
				src:       tt.fields.src,
			}

			tt.closer(cancelFn)

			if err := p.run(ctx, true); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
