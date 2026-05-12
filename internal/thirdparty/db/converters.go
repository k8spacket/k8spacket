package db

import (
	"fmt"

	tcp_model "github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	tls_model "github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	proto_tcp "github.com/k8spacket/k8spacket/internal/proto/nodegraph/model"
	proto_tls "github.com/k8spacket/k8spacket/internal/proto/tlsparser/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Converter functions for TLSDetails
func tlsDetailsToProto(in *tls_model.TLSDetails) *proto_tls.TLSDetails {
	if in == nil {
		return nil
	}
	return &proto_tls.TLSDetails{
		Id:                 in.Id,
		Domain:             in.Domain,
		Dst:                in.Dst,
		Port:               uint32(in.Port),
		ClientTLSVersions:  in.ClientTLSVersions,
		ClientCipherSuites: in.ClientCipherSuites,
		UsedTLSVersion:     in.UsedTLSVersion,
		UsedCipherSuite:    in.UsedCipherSuite,
		Certificate: &proto_tls.Certificate{
			NotBefore:   timestamppb.New(in.Certificate.NotBefore),
			NotAfter:    timestamppb.New(in.Certificate.NotAfter),
			ServerChain: in.Certificate.ServerChain,
			LastScrape:  timestamppb.New(in.Certificate.LastScrape),
		},
	}
}

func tlsDetailsFromProto(in *proto_tls.TLSDetails) *tls_model.TLSDetails {
	if in == nil {
		return nil
	}
	cert := tls_model.Certificate{}
	if in.Certificate != nil {
		cert = tls_model.Certificate{
			NotBefore:   in.Certificate.NotBefore.AsTime(),
			NotAfter:    in.Certificate.NotAfter.AsTime(),
			ServerChain: in.Certificate.ServerChain,
			LastScrape:  in.Certificate.LastScrape.AsTime(),
		}
	}
	return &tls_model.TLSDetails{
		Id:                 in.Id,
		Domain:             in.Domain,
		Dst:                in.Dst,
		Port:               uint16(in.Port),
		ClientTLSVersions:  in.ClientTLSVersions,
		ClientCipherSuites: in.ClientCipherSuites,
		UsedTLSVersion:     in.UsedTLSVersion,
		UsedCipherSuite:    in.UsedCipherSuite,
		Certificate:        cert,
	}
}

// Converter functions for TLSConnection
func tlsConnectionToProto(in *tls_model.TLSConnection) *proto_tls.TLSConnection {
	if in == nil {
		return nil
	}
	return &proto_tls.TLSConnection{
		Id:              in.Id,
		Src:             in.Src,
		SrcName:         in.SrcName,
		SrcNamespace:    in.SrcNamespace,
		Dst:             in.Dst,
		DstName:         in.DstName,
		DstPort:         uint32(in.DstPort),
		Domain:          in.Domain,
		UsedTLSVersion:  in.UsedTLSVersion,
		UsedCipherSuite: in.UsedCipherSuite,
		LastSeen:        timestamppb.New(in.LastSeen),
	}
}

func tlsConnectionFromProto(in *proto_tls.TLSConnection) *tls_model.TLSConnection {
	if in == nil {
		return nil
	}
	return &tls_model.TLSConnection{
		Id:              in.Id,
		Src:             in.Src,
		SrcName:         in.SrcName,
		SrcNamespace:    in.SrcNamespace,
		Dst:             in.Dst,
		DstName:         in.DstName,
		DstPort:         uint16(in.DstPort),
		Domain:          in.Domain,
		UsedTLSVersion:  in.UsedTLSVersion,
		UsedCipherSuite: in.UsedCipherSuite,
		LastSeen:        in.LastSeen.AsTime(),
	}
}

// Converter functions for ConnectionItem
func connectionItemToProto(in *tcp_model.ConnectionItem) *proto_tcp.ConnectionItem {
	if in == nil {
		return nil
	}
	return &proto_tcp.ConnectionItem{
		Src:            in.Src,
		SrcName:        in.SrcName,
		SrcNamespace:   in.SrcNamespace,
		Dst:            in.Dst,
		DstName:        in.DstName,
		DstNamespace:   in.DstNamespace,
		ConnCount:      in.ConnCount,
		ConnPersistent: in.ConnPersistent,
		BytesSent:      in.BytesSent,
		BytesReceived:  in.BytesReceived,
		Duration:       in.Duration,
		MaxDuration:    in.MaxDuration,
		LastSeen:       timestamppb.New(in.LastSeen),
	}
}

func connectionItemFromProto(in *proto_tcp.ConnectionItem) *tcp_model.ConnectionItem {
	if in == nil {
		return nil
	}
	return &tcp_model.ConnectionItem{
		Src:            in.Src,
		SrcName:        in.SrcName,
		SrcNamespace:   in.SrcNamespace,
		Dst:            in.Dst,
		DstName:        in.DstName,
		DstNamespace:   in.DstNamespace,
		ConnCount:      in.ConnCount,
		ConnPersistent: in.ConnPersistent,
		BytesSent:      in.BytesSent,
		BytesReceived:  in.BytesReceived,
		Duration:       in.Duration,
		MaxDuration:    in.MaxDuration,
		LastSeen:       in.LastSeen.AsTime(),
	}
}

// marshalProto marshals a domain model to protobuf
func marshalProto(v interface{}) ([]byte, error) {
	// Handle basic types that bolthold might try to encode
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case *tls_model.TLSDetails:
		protoVal := tlsDetailsToProto(val)
		return marshalMessage(protoVal)
	case tls_model.TLSDetails:
		protoVal := tlsDetailsToProto(&val)
		return marshalMessage(protoVal)
	case *tls_model.TLSConnection:
		protoVal := tlsConnectionToProto(val)
		return marshalMessage(protoVal)
	case tls_model.TLSConnection:
		protoVal := tlsConnectionToProto(&val)
		return marshalMessage(protoVal)
	case *tcp_model.ConnectionItem:
		protoVal := connectionItemToProto(val)
		return marshalMessage(protoVal)
	case tcp_model.ConnectionItem:
		protoVal := connectionItemToProto(&val)
		return marshalMessage(protoVal)
	default:
		return nil, fmt.Errorf("unsupported type for protobuf marshaling: %T", v)
	}
}

// unmarshalProto unmarshals protobuf to domain model
func unmarshalProto(data []byte, v interface{}) error {
	// Handle basic types that bolthold might try to decode
	switch val := v.(type) {
	case *string:
		*val = string(data)
		return nil
	case *[]byte:
		*val = data
		return nil
	case *tls_model.TLSDetails:
		protoVal := &proto_tls.TLSDetails{}
		if err := unmarshalMessage(data, protoVal); err != nil {
			return err
		}
		domainVal := tlsDetailsFromProto(protoVal)
		*val = *domainVal
		return nil
	case *tls_model.TLSConnection:
		protoVal := &proto_tls.TLSConnection{}
		if err := unmarshalMessage(data, protoVal); err != nil {
			return err
		}
		domainVal := tlsConnectionFromProto(protoVal)
		*val = *domainVal
		return nil
	case *tcp_model.ConnectionItem:
		protoVal := &proto_tcp.ConnectionItem{}
		if err := unmarshalMessage(data, protoVal); err != nil {
			return err
		}
		domainVal := connectionItemFromProto(protoVal)
		*val = *domainVal
		return nil
	default:
		return fmt.Errorf("unsupported type for protobuf unmarshaling: %T", v)
	}
}

// marshalMessage is a helper to marshal protobuf messages
func marshalMessage(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// unmarshalMessage is a helper to unmarshal protobuf messages
func unmarshalMessage(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}
