# yasuo

一个简单的图片压缩工具，支持压缩 JPG 和 PNG 格式的图片。

## 功能特点

- 支持 JPG 和 PNG 格式图片压缩
- 自动遍历当前目录下的所有图片文件
- 保持原始图片不变，生成新的压缩后文件
- 压缩后的文件名添加 "_compressed" 后缀

# yasuo
## 使用方法

1. 确保已安装 Go 环境
2. 克隆仓库到本地
3. 进入项目目录
4. 设置必要的环境变量：
   ```bash
   # 腾讯云API密钥
   export SECRETID="你的腾讯云SecretId"
   export SECRETKEY="你的腾讯云SecretKey"
   
   # COS存储桶URL
   export COS_BUCKET_URL="https://你的存储桶地址.cos.ap-region.myqcloud.com"
   
   # 匹配URL中的路径模式的正则表达式
   export MINIAPP_PATTERN="test/\\d{8}"
   ```
5. 运行程序：
   ```bash
   go run image_processor.go
   ```

## 依赖

- github.com/disintegration/imaging v1.6.2

## 注意事项

- 程序会在当前目录下生成压缩后的图片文件
- 压缩后的图片质量设置为80%
- 支持的图片格式：.jpg、.jpeg、.png