# AI Commit

一个使用 AI 生成 Git 提交信息的跨平台命令行工具，基于 Go 语言开发。

## 功能特性

- 跨平台支持（Windows、macOS、Linux）
- 使用 OpenAI API 生成智能提交信息
- 支持多种语言的提交信息
- 支持添加额外备注
- 通过配置文件进行灵活配置
- 自动处理工作目录和暂存区的差异

## 安装

### 从源码构建

1. 确保已安装 Go 1.20 或更高版本
2. 克隆或下载本项目
3. 在项目目录中运行：

```bash
go build -o aicommit main.go
```

4. 将生成的可执行文件添加到系统 PATH 中

## 配置

### 配置文件

程序首次运行时会自动在用户主目录创建配置文件 `~/.aicommit/config.json`。您需要编辑此文件设置 OpenAI API 密钥等参数。

### 配置项说明

| 配置项 | 类型 | 描述 | 默认值 | 示例 |
|--------|------|------|--------|------|
| `openai_endpoint` | string | OpenAI API 端点 | `https://api.openai.com/v1/chat/completions` | `https://api.openai.com/v1/chat/completions` |
| `api_key` | string | OpenAI API 密钥 | 必填 | `sk-xxx` |
| `default_lang` | string | 默认提交信息语言 | `en` | `zh` |
| `proxy_url` | string | 代理 URL（可选） | 空 | `http://proxy.example.com:8080` |
| `model` | string | 使用的模型名称 | `gpt-4o` | `gpt-3.5-turbo` |
| `max_tokens` | integer | 生成的最大令牌数 | `500` | `1000` |
| `temperature` | number | 生成温度，控制创意程度 | `0.7` | `0.5` |

### 配置文件示例

```json
{
  "openai_endpoint": "https://api.openai.com/v1/chat/completions",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "default_lang": "zh",
  "proxy_url": "",
  "model": "gpt-4o",
  "max_tokens": 500,
  "temperature": 0.7
}
```

## 使用方法

在 Git 仓库目录中运行：

```bash
aicommit
```

### 命令行参数

| 参数 | 描述 | 示例 |
|------|------|------|
| `-h, --help` | 显示帮助信息 | `aicommit --help` |
| `--lang=<lang>` | 设置提交信息的语言（覆盖配置文件） | `aicommit --lang=en` |
| `--notes=<text>` | 添加额外备注 | `aicommit --notes="修复了一个关键 bug"` |

### 示例

```bash
# 使用配置文件中的默认语言生成提交信息
aicommit

# 生成英文提交信息
aicommit --lang=en

# 生成中文提交信息并添加额外备注
aicommit --lang=zh --notes="紧急修复" 
```

## 工作原理

1. 解析命令行参数
2. 从配置文件读取配置
3. 检查 Git 仓库状态
4. 获取工作目录和暂存区的差异
5. 调用 OpenAI API 生成提交信息
6. 将所有更改添加到暂存区
7. 使用生成的信息提交更改

## 首次使用

1. 首次运行程序时，会自动创建配置文件 `~/.aicommit/config.json`
2. 编辑配置文件，设置您的 OpenAI API 密钥
3. 再次运行程序，即可正常使用

## 注意事项

- 本工具依赖 Git 命令行工具，请确保已安装 Git
- 请确保您的 OpenAI API 密钥有足够的余额
- 生成的提交信息可能需要手动调整，建议在提交前检查
- 请妥善保管您的 API 密钥，不要泄露给他人

## 许可证

MIT
