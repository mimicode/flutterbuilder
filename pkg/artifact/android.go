package artifact

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateAPK 验证Android APK文件
func (v *ArtifactValidatorImpl) ValidateAPK(apkPath string, config *ArtifactConfig) (*ValidationResult, error) {
	var details []ValidationDetail
	var success = true

	// 1. 检查APK文件是否存在
	exists, fileInfo, err := v.checkFileExists(apkPath)
	if err != nil {
		details = append(details, ValidationDetail{
			Check:    "APK文件存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("检查文件状态失败: %v", err),
			Critical: true,
		})
		return v.createValidationResult(false, apkPath, 0, details, err), err
	}

	if !exists {
		details = append(details, ValidationDetail{
			Check:    "APK文件存在性检查",
			Status:   "failed",
			Message:  fmt.Sprintf("APK文件不存在: %s", apkPath),
			Critical: true,
		})
		return v.createValidationResult(false, apkPath, 0, details, fmt.Errorf("APK文件不存在")), fmt.Errorf("APK文件不存在")
	}

	details = append(details, ValidationDetail{
		Check:    "APK文件存在性检查",
		Status:   "success",
		Message:  "APK文件存在",
		Critical: true,
	})

	fileSize := fileInfo.Size()

	// 2. 检查文件大小
	sizeOK, sizeMsg := v.checkFileSize(fileSize, config.MinFileSize, config.MaxFileSize)
	if !sizeOK {
		details = append(details, ValidationDetail{
			Check:    "APK文件大小检查",
			Status:   "failed",
			Message:  sizeMsg,
			Critical: true,
		})
		success = false
	} else {
		details = append(details, ValidationDetail{
			Check:    "APK文件大小检查",
			Status:   "success",
			Message:  sizeMsg,
			Critical: false,
		})
	}

	// 3. 检查调试信息目录（非关键）
	debugInfoPath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(apkPath)))), "debug-info")
	if debugExists, _, _ := v.checkFileExists(debugInfoPath); debugExists {
		details = append(details, ValidationDetail{
			Check:    "调试信息目录检查",
			Status:   "success",
			Message:  "调试信息目录存在",
			Critical: false,
		})
	} else {
		details = append(details, ValidationDetail{
			Check:    "调试信息目录检查",
			Status:   "warning",
			Message:  "调试信息目录不存在（可能未启用代码混淆）",
			Critical: false,
		})
	}

	// 4. APK完整性检查（如果启用）
	if config.ValidateIntegrity || (config.ValidationConfig != nil && config.ValidationConfig.EnableIntegrityCheck) {
		integrityOK, integrityMsg := v.validateAPKIntegrity(apkPath)
		if !integrityOK {
			details = append(details, ValidationDetail{
				Check:    "APK完整性检查",
				Status:   "failed",
				Message:  integrityMsg,
				Critical: true,
			})
			success = false
		} else {
			details = append(details, ValidationDetail{
				Check:    "APK完整性检查",
				Status:   "success",
				Message:  integrityMsg,
				Critical: false,
			})
		}
	}

	var resultErr error
	if !success {
		resultErr = fmt.Errorf("APK验证失败")
	}

	return v.createValidationResult(success, apkPath, fileSize, details, resultErr), resultErr
}

// validateAPKIntegrity 验证APK文件完整性
func (v *ArtifactValidatorImpl) validateAPKIntegrity(apkPath string) (bool, string) {
	// 尝试打开APK文件作为ZIP
	zipReader, err := zip.OpenReader(apkPath)
	if err != nil {
		return false, fmt.Sprintf("无法打开APK文件: %v", err)
	}
	defer zipReader.Close()

	var hasAndroidManifest, hasClassesDex, hasResources bool
	var fileCount int

	// 检查APK内必要文件
	for _, file := range zipReader.File {
		fileCount++
		fileName := strings.ToLower(file.Name)

		switch {
		case fileName == "androidmanifest.xml":
			hasAndroidManifest = true
		case strings.HasPrefix(fileName, "classes") && strings.HasSuffix(fileName, ".dex"):
			hasClassesDex = true
		case fileName == "resources.arsc":
			hasResources = true
		}
	}

	// 验证必要文件
	var missingFiles []string
	if !hasAndroidManifest {
		missingFiles = append(missingFiles, "AndroidManifest.xml")
	}
	if !hasClassesDex {
		missingFiles = append(missingFiles, "classes.dex")
	}

	if len(missingFiles) > 0 {
		return false, fmt.Sprintf("缺少必要文件: %s", strings.Join(missingFiles, ", "))
	}

	// 检查文件数量（基本合理性检查）
	if fileCount < 10 {
		return false, fmt.Sprintf("APK文件数量过少 (%d)，可能损坏", fileCount)
	}

	details := []string{
		"ZIP结构正常",
		"AndroidManifest.xml 存在",
		"classes.dex 存在",
	}

	if hasResources {
		details = append(details, "resources.arsc 存在")
	}

	return true, fmt.Sprintf("APK完整性正常 (%s)", strings.Join(details, ", "))
}