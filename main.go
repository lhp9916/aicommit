package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	configDirName  = ".aicommit"
	configFileName = "config.json"
)

// Config 配置结构体
type Config struct {
	OpenAIEndpoint string  `json:"openai_endpoint"`
	APIKey         string  `json:"api_key"`
	DefaultLang    string  `json:"default_lang"`
	ProxyURL       string  `json:"proxy_url,omitempty"`
	Model          string  `json:"model"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
}

var (
	config     Config
	extraNotes string
)

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []choice       `json:"choices"`
	Error   *errorResponse `json:"error,omitempty"`
}

type choice struct {
	Message message `json:"message"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// 命令行参数结构体
type cmdArgs struct {
	lang     string
	notes    string
	showHelp bool
}

func main() {
	// 先解析命令行参数，只检查 --help
	args := parseArgs()
	if args.showHelp {
		printHelp()
		os.Exit(0)
	}

	// 加载配置文件
	if err := loadConfig(); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 应用命令行参数覆盖配置
	if args.lang != "" {
		config.DefaultLang = args.lang
	}
	extraNotes = args.notes

	// 添加所有更改到暂存区
	runGitCommand("add", ".")
	// 检查 Git 状态
	fmt.Println("Checking the status of the working directory...")
	runGitCommand("status")

	// 获取 Git 差异
	diff := getGitDiff()
	if diff == "" {
		fmt.Println("No differences found.")
		os.Exit(0)
	}

	// 生成提交信息
	commitMessage := generateCommitMessage(diff, config.DefaultLang, extraNotes)
	if commitMessage == "" {
		fmt.Println("Unable to generate commit message.")
		os.Exit(1)
	}

	// 提交更改
	commitChanges(commitMessage)

	fmt.Println("Commit complete with message: ")
	fmt.Println()
	fmt.Println(commitMessage)
}

func parseArgs() cmdArgs {
	args := cmdArgs{
		lang:     "",
		notes:    "",
		showHelp: false,
	}

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			args.showHelp = true
		} else if strings.HasPrefix(arg, "--lang=") {
			args.lang = strings.TrimPrefix(arg, "--lang=")
		} else if strings.HasPrefix(arg, "--notes=") {
			args.notes = strings.TrimPrefix(arg, "--notes=")
		} else {
			fmt.Printf("Unknown parameter passed: %s\n", arg)
			args.showHelp = true
		}
	}

	return args
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// 构建配置目录路径
	configDir := filepath.Join(homeDir, configDirName)

	// 构建配置文件路径
	configPath := filepath.Join(configDir, configFileName)

	return configPath, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	// 默认配置
	defaultConfig := Config{
		OpenAIEndpoint: "https://api.openai.com/v1/chat/completions",
		APIKey:         "",
		DefaultLang:    "en",
		ProxyURL:       "",
		Model:          "gpt-4o",
		MaxTokens:      500,
		Temperature:    0.7,
	}

	// 创建配置目录
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 编码为 JSON
	jsonData, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("默认配置文件已创建: %s\n", configPath)
	fmt.Println("请编辑配置文件设置您的 OpenAI API 密钥")

	return nil
}

// loadConfig 加载配置文件
func loadConfig() error {
	// 获取配置文件路径
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建默认配置文件
		if err := createDefaultConfig(configPath); err != nil {
			return err
		}
		// 配置文件已创建，但没有API密钥，提示用户编辑
		os.Exit(0)
	}

	// 读取配置文件
	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 解析JSON
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return err
	}

	// 验证配置
	if config.APIKey == "" {
		fmt.Println("错误: 配置文件中未设置 API 密钥")
		fmt.Printf("请编辑配置文件: %s\n", configPath)
		os.Exit(1)
	}

	if config.OpenAIEndpoint == "" {
		config.OpenAIEndpoint = "https://api.openai.com/v1/chat/completions"
	}

	if config.DefaultLang == "" {
		config.DefaultLang = "en"
	}

	if config.Model == "" {
		config.Model = "gpt-4o"
	}

	if config.MaxTokens <= 0 {
		config.MaxTokens = 500
	}

	if config.Temperature <= 0 {
		config.Temperature = 0.7
	}

	return nil
}

func printHelp() {
	fmt.Println("AI Commit - 使用 AI 生成 Git 提交信息的工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  aicommit [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -h, --help     显示帮助信息")
	fmt.Println("  --lang=<lang>  设置提交信息的语言 (默认从配置文件读取)")
	fmt.Println("  --notes=<text> 添加额外备注")
	fmt.Println()
	fmt.Println("配置文件:")
	fmt.Println("  ~/.aicommit/config.json")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  aicommit")
	fmt.Println("  aicommit --lang=zh")
	fmt.Println("  aicommit --lang=zh --notes=紧急修复")
}

func runGitCommand(args ...string) string {
	cmd := exec.Command("git", args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running git command: %v\n", err)
		os.Exit(1)
	}

	return output.String()
}

func getGitDiff() string {
	// 获取工作目录差异
	workingDiff := runGitCommand("diff")
	// 获取暂存区差异
	stagedDiff := runGitCommand("diff", "--cached")

	return workingDiff + stagedDiff
}

func generateCommitMessage(diff, lang, notes string) string {
	// 构建请求体
	reqBody := openAIRequest{
		Model: config.Model,
		Messages: []message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze the following code changes and generate a concise Git commit message, providing it in the following languages: %s. Text only: \n\n%s\n\n %s \n\n", lang, diff, notes),
			},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
	}

	// 编码为 JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("POST", config.OpenAIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error calling OpenAI API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// 解析响应
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		os.Exit(1)
	}

	// 检查错误
	if openAIResp.Error != nil {
		fmt.Printf("Error from OpenAI API: %s\n", openAIResp.Error.Message)
		os.Exit(1)
	}

	// 返回生成的提交信息
	if len(openAIResp.Choices) > 0 {
		message := openAIResp.Choices[0].Message.Content
		// 去除可能的引号
		message = strings.TrimPrefix(message, `"`)
		message = strings.TrimSuffix(message, `"`)
		return strings.TrimSpace(message)
	}

	return ""
}

func commitChanges(message string) {
	// 提交更改
	runGitCommand("commit", "-m", message)
}
