package amqp091

import (
	"context"
	"encoding/base64"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"time"
)

const TracerName = "github.com/SyaibanAhmadRamadhan/amqp091"
const InstrumentVersion = "v1.0.0"

type Otel struct {
	tracer trace.Tracer
	attrs  []attribute.KeyValue
}

func NewOtel() *Otel {
	tp := otel.GetTracerProvider()
	return &Otel{
		tracer: tp.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		attrs:  []attribute.KeyValue{semconv.MessagingSystemRabbitmq},
	}
}

func recordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func (t *Otel) TracePublisherStart(ctx context.Context, payload *basicPublish) context.Context {
	spanName := "Rabbitmq Publish Message"
	ctx, span := t.tracer.Start(ctx, spanName)
	span.SetAttributes(

		attribute.String("messaging.rabbitmq.exchange", payload.Exchange),
		attribute.String("messaging.rabbitmq.destination.routing_key", payload.RoutingKey),
		//attribute.String("messaging.rabbitmq.body", truncateBody(payload.Body)),
		attribute.Bool("messaging.rabbitmq.mandatory", payload.Mandatory),
		attribute.Bool("messaging.rabbitmq.immediate", payload.Immediate),
		attribute.String("messaging.rabbitmq.properties.content_type", payload.Properties.ContentType),
		attribute.String("messaging.rabbitmq.properties.content_encoding", payload.Properties.ContentEncoding),
		attribute.Int("messaging.rabbitmq.properties.delivery_mode", int(payload.Properties.DeliveryMode)),
		attribute.Int("messaging.rabbitmq.properties.priority", int(payload.Properties.Priority)),
		attribute.String("messaging.rabbitmq.properties.correlation_id", payload.Properties.CorrelationId),
		attribute.String("messaging.rabbitmq.properties.reply_to", payload.Properties.ReplyTo),
		attribute.String("messaging.rabbitmq.properties.expiration", payload.Properties.Expiration),
		attribute.String("messaging.rabbitmq.properties.message_id", payload.Properties.MessageId),
		attribute.String("messaging.rabbitmq.properties.timestamp", payload.Properties.Timestamp.Format(time.RFC3339)),
		attribute.String("messaging.rabbitmq.properties.type", payload.Properties.Type),
		attribute.String("messaging.rabbitmq.properties.user_id", payload.Properties.UserId),
		attribute.String("messaging.rabbitmq.properties.app_id", payload.Properties.AppId),
		attribute.String("messaging.rabbitmq.properties.headers", formatHeaders(payload.Properties.Headers)),
	)
	span.End()
	return ctx
}

func (t *Otel) TracePublishEnd(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		recordError(span, err)
		span.End()
	}
}

func truncateBody(body []byte) string {
	const maxBodyLength = 1024

	utf8Body := base64.StdEncoding.EncodeToString(body)
	if len(body) > maxBodyLength {
		return string(utf8Body[:maxBodyLength]) + "...(truncated)"
	}
	return utf8Body
}

func formatHeaders(headers map[string]interface{}) string {
	if len(headers) == 0 {
		return ""
	}
	var sb strings.Builder
	for k, v := range headers {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprintf("%v", v))
		sb.WriteString(", ")
	}
	return strings.TrimSuffix(sb.String(), ", ")
}
