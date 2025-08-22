# macOS 签名与 ImageMagick 开发完整指南

------

## 1. macOS 签名简介

### 1.1 什么是 macOS 签名（Code Signing）

macOS 签名是苹果为保护系统安全提供的一种机制，核心目的：

- **验证应用/可执行文件的来源**
- **保证文件未被篡改**
- 与 **Gatekeeper** 联动：阻止未经验证的软件运行

签名信息嵌入二进制文件或 bundle 内，通过 **证书 Authority** 确定身份。

------

### 1.2 为什么需要签名

1. **防止 Gatekeeper 弹窗**
   - 未签名或自签名文件在 macOS 上运行时，会弹窗提示“无法验证软件是否包含恶意内容”
2. **保证可分发性**
   - 只有官方 Developer ID + notarization 才能开箱即用，无需用户手动允许
3. **确保动态库调用安全**
   - 对于依赖第三方 dylib 的程序（如 ImageMagick），签名可防止系统拒绝加载

------

## 2. macOS 签名机制

macOS 支持多种签名类型：

| 类型                       | 描述                            | 适用场景               |
| -------------------------- | ------------------------------- | ---------------------- |
| 自签名（Self-signed）      | 自己生成的证书，无苹果官方背书  | 开发阶段测试、内部使用 |
| 官方 Developer ID          | 苹果开发者证书，可 notarization | 正式分发，用户无需弹窗 |
| 临时或自签名 Keychain 证书 | 仅在本机信任                    | 适合多开发者内部测试   |

> 注意：自签名证书在其他 Mac 上仍会被 Gatekeeper 阻止，需要手动允许或移除 quarantine 属性。

------

## 3. 创建自签名证书（开发阶段）

### 3.1 准备工具

- macOS 自带 OpenSSL
- Keychain Access（钥匙串访问）

### 3.2 生成私钥和证书

```
# 生成私钥
openssl genrsa -out mykey.pem 2048

# 生成自签名证书
openssl req -new -x509 -key mykey.pem -out mycert.pem -days 3650
```

- `mykey.pem`：私钥
- `mycert.pem`：证书
- 生成时会提示填写一些信息，可随意填写开发相关内容

------

### 3.3 导入 Keychain

1. 双击 `mycert.pem`
2. 导入 Keychain 时，选择 **登录/系统/本机**（开发测试推荐选择“登录”）
3. 打开 Keychain Access → 找到证书 → 设置 **始终信任**

------

### 3.4 验证签名是否存在

```
security find-identity -p codesigning -v
```

- 如果看到类似：

```
1) 85BBDE294CE5D1F4B8E939F272B30747DA5C86B0 "duyl328"
1 valid identities found
```

说明自签名证书已创建成功，可用于签名。

------

## 4. 使用 codesign 对文件签名

### 4.1 签名可执行文件

```
codesign --force --sign "duyl328" /path/to/executable
```

说明：

- `--force`：覆盖已有签名
- `--sign`：指定证书名称或 SHA-1
- `--deep`：可递归签名 bundle 内所有二进制（适合 .app）

### 4.2 验证签名

```
codesign -dvv /path/to/executable
```

输出示例：

```
Executable=/path/to/executable
Identifier=magick
Format=Mach-O thin (x86_64)
CodeDirectory v=20400 size=447 flags=0x0(none) hashes=9+2 location=embedded
Signature size=1728
Authority=duyl328
```

- `Authority=duyl328` → 签名生效
- `TeamIdentifier` 自签名通常为空

### 4.3 注意事项

- 自签名文件在 **其他 Mac 上仍会弹窗**
- 命令行执行可通过手动允许或移除 quarantine：

```
xattr -r -d com.apple.quarantine /path/to/executable
```

------

## 5. ImageMagick 开发与签名实践

### 5.1 问题回顾

- 下载官方 tar.gz 后，解压 `/Users/xxx/ImageMagick-7.0.10`
- 执行：

```
/Users/xxx/ImageMagick-7.0.10/bin/magick -version
```

出现弹窗 → Gatekeeper 提示
 执行 dyld 找不到 libMagickCore 或 freetype dylib → 动态库路径问题

------

### 5.2 解决步骤

#### 5.2.1 开发阶段

