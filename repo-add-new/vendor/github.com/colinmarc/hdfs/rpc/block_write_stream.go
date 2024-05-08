package rpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"

	hdfs "github.com/colinmarc/hdfs/protocol/hadoop_hdfs"
	"github.com/golang/protobuf/proto"
)

const (
	outboundPacketSize = 65536
	outboundChunkSize  = 512
	maxPacketsInQueue  = 5
)

// blockWriteStream writes data out to a datanode, and reads acks back.
type blockWriteStream struct {
	block *hdfs.LocatedBlockProto

	conn   io.ReadWriter
	buf    bytes.Buffer
	offset int64
	closed bool

	packets chan outboundPacket
	seqno   int

	ackError        error
	acksDone        chan struct{}
	lastPacketSeqno int
}

type outboundPacket struct {
	seqno     int
	offset    int64
	last      bool
	checksums []byte
	data      []byte
}

type ackError struct {
	pipelineIndex int
	seqno         int
	status        hdfs.Status
}

func (ae ackError) Error() string {
	return fmt.Sprintf("Ack error from datanode: %s", ae.status.String())
}

var ErrInvalidSeqno = errors.New("invalid ack sequence number")

func newBlockWriteStream(conn io.ReadWriter, offset int64) *blockWriteStream {
	s := &blockWriteStream{
		conn:     conn,
		offset:   offset,
		seqno:    1,
		packets:  make(chan outboundPacket, maxPacketsInQueue),
		acksDone: make(chan struct{}),
	}

	// Ack packets in the background.
	go func() {
		s.ackPackets()
		close(s.acksDone)
	}()

	return s
}

// func newBlockWriteStreamForRecovery(conn io.ReadWriter, oldWriteStream *blockWriteStream) {
// 	s := &blockWriteStream{
// 		conn: conn,
// 		buf: oldWriteStream.buf,
// 		packets: oldWriteStream.packets,
// 		offset: oldWriteStream.offset,
// 		seqno: oldWriteStream.seqno,
// 		packets
// 	}

// 	go s.ackPackets()
// 	return s
// }

func (s *blockWriteStream) Write(b []byte) (int, error) {
	if s.closed {
		return 0, io.ErrClosedPipe
	}

	if err := s.getAckError(); err != nil {
		return 0, err
	}

	n, _ := s.buf.Write(b)
	err := s.flush(false)
	return n, err
}

// finish flushes the rest of the buffered bytes, and then sends a final empty
// packet signifying the end of the block.
func (s *blockWriteStream) finish() error {
	if s.closed {
		return nil
	}
	s.closed = true

	if err := s.getAckError(); err != nil {
		return err
	}

	err := s.flush(true)
	if err != nil {
		return err
	}

	// The last packet has no data; it's just a marker that the block is finished.
	lastPacket := outboundPacket{
		seqno:     s.seqno,
		offset:    s.offset,
		last:      true,
		checksums: []byte{},
		data:      []byte{},
	}
	s.packets <- lastPacket
	err = s.writePacket(lastPacket)
	if err != nil {
		return err
	}
	close(s.packets)

	// Wait for the ack loop to finish.
	<-s.acksDone

	// Check one more time for any ack errors.
	if err := s.getAckError(); err != nil {
		return err
	}

	return nil
}

