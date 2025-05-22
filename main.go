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

	"github.com/disintegration/imaging"
)

func main() {
	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 读取目录中的所有文件
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
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
		srcPath := filepath.Join(dir, filename)
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

		// 压缩图片
		resized := imaging.Resize(img, 0, 0, imaging.Lanczos)

		// 生成压缩后的文件名
		base := strings.TrimSuffix(filename, ext)
		dstPath := filepath.Join(dir, base+"_compressed"+ext)

		// 保存压缩后的图片
		err = imaging.Save(resized, dstPath, imaging.JPEGQuality(80))
		if err != nil {
			fmt.Printf("保存压缩后的图片 %s 失败: %v\n", dstPath, err)
			continue
		}

		fmt.Printf("成功压缩图片: %s -> %s\n", filename, filepath.Base(dstPath))
	}
}