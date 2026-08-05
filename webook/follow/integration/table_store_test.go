package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"webook/webook/follow/repository/dao"
)

type TableStoreDAOTestSuite struct {
	suite.Suite
	dao    *dao.TableStoreFollowRelationDao
	client *tablestore.TableStoreClient
}

func (s *TableStoreDAOTestSuite) SetupSuite() {
	endpoint := os.Getenv("TS_ENDPOINT")
	accessId := os.Getenv("TS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("TS_ACCESS_KEY_SECRET")
	instanceName := os.Getenv("TS_INSTANCE_NAME")
	if endpoint == "" || accessId == "" || accessKeySecret == "" || instanceName == "" {
		s.T().Skip("缺少 TableStore 环境变量，跳过测试")
	}
	s.client = tablestore.NewClient(endpoint, instanceName, accessId, accessKeySecret)
	s.InitTable()
	s.dao = dao.NewTableStoreDao(s.client)
}

func (s *TableStoreDAOTestSuite) TearDownSuite() {
	if s.client == nil {
		return
	}
	_, err := s.client.DeleteTable(&tablestore.DeleteTableRequest{
		TableName: dao.FollowRelationTableName,
	})
	require.NoError(s.T(), err)
}

func (s *TableStoreDAOTestSuite) TestAdd() {
	now := time.Now().UnixMilli()
	err := s.dao.CreateFollowRelation(context.Background(), dao.FollowRelation{
		Followee: 12,
		Follower: 13,
		Status:   dao.FollowRelationStatusActive,
		Ctime:    now,
		Utime:    now,
	})
	require.NoError(s.T(), err)
}

func (s *TableStoreDAOTestSuite) TestCntFollowee() {
	now := time.Now().UnixMilli()
	err := s.dao.CreateFollowRelation(context.Background(), dao.FollowRelation{
		Followee: 22,
		Follower: 23,
		Status:   dao.FollowRelationStatusActive,
		Ctime:    now,
		Utime:    now,
	})
	require.NoError(s.T(), err)
	res, err := s.dao.CntFollowee(context.Background(), 23)
	require.NoError(s.T(), err)
	require.Equal(s.T(), int64(1), res)
	res, err = s.dao.CntFollower(context.Background(), 22)
	require.NoError(s.T(), err)
	require.Equal(s.T(), int64(1), res)
}

func TestTableStoreDAO(t *testing.T) {
	suite.Run(t, new(TableStoreDAOTestSuite))
}

func (s *TableStoreDAOTestSuite) InitTable() {
	createTableRequest := new(tablestore.CreateTableRequest)

	tableMeta := new(tablestore.TableMeta)
	// 声明表名
	tableMeta.TableName = dao.FollowRelationTableName

	tableMeta.AddPrimaryKeyColumn("follower", tablestore.PrimaryKeyType_INTEGER)
	tableMeta.AddPrimaryKeyColumn("followee", tablestore.PrimaryKeyType_INTEGER)

	// 添加属性列 备注 创建时间 和更新时间
	tableMeta.AddDefinedColumn("utime", tablestore.DefinedColumn_INTEGER)
	tableMeta.AddDefinedColumn("ctime", tablestore.DefinedColumn_INTEGER)
	tableMeta.AddDefinedColumn("status", tablestore.DefinedColumn_INTEGER)
	tableOption := new(tablestore.TableOption)
	// 数据的过期时间
	tableOption.TimeToAlive = -1
	tableOption.MaxVersion = 1
	reservedThroughput := new(tablestore.ReservedThroughput)
	reservedThroughput.Readcap = 0
	reservedThroughput.Writecap = 0
	createTableRequest.TableMeta = tableMeta
	createTableRequest.TableOption = tableOption
	createTableRequest.ReservedThroughput = reservedThroughput
	_, err := s.client.CreateTable(createTableRequest)
	if err != nil {
		optsErr, ok := err.(*tablestore.OtsError)
		if ok {
			if optsErr.Code == "OTSObjectAlreadyExist" {
				return
			}
		}
		panic(err)
	}
}
