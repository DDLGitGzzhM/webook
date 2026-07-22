package article

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
)

// TestS3 你可以用这个来单独测试你的 OSS 配置对不对，有没有权限
func TestS3(t *testing.T) {
	// 腾讯云中对标 s3 和 OSS 的产品叫做 COS
	cosId, ok := os.LookupEnv("COS_APP_ID")
	if !ok {
		t.Skip("没有找到环境变量 COS_APP_ID")
	}
	cosKey, ok := os.LookupEnv("COS_APP_SECRET")
	if !ok {
		t.Skip("没有找到环境变量 COS_APP_SECRET")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(cosId, cosKey, ""),
		Region:      aws.String("ap-nanjing"),
		Endpoint:    aws.String("https://cos.ap-nanjing.myqcloud.com"),
		// 强制使用 /bucket/key 的形态
		S3ForcePathStyle: aws.Bool(true),
	})
	require.NoError(t, err)
	client := s3.New(sess)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err = client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("webook-1314583317"),
		Key:         aws.String("126"),
		Body:        bytes.NewReader([]byte("测试内容 abc")),
		ContentType: aws.String("text/plain;charset=utf-8"),
	})
	require.NoError(t, err)
	res, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String("webook-1314583317"),
		Key:    aws.String("测试文件"),
	})
	require.NoError(t, err)
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	t.Log(string(data))
}