1. **签名 magick 和 lib/\*.dylib**

```
codesign --force --sign "duyl328" /Users/xxx/ImageMagick-7.0.10/bin/magick
codesign --force --sign "duyl328" /Users/xxx/ImageMagick-7.0.10/lib/*.dylib
```

1. **DYLD 临时设置**

```
DYLD_LIBRARY_PATH=/Users/xxx/ImageMagick-7.0.10/lib \
/Users/xxx/ImageMagick-7.0.10/bin/magick -version
```

- 不污染全局环境
- 解决可执行文件找不到 dylib 问题

1. **安装第三方依赖（如 freetype）**

```
brew install freetype
brew install libpng jpeg libtiff
```

- 或安装 XQuartz（自带 freetype 库）

------

#### 5.2.2 打包阶段

1. 将 magick 和 lib/*.dylib 放入 `.app/Contents/MacOS/` 和 `.app/Contents/Frameworks/`
2. 修正 dylib 路径：

```
install_name_tool -change /原始路径/libMagickCore-7.Q16HDRI.8.dylib \
@rpath/libMagickCore-7.Q16HDRI.8.dylib \
/path/to/magick
```

1. **签名整个 bundle**

```
codesign --deep --force --sign "Developer ID Application: YourName" MyApp.app
```

1. **Notarization（苹果服务器验证）**

```
xcrun altool --notarize-app -f MyApp.zip --primary-bundle-id com.your.appid -u your@appleid.com -p app-specific-password
```

- 通过 notarization 后，用户开箱即可运行，无弹窗

------

## 6. 开发阶段实用技巧

- **自签名只在本机有效**
- **DYLD_LIBRARY_PATH** 可以在单条命令中临时指定，不污染全局环境
- **xattr -d com.apple.quarantine** 可移除下载文件的隔离标记
- Homebrew 安装 ImageMagick + 依赖是最简单的开发方式
- 打包分发时需处理动态库路径、签名和 notarization

------

## 7. 总结流程

| 阶段     | 步骤                                                         |
| -------- | ------------------------------------------------------------ |
| 开发测试 | 1. 自签名证书 → 2. codesign 可执行文件 + dylib → 3. 临时 DYLD_LIBRARY_PATH → 4. 安装必要依赖（Homebrew/XQuartz） |
| 分发     | 1. 打包可执行文件 + dylib 到 .app bundle → 2. install_name_tool 调整路径 → 3. codesign --deep 使用官方 Developer ID → 4. Notarization → 5. 生成 dmg 或 zip 分发 |

------

## 8. FAQ

**Q1：自签名文件在其他 Mac 上可用吗？**
 A1：只能运行，但会出现 Gatekeeper 弹窗，需要手动允许或解除 quarantine。

**Q2：为什么 DYLD_LIBRARY_PATH 不全局？**
 A2：开发阶段临时指定即可，不污染用户全局环境，方便测试。

**Q3：为什么 Homebrew 安装的 Magick 可以直接用？**
 A3：Homebrew 自动处理了 dylib 路径和依赖库，符合 macOS rpath 机制。

**Q4：正式发布必须买开发者账号吗？**
 A4：是的，要开箱即用、无弹窗，必须使用 **Developer ID + notarization**。

------

## 9. 参考命令汇总

```
# 生成自签名证书
openssl genrsa -out mykey.pem 2048
openssl req -new -x509 -key mykey.pem -out mycert.pem -days 3650

# 导入钥匙串
open mycert.pem

# 查看签名身份
security find-identity -p codesigning -v

# 签名可执行文件
codesign --force --sign "duyl328" /path/to/executable
codesign --force --sign "duyl328" /path/to/lib/*.dylib

# 验证签名
codesign -dvv /path/to/executable

# 移除 quarantine 属性
xattr -r -d com.apple.quarantine /path/to/executable

# 临时执行 ImageMagick，指定 dylib 路径
DYLD_LIBRARY_PATH=/path/to/lib /path/to/bin/magick -version

# 修改 dylib 依赖路径
install_name_tool -change /旧路径/libX.dylib @rpath/libX.dylib /path/to/magick

# Homebrew 安装 ImageMagick
brew install imagemagick
magick -version
```

------