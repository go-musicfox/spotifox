package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/arcspace/go-arc-sdk/apis/arc"
	"github.com/arcspace/go-librespot/Spotify"
	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	"github.com/arcspace/go-librespot/librespot/asset"
	"github.com/arcspace/go-librespot/librespot/core/connection"
	"github.com/arcspace/go-librespot/librespot/core/crypto"
	"github.com/arcspace/go-librespot/librespot/mercury"
)

func init() {
	respot.StartNewSession = StartSession
}

func StartSession(ctx *respot.SessionContext) (respot.Session, error) {
	s := &Session{
		ctx: ctx,
	}

	if s.ctx.Keys == nil {
		s.ctx.Keys = crypto.GenerateKeys()
	}

	if s.ctx.DeviceUID == "" {
		s.ctx.DeviceUID = respot.GenerateDeviceUID(s.ctx.DeviceName)
	}

	err := s.StartConnection()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Session represents an active Spotify connection
type Session struct {
	ctx        *respot.SessionContext
	closed     atomic.Int32
	tcpCon     io.ReadWriter           // plain I/O network connection to the server
	stream     connection.PacketStream // encrypted connection to the Spotify server
	mercury    *mercury.Client         // mercury client associated with this session
	downloader asset.Downloader        // manages downloads
	//discovery  *discovery.Discovery    // discovery service used for Spotify Connect devices discovery
}

func (s *Session) Stream() connection.PacketStream {
	return s.stream
}

// func (s *Session) Discovery() *discovery.Discovery {
// 	return s.discovery
// }

func (s *Session) Mercury() *mercury.Client {
	return s.mercury
}

func (s *Session) Downloader() asset.Downloader {
	return s.downloader
}

func (s *Session) Context() *respot.SessionContext {
	return s.ctx
}

func (s *Session) PinTrack(trackID string, opts respot.PinOpts) (arc.MediaAsset, error) {
	asset, err := s.downloader.PinTrack(trackID)
	if err != nil {
		return nil, err
	}
	if opts.StartInternally {
		err = asset.OnStart(s.ctx.Context)
	}
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *Session) StartConnection() error {

	apUrl, err := respot.APResolve()
	if err != nil {
		return fmt.Errorf("could not get ap url: %+v", err)
	}
	s.tcpCon, err = net.Dial("tcp", apUrl)
	if err != nil {
		return fmt.Errorf("could not connect to %q: %+v", apUrl, err)
	}

	// First, start by performing a plaintext connection and send the Hello message
	conn := connection.NewPlainConnection(s.tcpCon, s.tcpCon)

	helloMessage, err := makeHelloMessage(s.ctx.Keys.PubKey(), s.ctx.Keys.ClientNonce())
	if err != nil {
		return fmt.Errorf("could not make hello packet: %+v", err)
	}
	initClientPacket, err := conn.SendPrefixPacket([]byte{0, 4}, helloMessage)
	if err != nil {
		return fmt.Errorf("could not write client hello: %+v", err)
	}

	// Wait and read the hello reply
	initServerPacket, err := conn.RecvPacket()
	if err != nil {
		return fmt.Errorf("could not recv hello response: %+v", err)
	}

	response := Spotify.APResponseMessage{}
	err = proto.Unmarshal(initServerPacket[4:], &response)
	if err != nil {
		return fmt.Errorf("could not unmarshal hello response: %+v", err)
	}

	remoteKey := response.Challenge.LoginCryptoChallenge.DiffieHellman.Gs
	sharedKeys := s.ctx.Keys.AddRemoteKey(remoteKey, initClientPacket, initServerPacket)

	plainResponse := &Spotify.ClientResponsePlaintext{
		LoginCryptoResponse: &Spotify.LoginCryptoResponseUnion{
			DiffieHellman: &Spotify.LoginCryptoDiffieHellmanResponse{
				Hmac: sharedKeys.Challenge,
			},
		},
		PowResponse:    &Spotify.PoWResponseUnion{},
		CryptoResponse: &Spotify.CryptoResponseUnion{},
	}

	plainResponseMessage, err := proto.Marshal(plainResponse)
	if err != nil {
		return fmt.Errorf("could no marshal response: %+v", err)
	}

	_, err = conn.SendPrefixPacket([]byte{}, plainResponseMessage)
	if err != nil {
		return fmt.Errorf("could no write client plain response: %+v", err)
	}

	s.stream = crypto.CreateStream(sharedKeys, conn)
	s.mercury = mercury.CreateMercury(s.stream)
	s.downloader = asset.NewDownloader(s.stream, s.mercury)
	return nil
}

/*
	func sessionFromDiscovery(d *discovery.Discovery) (*Session, error) {
		s, err := setupSession()
		if err != nil {
			return nil, err
		}

		s.discovery = d
		s.DeviceId = d.DeviceId()
		s.DeviceName = d.DeviceName()

		err = s.startConnection()
		if err != nil {
			return s, err
		}

		loginPacket, err := s.getLoginBlobPacket(d.LoginBlob())
		if err != nil {
			return nil, fmt.Errorf("could not get login blob packet: %+v", err)
		}
		return s, s.doLogin(loginPacket, d.LoginBlob().Username)
	}

*/

func (s *Session) Close() error {
	s.closed.Store(1)
	return s.disconnect()
}

func (s *Session) disconnect() error {
	if s.tcpCon != nil {
		conn := s.tcpCon.(net.Conn)
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("could not close connection: %+v", err)
		}
		s.tcpCon = nil
	}
	return nil
}

