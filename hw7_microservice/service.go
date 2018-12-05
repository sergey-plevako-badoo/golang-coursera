package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type myMicroservice struct {
	Acls []*ACL
}

func (s myMicroservice) Check(ctx context.Context, stub *Nothing) (*Nothing, error) {
	fmt.Println("here")

	panic("implement me")
	return nil, nil
}

func (s myMicroservice) Add(ctx context.Context, stub *Nothing) (*Nothing, error) {
	fmt.Println("here")
	return nil, nil
}

func (s myMicroservice) Test(ctx context.Context, stub *Nothing) (*Nothing, error) {

	md, _ := metadata.FromIncomingContext(ctx)

	method, _ := grpc.Method(ctx)

	consumer, ok := md["consumer"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated consumer")
	}

	granted := s.checkPermissions(consumer[0], method)
	if !granted {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated consumer")
	}

	return &Nothing{}, nil
}



func (s myMicroservice) Logging(stub *Nothing, stream Admin_LoggingServer) error {
	fmt.Println("here")
	return nil
}

func (s myMicroservice) Statistics(statInverval *StatInterval, stream Admin_StatisticsServer) error {
	panic("implement me")
}

type ACL struct {
	User string
	Methods []string
}

func (s *myMicroservice) checkPermissions(user string, method string) bool {

	var permissionsGranted = false

	for _, acl := range s.Acls {
		if user == acl.User {
			for _, aclMethod := range acl.Methods {
				if aclMethod == method {
					permissionsGranted = true
				}
			}
		}
	}

	return permissionsGranted
}

func parseACL(aclJSON string) ([]*ACL, error) {
	var rawACL map[string]*json.RawMessage
	var acls []*ACL

	err := json.Unmarshal([]byte(aclJSON), &rawACL)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer("[", "", "]", "", "\"", "")

	for key, value := range rawACL {
		valueJSON, err := json.Marshal(&value)
		if err != nil {
			return nil, err
		}

		strValue := replacer.Replace(string(valueJSON))

		acls = append(acls, &ACL{
			User: key,
			Methods: strings.Split(strValue, ","),
		})
	}

	return acls, nil
}

func StartMyMicroservice(ctx context.Context, address string, aclData string) error {
	acls, err := parseACL(aclData)
	if err != nil {
		return err
	}

	go func() {
		if err := bootServer(ctx, address, acls); err != nil {
			log.Fatalf("Failed to serve %v", err)
		}
	}()

	return nil
}

func bootServer(ctx context.Context, address string, acl []*ACL) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s := grpc.NewServer()

	service := &myMicroservice{
		Acls: acl,
	}

	RegisterAdminServer(s, service)
	RegisterBizServer(s, service)

	go listenShutdownServer(ctx, s)

	err = s.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}

func listenShutdownServer(ctx context.Context, server *grpc.Server) {
	for {
		select {
		case <-ctx.Done():
			server.Stop()
			return
		}
	}
}