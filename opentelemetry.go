package amqp091

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"runtime/debug"
	"strings"
)

const TracerName = "github.com/SyaibanAhmadRamadhan/amqp091"
const InstrumentVersion = "v1.0.0"
const amqpLibName = "github.com/rabbitmq/amqp091-go"
const netProtocolVer = "0.9.1"

var (
	amqpLibVersion = "unknown"
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, dep := range info.Deps {
		if dep.Path == amqpLibName {
			amqpLibVersion = dep.Version
		}
	}
}

type Otel struct {
	tracer      trace.Tracer
	propagators propagation.TextMapPropagator
	attrs       []attribute.KeyValue
}

func NewOtel(uri URI) *Otel {
	tp := otel.GetTracerProvider()
	return &Otel{
		tracer:      tp.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		propagators: otel.GetTextMapPropagator(),
		attrs: []attribute.KeyValue{
			semconv.ServiceName(amqpLibName),
			semconv.ServiceVersion(amqpLibVersion),
			semconv.MessagingSystemRabbitmq,
			semconv.NetworkProtocolName(uri.Scheme),
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
			semconv.ServerAddress(uri.Host),
			semconv.ServerPort(uri.Port),
		},
	}
}

func recordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func (t *Otel) TracePublisherStart(ctx context.Context, input TracePublisherStartInput) context.Context {
	attrs := []attribute.KeyValue{
		semconv.MessagingRabbitmqDestinationRoutingKey(input.Key),
		semconv.MessagingOperationTypePublish,
		semconv.MessagingOperationName("publish"),
		semconv.MessagingMessageBodySize(len(input.Msg.Body)),
		semconv.MessagingMessageConversationID(input.Msg.CorrelationId),
		semconv.MessagingMessageID(input.Msg.MessageId),
		semconv.MessagingDestinationPublishAnonymous(input.Exchange == ""),
		semconv.MessagingDestinationPublishName(input.Exchange),
	}
	if input.Msg.CorrelationId != "" {
		attrs = append(attrs, semconv.MessagingMessageConversationID(input.Msg.CorrelationId))
	}
	if input.Msg.MessageId != "" {
		attrs = append(attrs, semconv.MessagingMessageID(input.Msg.MessageId))
	}
	attrs = append(attrs, t.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	ctx, span := t.tracer.Start(ctx, t.nameWhenPublish(input.Exchange), opts...)
	carrier := newPublishingMessageCarrier(&input.Msg)
	t.propagators.Inject(ctx, carrier)
	span.End()
	return ctx
}

func (t *Otel) RecordError(ctx context.Context, err error) {
	spanName := "Rabbitmq Recording Error"
	_, span := t.tracer.Start(ctx, spanName)
	recordError(span, err)
	span.End()
}

func (*Otel) nameWhenPublish(exchange string) string {
	if exchange == "" {
		exchange = "(default)"
	}
	return "publish " + exchange
}

func queueAnonymous(queue string) bool {
	return strings.HasPrefix(queue, "amq.gen-")
}

func (*Otel) nameWhenConsume(queue string) string {
	if queueAnonymous(queue) {
		queue = "(anonymous)"
	}
	return "process " + queue
}
