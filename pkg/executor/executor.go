package executor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
	RunCommand(cmd []string, cwd string) error
	RunCommandWithOutput(cmd []string, cwd string) (string, error)
}

// CommandExecutorImpl 命令执行器实现
type CommandExecutorImpl struct{}

// NewCommandExecutor 创建新的命令执行器
func NewCommandExecutor() CommandExecutor {
	return &CommandExecutorImpl{}
}

// RunCommand 运行命令（实时输出到控制台）
func (e *CommandExecutorImpl) RunCommand(cmd []string, cwd string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("命令不能为空")
	}

	var command *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows下使用cmd /c
		args := []string{"/c"}
		args = append(args, cmd...)
		command = exec.Command("cmd", args...)
	} else {
		// Unix系统下直接使用命令和参数，不通过shell
		command = exec.Command(cmd[0], cmd[1:]...)
	}

	command.Dir = cwd

	// 将标准输出和标准错误直接连接到控制台，实现实时输出
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// 执行命令
	err := command.Run()
	if err != nil {
		// 构建包含详细信息的错误消息
		cmdStr := strings.Join(cmd, " ")
		return fmt.Errorf("命令执行失败: %s\n工作目录: %s\n命令: %s",
			err.Error(), cwd, cmdStr)
	}

	return nil
}

// RunCommandWithOutput 运行命令并捕获输出
func (e *CommandExecutorImpl) RunCommandWithOutput(cmd []string, cwd string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("命令不能为空")
	}

	var command *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows下使用cmd /c
		args := []string{"/c"}
		args = append(args, cmd...)
		command = exec.Command("cmd", args...)
	} else {
		// Unix系统下直接使用命令和参数，不通过shell
		command = exec.Command(cmd[0], cmd[1:]...)
	}

	command.Dir = cwd

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
