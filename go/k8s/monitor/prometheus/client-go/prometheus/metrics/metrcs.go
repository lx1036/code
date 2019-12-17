package metrics

import "github.com/golang/protobuf/proto"

func init() {
	proto.RegisterType((*LabelPair)(nil), "io.prometheus.client.LabelPair")
	proto.RegisterType((*Gauge)(nil), "io.prometheus.client.Gauge")
	proto.RegisterType((*Counter)(nil), "io.prometheus.client.Counter")
	proto.RegisterType((*Quantile)(nil), "io.prometheus.client.Quantile")
	proto.RegisterType((*Summary)(nil), "io.prometheus.client.Summary")
	proto.RegisterType((*Untyped)(nil), "io.prometheus.client.Untyped")
	proto.RegisterType((*Histogram)(nil), "io.prometheus.client.Histogram")
	proto.RegisterType((*Bucket)(nil), "io.prometheus.client.Bucket")
	proto.RegisterType((*Metric)(nil), "io.prometheus.client.Metric")
	proto.RegisterType((*MetricFamily)(nil), "io.prometheus.client.MetricFamily")
	proto.RegisterEnum("io.prometheus.client.MetricType", MetricType_name, MetricType_value)
}

func init() {
	proto.RegisterFile("metrics.proto", fileDescriptor_metrics_c97c9a2b9560cb8f)
}

type Metric struct {
	Label                []*LabelPair `protobuf:"bytes,1,rep,name=label" json:"label,omitempty"`
	Gauge                *Gauge       `protobuf:"bytes,2,opt,name=gauge" json:"gauge,omitempty"`
	Counter              *Counter     `protobuf:"bytes,3,opt,name=counter" json:"counter,omitempty"`
	Summary              *Summary     `protobuf:"bytes,4,opt,name=summary" json:"summary,omitempty"`
	Untyped              *Untyped     `protobuf:"bytes,5,opt,name=untyped" json:"untyped,omitempty"`
	Histogram            *Histogram   `protobuf:"bytes,7,opt,name=histogram" json:"histogram,omitempty"`
	TimestampMs          *int64       `protobuf:"varint,6,opt,name=timestamp_ms,json=timestampMs" json:"timestamp_ms,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}
type MetricFamily struct {
	Name                 *string     `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Help                 *string     `protobuf:"bytes,2,opt,name=help" json:"help,omitempty"`
	Type                 *MetricType `protobuf:"varint,3,opt,name=type,enum=io.prometheus.client.MetricType" json:"type,omitempty"`
	Metric               []*Metric   `protobuf:"bytes,4,rep,name=metric" json:"metric,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

type LabelPair struct {
	Name                 *string  `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Value                *string  `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Gauge struct {
	Value                *float64 `protobuf:"fixed64,1,opt,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Counter struct {
	Value                *float64 `protobuf:"fixed64,1,opt,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
