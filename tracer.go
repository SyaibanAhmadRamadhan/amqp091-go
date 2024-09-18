package amqp091

import "context"

type PublishTracer interface {
	TracePublisherStart(ctx context.Context, input TracePublisherStartInput) context.Context
	RecordError(ctx context.Context, err error)
}

type TracePublisherStartInput struct {
	Msg       Publishing
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
}
