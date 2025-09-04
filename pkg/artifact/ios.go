package artifact

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateIPA 验证iOS IPA文件
func (v *ArtifactValidatorImpl) ValidateIPA(ipaPath string, config *ArtifactConfig) (*ValidationResult, error) {
	var details []ValidationDetail
	var success = true

	// 1. 检查IPA文件是否存在
	exists, fileInfo, err := v.checkFileExists(ipaPath)
	if err != nil {
		details = append(details, ValidationDetail{
			Check:    "IPA文件存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("检查文件状态失败: %v", err),
			Critical: true,
		})
		return v.createValidationResult(false, ipaPath, 0, details, err), err
	}

	if !exists {
		details = append(details, ValidationDetail{
			Check:    "IPA文件存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("IPA文件不存在: %s", ipaPath),
			Critical: true,
		})
		return v.createValidationResult(false, ipaPath, 0, details, fmt.Errorf("IPA文件不存在")), fmt.Errorf("IPA文件不存在")
	}

	details = append(details, ValidationDetail{
		Check:    "IPA文件存在性检查",
		Status:   "success",
		Message:  "IPA文件存在",
		Critical: true,
	})

	fileSize := fileInfo.Size()

	// 2. 检查文件大小
	sizeOK, sizeMsg := v.checkFileSize(fileSize, config.MinFileSize, config.MaxFileSize)
	if !sizeOK {
		details = append(details, ValidationDetail{
			Check:    "IPA文件大小检查",
			Status:   "failed",
			Message:  sizeMsg,
			Critical: true,
		})
		success = false
	} else {
		details = append(details, ValidationDetail{
			Check:    "IPA文件大小检查",
			Status:   "success",
			Message:  sizeMsg,
			Critical: false,
		})
	}

	// 3. IPA完整性检查（如果启用）
	if config.ValidateIntegrity || (config.ValidationConfig != nil && config.ValidationConfig.EnableIntegrityCheck) {
		integrityOK, integrityMsg := v.validateIPAIntegrity(ipaPath)
		if !integrityOK {
			details = append(details, ValidationDetail{
				Check:    "IPA完整性检查",
				Status:   "failed",
				Message:  integrityMsg,
				Critical: true,
			})
			success = false
		} else {
			details = append(details, ValidationDetail{
				Check:    "IPA完整性检查",
				Status:   "success",
				Message:  integrityMsg,
				Critical: false,
			})
		}
	}

	var resultErr error
	if !success {
		resultErr = fmt.Errorf("IPA验证失败")
	}

	return v.createValidationResult(success, ipaPath, fileSize, details, resultErr), resultErr
}

