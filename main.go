package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

func main() {
	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 创建old目录
	oldDir := filepath.Join(dir, "old")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		fmt.Printf("创建old目录失败: %v\n", err)
		return
	}

	// 创建以当前时间命名的目录
	timeDir := time.Now().Format("20060102150405")
	dstDir := filepath.Join(dir, timeDir)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		fmt.Printf("创建目标目录失败: %v\n", err)
		return
	}

	// 读取old目录中的所有文件
	files, err := ioutil.ReadDir(oldDir)
	if err != nil {
		fmt.Printf("读取old目录失败: %v\n", err)
		return
	}

	// 遍历所有文件
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// 只处理jpg和png文件
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		// 打开原始图片
		srcPath := filepath.Join(oldDir, filename)
		srcFile, err := os.Open(srcPath)
		if err != nil {
			fmt.Printf("打开文件 %s 失败: %v\n", filename, err)
			continue
		}
		defer srcFile.Close()

		// 解码图片
		var img image.Image
		if ext == ".png" {
			img, err = png.Decode(srcFile)
		} else {
			img, err = jpeg.Decode(srcFile)
		}
		if err != nil {
			fmt.Printf("解码图片 %s 失败: %v\n", filename, err)
			continue
		}

		// 获取原始图片尺寸
		bounds := img.Bounds()
		width := bounds.Max.X
		height := bounds.Max.Y

		// 压缩图片，保持原始尺寸
		resized := imaging.Resize(img, width, height, imaging.Lanczos)

		// 生成压缩后的文件名
		dstPath := filepath.Join(dstDir, filename)

		// 根据文件格式选择保存方式
		if ext == ".png" {
			err = imaging.Save(resized, dstPath, imaging.PNGCompressionLevel(9))
		} else {
			err = imaging.Save(resized, dstPath, imaging.JPEGQuality(50))
		}
		if err != nil {
			fmt.Printf("保存压缩后的图片 %s 失败: %v\n", dstPath, err)
			continue
		}

		// 获取原始文件和压缩后文件的大小
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			fmt.Printf("获取原始文件大小失败: %v\n", err)
			continue
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			fmt.Printf("获取压缩后文件大小失败: %v\n", err)
			continue
		}

		// 如果压缩后的文件更大，则删除压缩后的文件
		if dstInfo.Size() >= srcInfo.Size() {
			if err := os.Remove(dstPath); err != nil {
				fmt.Printf("删除压缩后的文件失败: %v\n", err)
			}
			fmt.Printf("跳过压缩图片 %s (压缩后文件更大)\n", filename)
			continue
		}

		fmt.Printf("成功压缩图片: %s (原始大小: %d bytes, 压缩后大小: %d bytes)\n", filename, srcInfo.Size(), dstInfo.Size())
	}
}