func (s *Session) isClosing() bool {
	select {
	case <-s.ctx.Context.Closing():
		return true
	default:
		if s.closed.Load() != 0 {
			return true
		}
	}

	return false
}

func (s *Session) doReconnect() error {
	s.disconnect()

	if s.isClosing() {
		return errors.New("respot session is closing")
	}

	err := s.StartConnection()
	if err != nil {
		return err
	}

	packet := s.makeLoginBlobPacket(
		s.ctx.Info.Username,
		s.ctx.Info.AuthBlob,
		Spotify.AuthenticationType_AUTHENTICATION_STORED_SPOTIFY_CREDENTIALS.Enum(),
	)
	return s.startSession(packet, s.ctx.Info.Username)
}

func (s *Session) planReconnect() {
	if s.isClosing() {
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		if err := s.doReconnect(); err != nil {
			s.planReconnect()
		}
	}()
}

func (s *Session) handle(cmd uint8, data []byte) error {
	switch {
	case cmd == connection.PacketPing:
		err := s.stream.SendPacket(connection.PacketPong, data)
		if err != nil {
			return fmt.Errorf("error handling ping: %+v", err)
		}

	case cmd == connection.PacketPongAck:
		// Pong reply, ignore

	case cmd == connection.PacketAesKey || cmd == connection.PacketAesKeyError || cmd == connection.PacketStreamChunkRes:
		// Audio key and data responses
		if err := s.downloader.HandleCmd(cmd, data); err != nil {
			return fmt.Errorf("could not handle cmd: %+v", err)
		}

	case cmd == connection.PacketCountryCode:
		s.ctx.Info.Country = string(data)

	case 0xb2 <= cmd && cmd <= 0xb6:
		// Mercury responses
		err := s.mercury.Handle(cmd, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("error handling 0xB?: %+v", err)
		}

	case cmd == connection.PacketSecretBlock:
		// Old RSA public key

	case cmd == connection.PacketLegacyWelcome:
		// Empty welcome packet

	case cmd == connection.PacketProductInfo:
		// Has some info about A/B testing status, product setup, etc... in an XML fashion.

	case cmd == 0x1f:
		// Unknown, data is zeroes only

	case cmd == connection.PacketLicenseVersion:
		// This is a simple blob containing the current Spotify license version (e.g. 1.0.1-FR). Format of the blob
		// is [ uint16 id (= 0x001), uint8 len, string license ]

	default:
		log.Printf("un implemented command received: %+v\n", cmd)
	}

	return nil
}

func (s *Session) Search(query string, limit int) (*mercury.SearchResponse, error) {
	return s.mercury.Search(query, limit, s.ctx.Info.Country, s.ctx.Info.Username)
}

func readInt(b *bytes.Buffer) uint32 {
	c, _ := b.ReadByte()
	lo := uint32(c)
	if lo&0x80 == 0 {
		return lo
	}

	c2, _ := b.ReadByte()
	hi := uint32(c2)
	return lo&0x7f | hi<<7
}

func readBytes(b *bytes.Buffer) ([]byte, error) {
	length := readInt(b)
	data := make([]byte, length)
	_, err := b.Read(data)
	return data, err
}

func makeHelloMessage(publicKey []byte, nonce []byte) ([]byte, error) {
	hello := &Spotify.ClientHello{
		BuildInfo: &Spotify.BuildInfo{
			Product:  Spotify.Product_PRODUCT_PARTNER.Enum(),
			Platform: Spotify.Platform_PLATFORM_IPHONE_ARM.Enum(),
			Version:  proto.Uint64(0x10800000000),
		},
		CryptosuitesSupported: []Spotify.Cryptosuite{
			Spotify.Cryptosuite_CRYPTO_SUITE_SHANNON},
		LoginCryptoHello: &Spotify.LoginCryptoHelloUnion{
			DiffieHellman: &Spotify.LoginCryptoDiffieHellmanHello{
				Gc:              publicKey,
				ServerKeysKnown: proto.Uint32(1),
			},
		},
		ClientNonce: nonce,
		FeatureSet: &Spotify.FeatureSet{
			Autoupdate2: proto.Bool(true),
		},
		Padding: []byte{0x1e},
	}
	return proto.Marshal(hello)
}
