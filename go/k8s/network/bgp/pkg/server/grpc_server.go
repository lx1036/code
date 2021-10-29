package server

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"net"
	"strings"
	"sync"

	api "github.com/osrg/gobgp/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	bgpServer  *BgpServer
	grpcServer *grpc.Server
	hosts      string
}

func newAPIserver(b *BgpServer, g *grpc.Server, hosts string) *Server {
	grpc.EnableTracing = false
	s := &Server{
		bgpServer:  b,
		grpcServer: g,
		hosts:      hosts,
	}
	api.RegisterGobgpApiServer(g, s)
	return s
}

func (s *Server) serve() error {
	var wg sync.WaitGroup
	l := []net.Listener{}
	var err error
	for _, host := range strings.Split(s.hosts, ",") {
		network, address := parseHost(host)
		var lis net.Listener
		lis, err = net.Listen(network, address)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "grpc",
				"Key":   host,
				"Error": err,
			}).Warn("listen failed")
			break
		}
		l = append(l, lis)
	}
	if err != nil {
		for _, lis := range l {
			lis.Close()
		}
		return err
	}

	wg.Add(len(l))
	serve := func(lis net.Listener) {
		defer wg.Done()
		err := s.grpcServer.Serve(lis)
		log.WithFields(log.Fields{
			"Topic": "grpc",
			"Key":   lis,
			"Error": err,
		}).Warn("accept failed")
	}

	for _, lis := range l {
		go serve(lis)
	}
	wg.Wait()
	return nil
}

func parseHost(host string) (string, string) {
	const unixScheme = "unix://"
	if strings.HasPrefix(host, unixScheme) {
		return "unix", host[len(unixScheme):]
	}
	return "tcp", host
}

func (s *Server) StartBgp(context context.Context, request *api.StartBgpRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) StopBgp(context context.Context, request *api.StopBgpRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) GetBgp(context context.Context, request *api.GetBgpRequest) (*api.GetBgpResponse, error) {
	panic("implement me")
}

func (s *Server) AddPeer(context context.Context, request *api.AddPeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeletePeer(context context.Context, request *api.DeletePeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListPeer(request *api.ListPeerRequest, peerServer api.GobgpApi_ListPeerServer) error {
	panic("implement me")
}

func (s *Server) UpdatePeer(context context.Context, request *api.UpdatePeerRequest) (*api.UpdatePeerResponse, error) {
	panic("implement me")
}

func (s *Server) ResetPeer(context context.Context, request *api.ResetPeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ShutdownPeer(context context.Context, request *api.ShutdownPeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) EnablePeer(context context.Context, request *api.EnablePeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DisablePeer(context context.Context, request *api.DisablePeerRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) MonitorPeer(request *api.MonitorPeerRequest, peerServer api.GobgpApi_MonitorPeerServer) error {
	panic("implement me")
}

func (s *Server) AddPeerGroup(context context.Context, request *api.AddPeerGroupRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeletePeerGroup(context context.Context, request *api.DeletePeerGroupRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListPeerGroup(request *api.ListPeerGroupRequest, groupServer api.GobgpApi_ListPeerGroupServer) error {
	panic("implement me")
}

func (s *Server) UpdatePeerGroup(context context.Context, request *api.UpdatePeerGroupRequest) (*api.UpdatePeerGroupResponse, error) {
	panic("implement me")
}

func (s *Server) AddDynamicNeighbor(context context.Context, request *api.AddDynamicNeighborRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListDynamicNeighbor(request *api.ListDynamicNeighborRequest, neighborServer api.GobgpApi_ListDynamicNeighborServer) error {
	panic("implement me")
}

func (s *Server) DeleteDynamicNeighbor(context context.Context, request *api.DeleteDynamicNeighborRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) AddPath(context context.Context, request *api.AddPathRequest) (*api.AddPathResponse, error) {
	panic("implement me")
}

func (s *Server) DeletePath(context context.Context, request *api.DeletePathRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListPath(request *api.ListPathRequest, pathServer api.GobgpApi_ListPathServer) error {
	panic("implement me")
}

func (s *Server) AddPathStream(streamServer api.GobgpApi_AddPathStreamServer) error {
	panic("implement me")
}

func (s *Server) GetTable(context context.Context, request *api.GetTableRequest) (*api.GetTableResponse, error) {
	panic("implement me")
}

func (s *Server) MonitorTable(request *api.MonitorTableRequest, tableServer api.GobgpApi_MonitorTableServer) error {
	panic("implement me")
}

func (s *Server) AddVrf(context context.Context, request *api.AddVrfRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeleteVrf(context context.Context, request *api.DeleteVrfRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListVrf(request *api.ListVrfRequest, vrfServer api.GobgpApi_ListVrfServer) error {
	panic("implement me")
}

func (s *Server) AddPolicy(context context.Context, request *api.AddPolicyRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeletePolicy(context context.Context, request *api.DeletePolicyRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListPolicy(request *api.ListPolicyRequest, policyServer api.GobgpApi_ListPolicyServer) error {
	panic("implement me")
}

func (s *Server) SetPolicies(context context.Context, request *api.SetPoliciesRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) AddDefinedSet(context context.Context, request *api.AddDefinedSetRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeleteDefinedSet(context context.Context, request *api.DeleteDefinedSetRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListDefinedSet(request *api.ListDefinedSetRequest, setServer api.GobgpApi_ListDefinedSetServer) error {
	panic("implement me")
}

func (s *Server) AddStatement(context context.Context, request *api.AddStatementRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeleteStatement(context context.Context, request *api.DeleteStatementRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListStatement(request *api.ListStatementRequest, statementServer api.GobgpApi_ListStatementServer) error {
	panic("implement me")
}

func (s *Server) AddPolicyAssignment(context context.Context, request *api.AddPolicyAssignmentRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeletePolicyAssignment(context context.Context, request *api.DeletePolicyAssignmentRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListPolicyAssignment(request *api.ListPolicyAssignmentRequest, assignmentServer api.GobgpApi_ListPolicyAssignmentServer) error {
	panic("implement me")
}

func (s *Server) SetPolicyAssignment(context context.Context, request *api.SetPolicyAssignmentRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) AddRpki(context context.Context, request *api.AddRpkiRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeleteRpki(context context.Context, request *api.DeleteRpkiRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListRpki(request *api.ListRpkiRequest, rpkiServer api.GobgpApi_ListRpkiServer) error {
	panic("implement me")
}

func (s *Server) EnableRpki(context context.Context, request *api.EnableRpkiRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DisableRpki(context context.Context, request *api.DisableRpkiRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ResetRpki(context context.Context, request *api.ResetRpkiRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) ListRpkiTable(request *api.ListRpkiTableRequest, tableServer api.GobgpApi_ListRpkiTableServer) error {
	panic("implement me")
}

func (s *Server) EnableZebra(context context.Context, request *api.EnableZebraRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) EnableMrt(context context.Context, request *api.EnableMrtRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DisableMrt(context context.Context, request *api.DisableMrtRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) AddBmp(context context.Context, request *api.AddBmpRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) DeleteBmp(context context.Context, request *api.DeleteBmpRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (s *Server) SetLogLevel(context context.Context, request *api.SetLogLevelRequest) (*empty.Empty, error) {
	panic("implement me")
}