// flush parcels out the buffered bytes into packets, which it then flushes to
// the datanode. We keep around a reference to the packet, in case the ack
// fails, and we need to send it again later.
func (s *blockWriteStream) flush(force bool) error {
	if err := s.getAckError(); err != nil {
		return err
	}

	for s.buf.Len() > 0 && (force || s.buf.Len() >= outboundPacketSize) {
		packet := s.makePacket()
		s.packets <- packet
		s.offset += int64(len(packet.data))
		s.seqno++

		err := s.writePacket(packet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *blockWriteStream) makePacket() outboundPacket {
	packetLength := outboundPacketSize
	if s.buf.Len() < outboundPacketSize {
		packetLength = s.buf.Len()
	}

	// If we're starting from a weird offset (usually because of an Append), HDFS
	// gets unhappy unless we first align to a chunk boundary with a small packet.
	// Otherwise it yells at us with "a partial chunk must be sent in an
	// individual packet" or just complains about a corrupted block.
	alignment := int(s.offset) % outboundChunkSize
	if alignment > 0 && packetLength > (outboundChunkSize-alignment) {
		packetLength = outboundChunkSize - alignment
	}

	numChunks := int(math.Ceil(float64(packetLength) / float64(outboundChunkSize)))
	packet := outboundPacket{
		seqno:     s.seqno,
		offset:    s.offset,
		last:      false,
		checksums: make([]byte, numChunks*4),
		data:      make([]byte, packetLength),
	}

	// TODO: we shouldn't actually need this extra copy. We should also be able
	// to "reuse" packets.
	io.ReadFull(&s.buf, packet.data)

	// Fill in the checksum for each chunk of data.
	for i := 0; i < numChunks; i++ {
		chunkOff := i * outboundChunkSize
		chunkEnd := chunkOff + outboundChunkSize
		if chunkEnd >= len(packet.data) {
			chunkEnd = len(packet.data)
		}

		checksum := crc32.Checksum(packet.data[chunkOff:chunkEnd], crc32.IEEETable)
		binary.BigEndian.PutUint32(packet.checksums[i*4:], checksum)
	}

	return packet
}

// ackPackets is meant to run in the background, reading acks and setting
// ackError if one fails.
func (s *blockWriteStream) ackPackets() {
	reader := bufio.NewReader(s.conn)

	for {
		p, ok := <-s.packets
		if !ok {
			// All packets all acked.
			return
		}

		// If we fail to read the ack at all, that counts as a failure from the
		// first datanode (the one we're connected to).
		ack := &hdfs.PipelineAckProto{}
		err := readPrefixedMessage(reader, ack)
		if err != nil {
			s.ackError = err
			break
		}

		seqno := int(ack.GetSeqno())
		for i, status := range ack.GetReply() {
			if status != hdfs.Status_SUCCESS {
				s.ackError = ackError{status: status, seqno: seqno, pipelineIndex: i}
				break
			}
		}

		if seqno != p.seqno {
			s.ackError = ErrInvalidSeqno
			break
		}
	}

	// Once we've seen an error, just keep reading packets off the channel (but
	// not off the socket) until the writing thread figures it out. If we don't,
	// the upstream thread could deadlock waiting for the channel to have space.
	for _ = range s.packets {
	}
}

func (s *blockWriteStream) getAckError() error {
	select {
	case <-s.acksDone:
		if s.ackError != nil {
			return s.ackError
		}
	default:
	}

	return nil
}

// A packet for the datanode:
// +-----------------------------------------------------------+
// |  uint32 length of the packet                              |
// +-----------------------------------------------------------+
// |  size of the PacketHeaderProto, uint16                    |
// +-----------------------------------------------------------+
// |  PacketHeaderProto                                        |
// +-----------------------------------------------------------+
// |  N checksums, 4 bytes each                                |
// +-----------------------------------------------------------+
// |  N chunks of payload data                                 |
// +-----------------------------------------------------------+
func (s *blockWriteStream) writePacket(p outboundPacket) error {
	headerInfo := &hdfs.PacketHeaderProto{
		OffsetInBlock:     proto.Int64(p.offset),
		Seqno:             proto.Int64(int64(p.seqno)),
		LastPacketInBlock: proto.Bool(p.last),
		DataLen:           proto.Int32(int32(len(p.data))),
	}

	header := make([]byte, 6)
	infoBytes, err := proto.Marshal(headerInfo)
	if err != nil {
		return err
	}

	// Don't ask me why this doesn't include the header proto...
	totalLength := len(p.data) + len(p.checksums) + 4
	binary.BigEndian.PutUint32(header, uint32(totalLength))
	binary.BigEndian.PutUint16(header[4:], uint16(len(infoBytes)))
	header = append(header, infoBytes...)

	_, err = s.conn.Write(header)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(p.checksums)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(p.data)
	if err != nil {
		return err
	}

	return nil
}
