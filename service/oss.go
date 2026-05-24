package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/logger"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// OSS 阿里云 OSS 对象存储服务
type OSS struct {
	client     *oss.Client
	bucketName string
	endpoint   string
	useLocal   bool   // OSS_ACCESS_KEY_ID == "no" 时使用本地存储
	localBase  string // 本地存储根目录
}

// NewOSS 初始化 OSS 客户端
func NewOSS() *OSS {
	// 当 OSS_ACCESS_KEY_ID 配置为 "no" 时，使用本地存储
	if config.Config.OSS_ACCESS_KEY_ID == "no" {
		return &OSS{
			useLocal:  true,
			localBase: "./uploads",
		}
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				config.Config.OSS_ACCESS_KEY_ID,
				config.Config.OSS_ACCESS_KEY_SECRET,
			),
		).
		WithRegion(config.Config.OSS_REGION)

	client := oss.NewClient(cfg)

	bucketName := config.Config.OSS_BUCKET_NAME
	endpoint := fmt.Sprintf("https://%s.oss.%s.aliyuncs.com", bucketName, config.Config.OSS_REGION)

	return &OSS{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
	}
}

// UploadFile 上传文件到 OSS 或本地，返回可访问的 URL
func (o *OSS) UploadFile(file *multipart.FileHeader, subDir string) (string, error) {
	if o.useLocal {
		return o.uploadLocal(file, subDir)
	}
	return o.uploadOSS(file, subDir)
}

// uploadLocal 保存文件到本地磁盘
func (o *OSS) uploadLocal(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 确保目录存在
	dir := filepath.Join(o.localBase, subDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(dir, filename)

	// 写入文件
	dst, err := os.Create(savePath)
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	// 返回相对路径作为 URL（前端通过 /uploads/ 静态路由访问）
	url := "/uploads/" + subDir + "/" + filename
	logger.Infof("local upload success: %s\n", url)
	return url, nil
}

// uploadOSS 上传文件到阿里云 OSS
func (o *OSS) uploadOSS(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 读取文件内容到内存
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	// 生成 OSS 中的对象路径
	ext := filepath.Ext(file.Filename)
	objectKey := fmt.Sprintf("%s/%d%s", subDir, time.Now().UnixNano(), ext)

	// 上传到 OSS
	putRequest := &oss.PutObjectRequest{
		Bucket:        oss.Ptr(o.bucketName),
		Key:           oss.Ptr(objectKey),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentLength: oss.Ptr(int64(buf.Len())),
	}

	result, err := o.client.PutObject(context.Background(), putRequest)
	if err != nil {
		logger.Errorf("OSS upload failed: %v", err)
		return "", common.ErrNew(err, common.SysErr)
	}
	logger.Infof("OSS upload success, etag: %v\n", oss.ToString(result.ETag))

	// 返回公网可访问的 URL
	url := fmt.Sprintf("%s/%s", o.endpoint, objectKey)
	return url, nil
}

// CreateBucket 创建 OSS 桶（通常只需执行一次）
func (o *OSS) CreateBucket(ctx context.Context, bucketName string) error {
	request := &oss.PutBucketRequest{
		Bucket: oss.Ptr(bucketName),
	}

	_, err := o.client.PutBucket(ctx, request)
	if err != nil {
		logger.Errorf("failed to put bucket %v", err)
		return common.ErrNew(err, common.SysErr)
	}

	logger.Infof("bucket %s created successfully\n", bucketName)
	return nil
}

// GetObject 从 OSS 或本地获取文件对象，返回 Reader、Content-Type 和文件大小
func (o *OSS) GetObject(objectKey string) (io.ReadCloser, string, int64, error) {
	if o.useLocal {
		return o.getLocal(objectKey)
	}
	return o.getOSS(objectKey)
}

// getLocal 从本地磁盘读取文件
func (o *OSS) getLocal(objectKey string) (io.ReadCloser, string, int64, error) {
	// objectKey 格式: /uploads/photos/xxx.jpg → 转为本地路径
	filePath := filepath.Join(o.localBase, objectKey[len("/uploads/"):])

	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("local get object failed: %v", err)
		return nil, "", 0, fmt.Errorf("local get object failed: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, "", 0, fmt.Errorf("local stat failed: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return file, contentType, stat.Size(), nil
}

// getOSS 从阿里云 OSS 获取文件对象
func (o *OSS) getOSS(objectKey string) (io.ReadCloser, string, int64, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(objectKey),
	}

	result, err := o.client.GetObject(context.Background(), request)
	if err != nil {
		logger.Errorf("OSS get object failed: %v", err)
		return nil, "", 0, fmt.Errorf("OSS get object failed: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	contentLength := result.ContentLength

	return result.Body, contentType, contentLength, nil
}

// ExtractObjectKey 从 URL 中提取 object key（本地模式下直接返回原始路径）
func (o *OSS) ExtractObjectKey(rawURL string) string {
	if o.useLocal {
		return rawURL
	}
	prefix := fmt.Sprintf("%s/", o.endpoint)
	if len(rawURL) > len(prefix) && rawURL[:len(prefix)] == prefix {
		return rawURL[len(prefix):]
	}
	return rawURL
}
