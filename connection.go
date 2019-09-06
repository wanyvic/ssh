// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
)

// OpenChannelError is returned if the other side rejects an
// OpenChannel request.
type OpenChannelError struct {
	Reason  RejectionReason
	Message string
}

func (e *OpenChannelError) Error() string {
	return fmt.Sprintf("ssh: rejected: %s (%s)", e.Reason, e.Message)
}

// ConnMetadata holds metadata for the connection.
type ConnMetadata interface {
	// User returns the user ID for this connection.
	User() string

	// SessionID returns the session hash, also denoted by H.
	SessionID() []byte

	// ClientVersion returns the client's version string as hashed
	// into the session ID.
	ClientVersion() []byte

	// ServerVersion returns the server's version string as hashed
	// into the session ID.
	ServerVersion() []byte

	// RemoteAddr returns the remote address for this connection.
	RemoteMultiaddr() ma.Multiaddr

	// LocalAddr returns the local address for this connection.
	LocalMultiaddr() ma.Multiaddr
}

// Conn represents an SSH connection for both server and client roles.
// Conn is the basis for implementing an application layer, such
// as ClientConn, which implements the traditional shell access for
// clients.
type Conn interface {
	ConnMetadata

	// SendRequest sends a global request, and returns the
	// reply. If wantReply is true, it returns the response status
	// and payload. See also RFC4254, section 4.
	SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error)

	// OpenChannel tries to open an channel. If the request is
	// rejected, it returns *OpenChannelError. On success it returns
	// the SSH Channel and a Go channel for incoming, out-of-band
	// requests. The Go channel must be serviced, or the
	// connection will hang.
	OpenChannel(name string, data []byte) (Channel, <-chan *Request, error)

	// Close closes the underlying network connection
	Close() error

	// Wait blocks until the connection has shut down, and returns the
	// error causing the shutdown.
	Wait() error

	// TODO(hanwen): consider exposing:
	//   RequestKeyChange
	//   Disconnect
}

// DiscardRequests consumes and rejects all requests from the
// passed-in channel.
func DiscardRequests(in <-chan *Request) {
	for req := range in {
		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}

// A connection represents an incoming connection.
type connection struct {
	transport *handshakeTransport
	sshConn

	// The connection protocol.
	*mux
}

func (c *connection) Close() error {
	return c.sshConn.stream.Close()
}

// sshconn provides net.Conn metadata, but disallows direct reads and
// writes.
type sshConn struct {
	stream network.Stream

	user          string
	sessionID     []byte
	clientVersion []byte
	serverVersion []byte
}

func dup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (c *sshConn) User() string {
	return c.user
}

func (c *sshConn) RemoteMultiaddr() ma.Multiaddr {
	return c.stream.Conn().RemoteMultiaddr()
}

func (c *sshConn) Close() error {
	return c.stream.Close()
}

func (c *sshConn) LocalMultiaddr() ma.Multiaddr {
	return c.stream.Conn().LocalMultiaddr()
}

func (c *sshConn) SessionID() []byte {
	return dup(c.sessionID)
}

func (c *sshConn) ClientVersion() []byte {
	return dup(c.clientVersion)
}

func (c *sshConn) ServerVersion() []byte {
	return dup(c.serverVersion)
}
