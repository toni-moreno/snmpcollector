package mock

import (
	"net"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/gosnmp/gosnmp"
)

var (
	log *logrus.Logger
)

// SetLogger xx
func SetLogger(l *logrus.Logger) {
	log = l
}

// SnmpServer mock object
type SnmpServer struct {
	Listen      string
	SnmpVersion gosnmp.SnmpVersion
	Community   string
	pc          net.PacketConn
	quit        bool
	qMutex      sync.RWMutex
	Want        []gosnmp.SnmpPDU
}

func (s *SnmpServer) ResponseForPkt(i *gosnmp.SnmpPacket) (*gosnmp.SnmpPacket, error) {
	// Find for which SubAgent

	switch i.PDUType {
	case gosnmp.GetRequest:
		i.PDUType = gosnmp.GetResponse
		log.Infof("GET REQUEST")
		for k, v := range i.Variables {
			found := false
			findex := -1
			for ke, ve := range s.Want {
				if ve.Name == v.Name {
					found = true
					findex = ke
					break
				}
			}
			if found {
				i.Variables[k] = s.Want[findex]
				log.Infof("found response value %v for type %x for Name %s", i.Variables[k].Value, i.Variables[k].Type, i.Variables[k].Name)
			} else {
				log.Warnf("not found value for request %d , name %s", k, v.Name)
			}

		}

	case gosnmp.GetNextRequest:
		log.Infof("GET NEXT")
		fallthrough
	case gosnmp.GetBulkRequest:
		log.Infof("GET BULK")
		i.PDUType = gosnmp.GetResponse
		length := len(i.Variables)
		queryForOid := i.Variables[length-1].Name
		var result []gosnmp.SnmpPDU
		for _, v := range i.Variables {
			for _, vw := range s.Want {
				if strings.HasPrefix(vw.Name, v.Name) {
					result = append(result, vw)

				}
			}
		}
		result = append(result, gosnmp.SnmpPDU{Name: queryForOid, Type: gosnmp.EndOfMibView, Value: nil})
		i.Variables = result
	case gosnmp.SetRequest:
		//return t.serveSetRequest(response)
		i.PDUType = gosnmp.GetResponse
	case gosnmp.Trap, gosnmp.SNMPv2Trap, gosnmp.InformRequest:
		//return t.serveTrap(response)
	default:
		return nil, errors.WithStack(ErrUnsupportedOperation)
	}

	return i, nil

}

var ErrUnsupportedProtoVersion = errors.New("ErrUnsupportedProtoVersion")
var ErrNoSNMPInstance = errors.New("ErrNoSNMPInstance")
var ErrUnsupportedOperation = errors.New("ErrUnsupportedOperation")
var ErrNoPermission = errors.New("ErrNoPermission")
var ErrUnsupportedPacketData = errors.New("ErrUnsupportedPacketData")

func (s *SnmpServer) fillErrorPkt(err error, io *gosnmp.SnmpPacket) error {
	io.PDUType = gosnmp.GetResponse
	if errors.Is(err, ErrNoSNMPInstance) {
		io.Error = gosnmp.NoAccess
	} else if errors.Is(err, ErrUnsupportedOperation) {
		io.Error = gosnmp.ResourceUnavailable
	} else if errors.Is(err, ErrNoPermission) {
		io.Error = gosnmp.AuthorizationError
	} else if errors.Is(err, ErrUnsupportedPacketData) {
		io.Error = gosnmp.BadValue
	} else {
		io.Error = gosnmp.GenErr
	}
	io.ErrorIndex = 0
	return nil
}

func (s *SnmpServer) marshalPkt(pkt *gosnmp.SnmpPacket, err error) ([]byte, error) {
	// when err. marshal error pkt
	if pkt == nil {
		pkt = &gosnmp.SnmpPacket{}
	}
	if err != nil {
		log.Debugf("Will marshal: %v", err)

		errFill := s.fillErrorPkt(err, pkt)
		if errFill != nil {
			return nil, err
		}

		return pkt.MarshalMsg()
	}
	log.Debugf("Marshall PKT: %+v", pkt)
	out, err := pkt.MarshalMsg()
	log.Debugf("Marshall PKT: %+v", out)
	return out, err
}

func (s *SnmpServer) serve(addr net.Addr, buf []byte) {

	var response []byte
	var err error
	vhandle := gosnmp.GoSNMP{}
	vhandle.Logger = log
	request, decodeError := vhandle.SnmpDecodePacket(buf)
	if decodeError != nil {
		log.Errorf("Error on Decode packet %s", decodeError)
		return
	}
	switch request.Version {
	case gosnmp.Version1:
		log.Infof("Got SnmpVersion 1 packet: %+v", request)
		response, err = s.marshalPkt(s.ResponseForPkt(request))
		if err != nil {
			log.Errorf("Error on decode: %s", err)
			return
		}
	case gosnmp.Version2c:
		log.Infof("Got SnmpVersion 2c packet: %+v", request)
		response, err = s.marshalPkt(s.ResponseForPkt(request))
		if err != nil {
			log.Errorf("Error on decode: %s", err)
			return
		}
	case gosnmp.Version3:
		log.Infof("Got SnmpVersion 3 packet: %+v", request)
		log.Errorf("unsupported v3 protocol on mock test")
		return
	default:
		log.Infof("Unknown SnmpVersion for packet: %v", request)
	}

	n, err := s.pc.WriteTo(response, addr)
	if err != nil {
		log.Errorf("Can not write response %s", err)
	}
	log.Infof("OK: sending %d bytes of response", n)
}

func (s *SnmpServer) setFinish() {
	s.qMutex.Lock()
	defer s.qMutex.Unlock()
	s.quit = true
}

func (s *SnmpServer) getFinish() bool {
	s.qMutex.Lock()
	defer s.qMutex.Unlock()
	return s.quit
}

// Start snmp mock server
func (s *SnmpServer) Start() error {
	var err error
	s.pc, err = net.ListenPacket("udp", s.Listen)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	go func() {

		for {
			if s.getFinish() {
				return
			}

			buf := make([]byte, 4096)
			n, addr, err := s.pc.ReadFrom(buf)
			if err != nil {
				log.Errorf("Error on read data: %s", err)
				continue
			}
			log.Infof("Read [%d] from %+v", n, addr)
			go s.serve(addr, buf[:n])
			log.Infof("Next...")
		}
	}()
	return nil
}

// Stop mock server
func (s *SnmpServer) Stop() error {
	s.setFinish()
	//s.quit <- true
	return s.pc.Close()
}
