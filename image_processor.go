package main

// 图片处理器
// 环境变量说明：
// - SECRETID: 腾讯云API密钥ID
// - SECRETKEY: 腾讯云API密钥Key
// - COS_BUCKET_URL: 腾讯云COS存储桶URL
// - MINIAPP_PATTERN: 用于匹配URL中的路径模式的正则表达式，例如：`test/\d{8}`

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentyun/cos-go-sdk-v5"
)

func main() {
	// 检查必要的环境变量
	secretID := os.Getenv("SECRETID")
	secretKey := os.Getenv("SECRETKEY")
	if secretID == "" || secretKey == "" {
		fmt.Println("错误：未设置SECRETID或SECRETKEY环境变量")
		return
	}

	// 初始化CDN客户端
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"
	cdnClient, _ := cdn.NewClient(credential, "", cpf)

	// 获取CDN URL列表
	// 获取当天0点时间和当前时间
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	request := cdn.NewListTopDataRequest()
	request.StartTime = common.StringPtr(startTime.Format("2006-01-02 15:04:05"))
	request.EndTime = common.StringPtr(now.Format("2006-01-02 15:04:05"))
	request.Metric = common.StringPtr("url")
	request.Filter = common.StringPtr("flux")
	request.Limit = common.Int64Ptr(100)

	response, err := cdnClient.ListTopData(request)
	if err != nil {
		fmt.Printf("获取CDN URL列表失败: %v\n", err)
		return
	}

	// 从环境变量获取COS桶地址
	cosBucketURL := os.Getenv("COS_BUCKET_URL")
	if cosBucketURL == "" {
		fmt.Println("错误：未设置COS_BUCKET_URL环境变量")
		return
	}

	// 从环境变量获取miniapp路径模式
	miniappPattern := os.Getenv("MINIAPP_PATTERN")
	if miniappPattern == "" {
		fmt.Println("错误：未设置MINIAPP_PATTERN环境变量")
		return
	}

	// 初始化COS客户端
	u, err := url.Parse(cosBucketURL)
	if err != nil {
		fmt.Printf("解析COS桶地址失败: %v\n", err)
		return
	}
	b := &cos.BaseURL{BucketURL: u}
	cosClient := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	// 创建临时目录
	tempDir := "temp_images"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// 处理URL列表
	var urlsToRefresh []string
	for _, detail := range response.Response.Data {
		url := *detail.Resource
		// 只处理jpg和png文件
		if !strings.HasSuffix(strings.ToLower(url), ".jpg") &&
			!strings.HasSuffix(strings.ToLower(url), ".jpeg") &&
			!strings.HasSuffix(strings.ToLower(url), ".png") {
			continue
		}

		// 解析路径信息
		re := regexp.MustCompile(miniappPattern)
		match := re.FindString(url)
		if match == "" {
			continue
		}

		// 下载图片
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("下载图片失败 %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		// 保存原始图片
		filename := filepath.Base(url)
		tempPath := filepath.Join(tempDir, filename)
		tempFile, err := os.Create(tempPath)
		if err != nil {
			fmt.Printf("创建临时文件失败 %s: %v\n", filename, err)
			continue
		}
		_, err = io.Copy(tempFile, resp.Body)
		tempFile.Close()
		if err != nil {
			fmt.Printf("保存图片失败 %s: %v\n", filename, err)
			continue
		}

		// 压缩图片
		srcImg, err := imaging.Open(tempPath)
		if err != nil {
			fmt.Printf("打开图片失败 %s: %v\n", filename, err)
			continue
		}

		// 保持原始尺寸进行压缩
		compressedImg := imaging.Resize(srcImg, srcImg.Bounds().Dx(), srcImg.Bounds().Dy(), imaging.Lanczos)
		compressedPath := filepath.Join(tempDir, "compressed_"+filename)
		err = imaging.Save(compressedImg, compressedPath, imaging.JPEGQuality(60))
		if err != nil {
			fmt.Printf("压缩图片失败 %s: %v\n", filename, err)
			continue
		}

		// 上传到COS
		cosPath := filepath.Join("yasuo", match, filename)
		f, err := os.Open(compressedPath)
		if err != nil {
			fmt.Printf("打开压缩后的图片失败 %s: %v\n", filename, err)
			continue
		}
		_, err = cosClient.Object.Put(context.Background(), cosPath, f, nil)
		f.Close()
		if err != nil {
			fmt.Printf("上传到COS失败 %s: %v\n", filename, err)
			continue
		}

		// 添加到刷新列表
		urlsToRefresh = append(urlsToRefresh, url)
	}

	// 刷新CDN URL
	if len(urlsToRefresh) > 0 {
		refreshRequest := cdn.NewPurgeUrlsCacheRequest()
		refreshRequest.Urls = common.StringPtrs(urlsToRefresh)
		refreshResponse, err := cdnClient.PurgeUrlsCache(refreshRequest)
		if err != nil {
			fmt.Printf("刷新CDN缓存失败: %v\n", err)
			return
		}
		fmt.Printf("成功刷新CDN缓存: %s\n", refreshResponse.ToJsonString())
	}
}