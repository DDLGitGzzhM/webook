package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *clientv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := clientv3.NewFromURL("http://localhost:12379")
	assert.NoError(s.T(), err)
	s.cli = cli
}

// TestEtcdServer 测试 ETCD 作为服务注册与发现中心
// 这是服务端的注册过程
func (s *EtcdTestSuite) TestEtcdServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli,
		"service/user")
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 要以 /service/user 为前缀
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	// 5s
	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	// metadata 一般用在客户端
	err = em.AddEndpoint(ctx, key,
		endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))
	assert.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		for resp := range ch {
			t.Log(resp.String())
		}
	}()
	require.NoError(t, err)
	ticker := time.NewTicker(time.Second)
	go func() {
		// 模拟注册的元数据变化
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err1 := em.Update(ctx1, []*endpoints.UpdateWithOpts{
				endpoints.NewAddUpdateOpts(key, endpoints.Endpoint{
					Addr:     addr,
					Metadata: now,
				}, clientv3.WithLease(leaseResp.ID)),
			})
			cancel1()
			require.NoError(t, err1)
		}
	}()

	server := grpc.NewServer()
	l, err := net.Listen("tcp", ":8090")
	assert.NoError(t, err)

	us := &Server{}
	RegisterUserServiceServer(server, us)
	err = server.Serve(l)
	t.Log(err)

	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	assert.NoError(t, err)
	server.GracefulStop()
}

// TestEtcdClient 测试 ETCD 作为服务注册与发现中心
// 这是客户端的发现过程
func (s *EtcdTestSuite) TestEtcdClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	assert.NoError(t, err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	userClient := NewUserServiceClient(cc)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	ctx := context.Background()
	resp, err := userClient.GetById(ctx, &GetByIdReq{
		Id: 123,
	})
	require.NoError(t, err)
	t.Log(resp.User)
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
