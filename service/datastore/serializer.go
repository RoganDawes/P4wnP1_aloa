package datastore

import (
	"bytes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/encoding/proto"
	"io/ioutil"
	"errors"
)


type Serializer interface {
	Encode(val interface{}) (res []byte, err error)
	Decode(source []byte, destination interface{}) (err error)
}

type SerializerProtobuf struct {
	Codec encoding.Codec
	Compressor encoding.Compressor
}

func (s *SerializerProtobuf) Encode(val interface{}) (res []byte, err error) {
	raw,err := s.Codec.Marshal(val)
	if err != nil { return res,err }

	if s.Compressor == nil {
		return raw,nil
	} else {
		compressTarget := &bytes.Buffer{}
		compressSrc,err := s.Compressor.Compress(compressTarget)
		if err != nil { return res,err }
		n,err := compressSrc.Write(raw)
		if err != nil { return res,err }
		if n != len(raw) { return res,errors.New("Error while compressing serialized data")}

		if err = compressSrc.Close(); err != nil {
			return res,err
		}
		return compressTarget.Bytes(), nil
	}

	return
}

func (s *SerializerProtobuf) Decode(source []byte, destination interface{}) (err error) {
	if s.Compressor == nil {
		return s.Codec.Unmarshal(source, destination)
	} else {
		targetReader,err := s.Compressor.Decompress(bytes.NewReader(source))
		if err != nil { return err }
		decompressed, err := ioutil.ReadAll(targetReader)
		if err != nil { return err }
		return s.Codec.Unmarshal(decompressed, destination)
	}

}

func NewSerializerProtobuf(compress bool) *SerializerProtobuf {
	sz := &SerializerProtobuf{
		Codec: encoding.GetCodec(proto.Name),
	}
	if compress {
		sz.Compressor = encoding.GetCompressor(gzip.Name)
	}
	return sz
}

