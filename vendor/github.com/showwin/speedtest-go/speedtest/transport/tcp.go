package transport

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	pingPrefix = []byte{0x50, 0x49, 0x4e, 0x47, 0x20}
	// downloadPrefix = []byte{0x44, 0x4F, 0x57, 0x4E, 0x4C, 0x4F, 0x41, 0x44, 0x20}
	// uploadPrefix   = []byte{0x55, 0x50, 0x4C, 0x4F, 0x41, 0x44, 0x20}
	initPacket = []byte{0x49, 0x4e, 0x49, 0x54, 0x50, 0x4c, 0x4f, 0x53, 0x53}
	packetLoss = []byte{0x50, 0x4c, 0x4f, 0x53, 0x53}
	hiFormat   = []byte{0x48, 0x49}
	quitFormat = []byte{0x51, 0x55, 0x49, 0x54}
)

var (
	ErrEchoData                    = errors.New("incorrect echo data")
	ErrEmptyConn                   = errors.New("empty conn")
	ErrUnsupported                 = errors.New("unsupported protocol") // Some servers have disabled ip:8080, we return this error.
	ErrUninitializedPacketLossInst = errors.New("uninitialized packet loss inst")
)

func pingFormat(locTime int64) []byte {
	return strconv.AppendInt(pingPrefix, locTime, 10)
}

type Client struct {
	id      string
	conn    net.Conn
	host    string
	version string

	dialer *net.Dialer

	reader *bufio.Reader
}

func NewClient(dialer *net.Dialer) (*Client, error) {
	uuid, err := generateUUID()
	if err != nil {
		return nil, err
	}
	return &Client{
		id:     uuid,
		dialer: dialer,
	}, nil
}

func (client *Client) ID() string {
	return client.id
}

func (client *Client) Connect(ctx context.Context, host string) (err error) {
	client.host = host
	client.conn, err = client.dialer.DialContext(ctx, "tcp", client.host)
	if err != nil {
		return err
	}
	client.reader = bufio.NewReader(client.conn)
	return nil
}

func (client *Client) Disconnect() (err error) {
	_, _ = client.conn.Write(quitFormat)
	client.conn = nil
	client.reader = nil
	client.version = ""
	return
}

func (client *Client) Write(data []byte) (err error) {
	if client.conn == nil {
		return ErrEmptyConn
	}
	_, err = fmt.Fprintf(client.conn, "%s\n", data)
	return
}

func (client *Client) Read() ([]byte, error) {
	if client.conn == nil {
		return nil, ErrEmptyConn
	}
	return client.reader.ReadBytes('\n')
}

func (client *Client) Version() string {
	if len(client.version) == 0 {
		err := client.Write(hiFormat)
		if err == nil {
			message, err := client.Read()
			if err != nil || len(message) < 8 {
				return "unknown"
			}
			client.version = string(message[6 : len(message)-1])
		}
	}
	return client.version
}

// PingContext Measure latency(RTT) between client and server.
// We use the 2RTT method to obtain three RTT result in
// order to get more data in less time (t2-t0, t4-t2, t3-t1).
// And give lower weight to the delay measured by the server.
// local factor = 0.4 * 2 and remote factor = 0.2
// latency = 0.4 * (t2 - t0) + 0.4 * (t4 - t2) + 0.2 * (t3 - t1)
// @return cumulative delay in nanoseconds
func (client *Client) PingContext(ctx context.Context) (int64, error) {
	resultChan := make(chan error, 1)

	var accumulatedLatency int64 = 0
	var firstReceivedByServer int64 // t1

	go func() {
		for i := 0; i < 2; i++ {
			t0 := time.Now().UnixNano()
			if err := client.Write(pingFormat(t0)); err != nil {
				resultChan <- err
				return
			}
			data, err := client.Read()
			t2 := time.Now().UnixNano()
			if err != nil {
				resultChan <- err
				return
			}
			if len(data) != 19 {
				resultChan <- ErrEchoData
				return
			}
			tx, err := strconv.ParseInt(string(data[5:18]), 10, 64)
			if err != nil {
				resultChan <- err
				return
			}
			accumulatedLatency += (t2 - t0) * 4 / 10 // 0.4
			if i == 0 {
				firstReceivedByServer = tx
			} else {
				// append server-side latency result
				accumulatedLatency += (tx - firstReceivedByServer) * 1000 * 1000 * 2 / 10 // 0.2
			}
		}
		resultChan <- nil
		close(resultChan)
	}()

	select {
	case err := <-resultChan:
		return accumulatedLatency, err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (client *Client) InitPacketLoss() error {
	id := client.id
	payload := append(hiFormat, 0x20)
	payload = append(payload, []byte(id)...)
	err := client.Write(payload)
	if err != nil {
		return err
	}
	return client.Write(initPacket)
}

// PLoss Packet loss statistics
// The packet loss here generally refers to uplink packet loss.
// We use the following formula to calculate the packet loss:
// packetLoss = [1 - (Sent - Dup) / (Max + 1)] * 100%
type PLoss struct {
	Sent int `json:"sent"` // Number of sent packets acknowledged by the remote.
	Dup  int `json:"dup"`  // Number of duplicate packets acknowledged by the remote.
	Max  int `json:"max"`  // The maximum index value received by the remote.
}

func (p PLoss) String() string {
	if p.Sent == 0 {
		// if p.Sent == 0, maybe all data is dropped by the upper gateway.
		// we believe this feature is not applicable on this server now.
		return "Packet Loss: N/A"
	}
	return fmt.Sprintf("Packet Loss: %.2f%% (Sent: %d/Dup: %d/Max: %d)", p.Loss()*100, p.Sent, p.Dup, p.Max)
}

func (p PLoss) Loss() float64 {
	if p.Sent == 0 {
		return -1
	}
	return 1 - (float64(p.Sent-p.Dup))/float64(p.Max+1)
}

func (p PLoss) LossPercent() float64 {
	if p.Sent == 0 {
		return -1
	}
	return p.Loss() * 100
}

func (client *Client) PacketLoss() (*PLoss, error) {
	err := client.Write(packetLoss)
	if err != nil {
		return nil, err
	}
	result, err := client.Read()
	if err != nil {
		return nil, err
	}
	splitResult := bytes.Split(result, []byte{0x20})
	if len(splitResult) < 3 || !bytes.Equal(splitResult[0], packetLoss) {
		return nil, nil
	}
	x0, err := strconv.Atoi(string(splitResult[1]))
	if err != nil {
		return nil, err
	}
	x1, err := strconv.Atoi(string(splitResult[2]))
	if err != nil {
		return nil, err
	}
	x2, err := strconv.Atoi(string(bytes.TrimRight(splitResult[3], "\n")))
	if err != nil {
		return nil, err
	}
	return &PLoss{
		Sent: x0,
		Dup:  x1,
		Max:  x2,
	}, nil
}

func (client *Client) Download() {
	panic("Unimplemented method: Client.Download()")
}

func (client *Client) Upload() {
	panic("Unimplemented method: Client.Upload()")
}