// ValidateIOSApp 验证iOS App目录
func (v *ArtifactValidatorImpl) ValidateIOSApp(appPath string, config *ArtifactConfig) (*ValidationResult, error) {
	var details []ValidationDetail
	var success = true

	// 1. 检查App目录是否存在
	exists, fileInfo, err := v.checkFileExists(appPath)
	if err != nil {
		details = append(details, ValidationDetail{
			Check:    "iOS App目录存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("检查目录状态失败: %v", err),
			Critical: true,
		})
		return v.createValidationResult(false, appPath, 0, details, err), err
	}

	if !exists || !fileInfo.IsDir() {
		details = append(details, ValidationDetail{
			Check:    "iOS App目录存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("iOS App目录不存在: %s", appPath),
			Critical: true,
		})
		return v.createValidationResult(false, appPath, 0, details, fmt.Errorf("iOS App目录不存在")), fmt.Errorf("iOS App目录不存在")
	}

	details = append(details, ValidationDetail{
		Check:    "iOS App目录存在性检查",
		Status:   "success",
		Message:  "iOS App目录存在",
		Critical: true,
	})

	// 2. 计算目录大小
	dirSize, err := v.calculateDirectorySize(appPath)
	if err != nil {
		details = append(details, ValidationDetail{
			Check:    "iOS App目录大小计算",
			Status:   "warning",
			Message:  fmt.Sprintf("无法计算目录大小: %v", err),
			Critical: false,
		})
	} else {
		// 检查目录大小
		sizeOK, sizeMsg := v.checkFileSize(dirSize, config.MinFileSize, config.MaxFileSize)
		if !sizeOK {
			details = append(details, ValidationDetail{
				Check:    "iOS App目录大小检查",
				Status:   "failed",
				Message:  sizeMsg,
				Critical: true,
			})
			success = false
		} else {
			details = append(details, ValidationDetail{
				Check:    "iOS App目录大小检查",
				Status:   "success",
				Message:  sizeMsg,
				Critical: false,
			})
		}
	}

	// 3. 检查必要文件
	requiredFiles := []string{"Info.plist", "Runner"}
	for _, requiredFile := range requiredFiles {
		filePath := filepath.Join(appPath, requiredFile)
		if fileExists, _, _ := v.checkFileExists(filePath); fileExists {
			details = append(details, ValidationDetail{
				Check:    fmt.Sprintf("必要文件检查 (%s)", requiredFile),
				Status:   "success",
				Message:  fmt.Sprintf("%s 文件存在", requiredFile),
				Critical: true,
			})
		} else {
			details = append(details, ValidationDetail{
				Check:    fmt.Sprintf("必要文件检查 (%s)", requiredFile),
				Status:   "failed",
				Message:  fmt.Sprintf("%s 文件不存在", requiredFile),
				Critical: true,
			})
			success = false
		}
	}

	// 4. 检查可执行文件权限（非关键）
	executablePath := filepath.Join(appPath, "Runner")
	if execExists, execInfo, _ := v.checkFileExists(executablePath); execExists {
		if execInfo.Mode()&0111 != 0 {
			details = append(details, ValidationDetail{
				Check:    "可执行文件权限检查",
				Status:   "success",
				Message:  "Runner可执行文件权限正常",
				Critical: false,
			})
		} else {
			details = append(details, ValidationDetail{
				Check:    "可执行文件权限检查",
				Status:   "warning",
				Message:  "Runner可执行文件可能没有执行权限",
				Critical: false,
			})
		}
	}

	// 5. 检查Frameworks目录（非关键）
	frameworksPath := filepath.Join(appPath, "Frameworks")
	if frameworksExists, _, _ := v.checkFileExists(frameworksPath); frameworksExists {
		details = append(details, ValidationDetail{
			Check:    "Frameworks目录检查",
			Status:   "success",
			Message:  "Frameworks目录存在",
			Critical: false,
		})
	} else {
		details = append(details, ValidationDetail{
			Check:    "Frameworks目录检查",
			Status:   "warning",
			Message:  "Frameworks目录不存在（可能是静态链接应用）",
			Critical: false,
		})
	}

	var resultErr error
	if !success {
		resultErr = fmt.Errorf("iOS App验证失败")
	}

	return v.createValidationResult(success, appPath, dirSize, details, resultErr), resultErr
}

// validateIPAIntegrity 验证IPA文件完整性
func (v *ArtifactValidatorImpl) validateIPAIntegrity(ipaPath string) (bool, string) {
	// 尝试打开IPA文件作为ZIP
	zipReader, err := zip.OpenReader(ipaPath)
	if err != nil {
		return false, fmt.Sprintf("无法打开IPA文件: %v", err)
	}
	defer zipReader.Close()

	var hasPayload, hasAppBundle bool
	var fileCount int
	var appBundleName string

	// 检查IPA内必要文件
	for _, file := range zipReader.File {
		fileCount++
		fileName := file.Name

		switch {
		case strings.HasPrefix(fileName, "Payload/"):
			hasPayload = true
			// 检查是否有.app包
			if strings.Contains(fileName, ".app/") || strings.HasSuffix(fileName, ".app") {
				hasAppBundle = true
				parts := strings.Split(fileName, "/")
				if len(parts) >= 2 {
					// 获取 .app 包名
					for _, part := range parts {
						if strings.HasSuffix(part, ".app") {
							appBundleName = part
							break
						}
					}
				}
			}
		}
	}

	// 验证必要结构
	var missingComponents []string
	if !hasPayload {
		missingComponents = append(missingComponents, "Payload目录")
	}
	if !hasAppBundle {
		missingComponents = append(missingComponents, ".app包")
	}

	if len(missingComponents) > 0 {
		return false, fmt.Sprintf("缺少必要组件: %s", strings.Join(missingComponents, ", "))
	}

	// 检查文件数量（基本合理性检查）
	if fileCount < 10 {
		return false, fmt.Sprintf("IPA文件数量过少 (%d)，可能损坏", fileCount)
	}

	details := []string{
		"ZIP结构正常",
		"Payload目录存在",
	}

	if appBundleName != "" {
		details = append(details, fmt.Sprintf("%s 应用包存在", appBundleName))
	} else {
		details = append(details, ".app 应用包存在")
	}

	return true, fmt.Sprintf("IPA完整性正常 (%s)", strings.Join(details, ", "))
}

// calculateDirectorySize 计算目录总大小
func (v *ArtifactValidatorImpl) calculateDirectorySize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}