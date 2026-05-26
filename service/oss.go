package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/logger"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/google/uuid"
)

// OSS 阿里云 OSS 对象存储服务（兼容本地存储）
type OSS struct {
	client     *oss.Client
	bucketName string
	endpoint   string
	useLocal   bool   // 本地存储模式标识
	localBase  string // 本地存储根目录
}

// NewOSS 初始化 OSS 客户端
// 当配置项 OSS_USE_LOCAL 为 true 时，所有操作指向本地磁盘
func NewOSS() *OSS {
	if config.Config.OSS_USE_LOCAL == "true" {
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

// UploadFile 上传文件，返回可访问的 URL
func (o *OSS) UploadFile(file *multipart.FileHeader, subDir string) (string, error) {
	// 统一校验子目录，防止路径穿越
	cleanDir, err := sanitizeSubDir(subDir)
	if err != nil {
		return "", common.ErrNew(err, common.ParamErr)
	}

	if o.useLocal {
		return o.uploadLocal(file, cleanDir)
	}
	return o.uploadOSS(file, cleanDir)
}

// ------------------------- 本地存储 -------------------------

func (o *OSS) uploadLocal(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	dir := filepath.Join(o.localBase, subDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	filename := generateFilename(file.Filename)
	savePath := filepath.Join(dir, filename)

	dst, err := os.Create(savePath)
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)
	logger.Infof("local upload success: %s", url)
	return url, nil
}

func (o *OSS) getLocal(objectKey string) (io.ReadCloser, string, int64, error) {
	// objectKey 预期格式: /uploads/photos/xxx.jpg
	const prefix = "/uploads/"
	if !strings.HasPrefix(objectKey, prefix) {
		return nil, "", 0, fmt.Errorf("invalid local object key: %s", objectKey)
	}

	relPath := objectKey[len(prefix):]
	// 防止相对路径穿越
	cleanPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanPath, "..") {
		return nil, "", 0, fmt.Errorf("invalid local object key: %s", objectKey)
	}
	filePath := filepath.Join(o.localBase, cleanPath)

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

// ------------------------- 阿里云 OSS -------------------------

func (o *OSS) uploadOSS(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	objectKey := fmt.Sprintf("%s/%s", subDir, generateFilename(file.Filename))

	// 根据扩展名设置 Content-Type
	contentType := mime.TypeByExtension(filepath.Ext(file.Filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	putRequest := &oss.PutObjectRequest{
		Bucket:        oss.Ptr(o.bucketName),
		Key:           oss.Ptr(objectKey),
		Body:          src, // 直接使用文件流，避免全部读入内存
		ContentType:   oss.Ptr(contentType),
		ContentLength: oss.Ptr(file.Size),
	}

	result, err := o.client.PutObject(context.Background(), putRequest)
	if err != nil {
		logger.Errorf("OSS upload failed: %v", err)
		return "", common.ErrNew(err, common.SysErr)
	}
	logger.Infof("OSS upload success, etag: %v", oss.ToString(result.ETag))

	url := fmt.Sprintf("%s/%s", o.endpoint, objectKey)
	return url, nil
}

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

	contentType := "application/octet-stream"
	if result.ContentType != nil && *result.ContentType != "" {
		contentType = *result.ContentType
	}

	// contentLength := int64(0)
	// if result.ContentLength != nil {
	// 	contentLength = *result.ContentLength
	// }
	contentLength := result.ContentLength
	return result.Body, contentType, contentLength, nil
}

// CreateBucket 创建 OSS 桶（不支持本地模式）
func (o *OSS) CreateBucket(ctx context.Context, bucketName string) error {
	if o.useLocal {
		return common.ErrNew(fmt.Errorf("CreateBucket not supported in local mode"), common.SysErr)
	}

	request := &oss.PutBucketRequest{
		Bucket: oss.Ptr(bucketName),
	}
	_, err := o.client.PutBucket(ctx, request)
	if err != nil {
		logger.Errorf("failed to put bucket %v", err)
		return common.ErrNew(err, common.SysErr)
	}
	logger.Infof("bucket %s created successfully", bucketName)
	return nil
}

// GetObject 获取文件对象，返回 Reader、Content-Type、文件大小
func (o *OSS) GetObject(objectKey string) (io.ReadCloser, string, int64, error) {
	if o.useLocal {
		return o.getLocal(objectKey)
	}
	return o.getOSS(objectKey)
}

// ExtractObjectKey 从完整 URL 中提取对象存储的 Key
func (o *OSS) ExtractObjectKey(rawURL string) string {
	if o.useLocal {
		return rawURL
	}
	prefix := fmt.Sprintf("%s/", o.endpoint)
	if strings.HasPrefix(rawURL, prefix) {
		return rawURL[len(prefix):]
	}
	// 无法识别则原样返回，后续操作可能失败
	return rawURL
}

// ------------------------- 通用工具方法 -------------------------

// generateFilename 生成唯一的文件名（保留原始扩展名）
func generateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	// 使用纳秒时间戳 + UUID 前8位保证唯一性
	uid := strings.ReplaceAll(uuid.New().String(), "-", "")
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uid[:8], ext)
}

// sanitizeSubDir 清理并校验子目录，禁止路径穿越
func sanitizeSubDir(dir string) (string, error) {
	dir = filepath.Clean(dir)
	if dir == "." {
		dir = ""
	}
	// 禁止绝对路径和上级引用
	if filepath.IsAbs(dir) || strings.HasPrefix(dir, "..") {
		return "", fmt.Errorf("invalid sub directory: %s", dir)
	}
	return dir, nil
}
