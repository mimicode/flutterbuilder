package artifact

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mimicode/flutterbuilder/pkg/types"
)

func TestArtifactValidator_ValidateAPK(t *testing.T) {
	validator := NewArtifactValidator()

	t.Run("APK文件不存在", func(t *testing.T) {
		config := &ArtifactConfig{
			Platform:          PlatformAPK,
			SourcePath:        "/nonexistent",
			MinFileSize:       DefaultAndroidMinSize,
			MaxFileSize:       DefaultAndroidMaxSize,
			ValidateIntegrity: false,
		}

		result, err := validator.ValidateAPK("/nonexistent/app.apk", config)
		if err == nil {
			t.Error("预期应该返回错误")
		}
		if result.Success {
			t.Error("预期验证应该失败")
		}
		if len(result.ValidationDetails) == 0 {
			t.Error("预期应该有验证详情")
		}
	})

	t.Run("创建和验证有效APK", func(t *testing.T) {
		// 创建临时APK文件
		tempDir := t.TempDir()
		apkPath := filepath.Join(tempDir, "test.apk")

		// 创建一个最小的有效APK（ZIP格式）
		err := createTestAPK(apkPath)
		if err != nil {
			t.Fatalf("创建测试APK失败: %v", err)
		}

		config := &ArtifactConfig{
			Platform:          PlatformAPK,
			SourcePath:        tempDir,
			MinFileSize:       100, // 100字节，很小的要求
			MaxFileSize:       DefaultAndroidMaxSize,
			ValidateIntegrity: true,
		}

		result, err := validator.ValidateAPK(apkPath, config)
		if err != nil {
			t.Errorf("验证APK失败: %v", err)
			if result != nil {
				for _, detail := range result.ValidationDetails {
					t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
				}
			}
		}
		if !result.Success {
			t.Error("预期验证应该成功")
			for _, detail := range result.ValidationDetails {
				t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
			}
		}
		if result.FileSize == 0 {
			t.Error("预期文件大小应该大于0")
		}
	})

	t.Run("APK文件大小超出限制", func(t *testing.T) {
		// 创建临时APK文件
		tempDir := t.TempDir()
		apkPath := filepath.Join(tempDir, "test.apk")

		err := createTestAPK(apkPath)
		if err != nil {
			t.Fatalf("创建测试APK失败: %v", err)
		}

		config := &ArtifactConfig{
			Platform:          PlatformAPK,
			SourcePath:        tempDir,
			MinFileSize:       1024 * 1024 * 100, // 100MB，很大的要求
			MaxFileSize:       DefaultAndroidMaxSize,
			ValidateIntegrity: false,
		}

		result, err := validator.ValidateAPK(apkPath, config)
		if err == nil {
			t.Error("预期应该返回错误")
		}
		if result.Success {
			t.Error("预期验证应该失败（文件太小）")
		}
	})
}

func TestArtifactValidator_ValidateIOSApp(t *testing.T) {
	validator := NewArtifactValidator()

	t.Run("iOS App目录不存在", func(t *testing.T) {
		config := &ArtifactConfig{
			Platform:          PlatformIOS,
			SourcePath:        "/nonexistent",
			MinFileSize:       DefaultIOSAppMinSize,
			ValidateIntegrity: false,
		}

		result, err := validator.ValidateIOSApp("/nonexistent/Runner.app", config)
		if err == nil {
			t.Error("预期应该返回错误")
		}
		if result.Success {
			t.Error("预期验证应该失败")
		}
	})

	t.Run("创建和验证有效iOS App", func(t *testing.T) {
		// 创建临时iOS App目录
		tempDir := t.TempDir()
		appPath := filepath.Join(tempDir, "Runner.app")

		err := createTestIOSApp(appPath)
		if err != nil {
			t.Fatalf("创建测试iOS App失败: %v", err)
		}

		config := &ArtifactConfig{
			Platform:          PlatformIOS,
			SourcePath:        tempDir,
			MinFileSize:       100, // 100字节，很小的要求
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			ValidateIntegrity: false,
		}

		result, err := validator.ValidateIOSApp(appPath, config)
		if err != nil {
			t.Errorf("验证iOS App失败: %v", err)
			if result != nil {
				for _, detail := range result.ValidationDetails {
					t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
				}
			}
		}
		if !result.Success {
			t.Error("预期验证应该成功")
			for _, detail := range result.ValidationDetails {
				t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
			}
		}
		if result.FileSize == 0 {
			t.Error("预期目录大小应该大于0")
		}
	})
}

func TestArtifactValidator_ValidateIPA(t *testing.T) {
	validator := NewArtifactValidator()

	t.Run("IPA文件不存在", func(t *testing.T) {
		config := &ArtifactConfig{
			Platform:          PlatformIOS,
			SourcePath:        "/nonexistent",
			MinFileSize:       DefaultIOSIPAMinSize,
			MaxFileSize:       DefaultIOSIPAMaxSize,
			ValidateIntegrity: false,
		}

		result, err := validator.ValidateIPA("/nonexistent/test.ipa", config)
		if err == nil {
			t.Error("预期应该返回错误")
		}
		if result.Success {
			t.Error("预期验证应该失败")
		}
	})

	t.Run("创建和验证有效IPA", func(t *testing.T) {
		// 创建临时IPA文件
		tempDir := t.TempDir()
		ipaPath := filepath.Join(tempDir, "test.ipa")

		err := createTestIPA(ipaPath)
		if err != nil {
			t.Fatalf("创建测试IPA失败: %v", err)
		}

		config := &ArtifactConfig{
			Platform:          PlatformIOS,
			SourcePath:        tempDir,
			MinFileSize:       100, // 100字节，很小的要求
			MaxFileSize:       DefaultIOSIPAMaxSize,
			ValidateIntegrity: true,
		}

		result, err := validator.ValidateIPA(ipaPath, config)
		if err != nil {
			t.Errorf("验证IPA失败: %v", err)
			if result != nil {
				for _, detail := range result.ValidationDetails {
					t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
				}
			}
		}
		if !result.Success {
			t.Error("预期验证应该成功")
			for _, detail := range result.ValidationDetails {
				t.Logf("  %s: %s - %s", detail.Check, detail.Status, detail.Message)
			}
		}
		if result.FileSize == 0 {
			t.Error("预期文件大小应该大于0")
		}
	})
}

func TestArtifactValidator_GetExpectedPaths(t *testing.T) {
	validator := NewArtifactValidator()

	t.Run("Android平台", func(t *testing.T) {
		paths, err := validator.GetExpectedPaths(PlatformAPK, "/test/project", nil)
		if err != nil {
			t.Errorf("获取Android预期路径失败: %v", err)
		}
		if len(paths) != 1 {
			t.Errorf("预期1个路径，实际得到%d个", len(paths))
		}
		expectedPath := filepath.Join("/test/project", "build", "app", "outputs", "flutter-apk", "app-release.apk")
		if paths[0] != expectedPath {
			t.Errorf("预期路径%s，实际得到%s", expectedPath, paths[0])
		}
	})

	t.Run("iOS平台无证书", func(t *testing.T) {
		paths, err := validator.GetExpectedPaths(PlatformIOS, "/test/project", nil)
		if err != nil {
			t.Errorf("获取iOS预期路径失败: %v", err)
		}
		if len(paths) != 1 {
			t.Errorf("预期1个路径，实际得到%d个", len(paths))
		}
		expectedPath := filepath.Join("/test/project", "build", "ios", "iphoneos", "Runner.app")
		if paths[0] != expectedPath {
			t.Errorf("预期路径%s，实际得到%s", expectedPath, paths[0])
		}
	})

	t.Run("iOS平台有证书", func(t *testing.T) {
		iosConfig := &types.IOSConfig{
			TeamID:   "ABCD123456",
			BundleID: "com.example.app",
		}
		paths, err := validator.GetExpectedPaths(PlatformIOS, "/test/project", iosConfig)
		if err != nil {
			t.Errorf("获取iOS预期路径失败: %v", err)
		}
		if len(paths) != 1 {
			t.Errorf("预期1个路径，实际得到%d个", len(paths))
		}
		expectedPath := filepath.Join("/test/project", "build", "ios", "ipa")
		if paths[0] != expectedPath {
			t.Errorf("预期路径%s，实际得到%s", expectedPath, paths[0])
		}
	})
}

func TestValidationConfig(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		config := GetDefaultValidationConfig()
		if !config.EnableValidation {
			t.Error("默认应该启用验证")
		}
		if !config.EnableIntegrityCheck {
			t.Error("默认应该启用完整性检查")
		}
		if config.CustomMinSize != 0 {
			t.Error("默认最小大小应该为0")
		}
		if config.CustomMaxSize != 0 {
			t.Error("默认最大大小应该为0")
		}
	})

	t.Run("禁用验证", func(t *testing.T) {
		validator := NewArtifactValidator()
		config := &ArtifactConfig{
			Platform:    PlatformAPK,
			SourcePath:  "/test",
			ValidationConfig: &ArtifactValidationConfig{
				EnableValidation: false,
			},
		}

		result, err := validator.ValidateArtifact(config)
		if err != nil {
			t.Errorf("验证失败: %v", err)
		}
		if !result.Success {
			t.Error("禁用验证时应该返回成功")
		}
	})
}

// 辅助函数：创建测试APK文件
func createTestAPK(apkPath string) error {
	file, err := os.Create(apkPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建ZIP writer
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 添加AndroidManifest.xml
	manifest, err := zipWriter.Create("AndroidManifest.xml")
	if err != nil {
		return err
	}
	manifest.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.test">
    <application android:label="Test App">
        <activity android:name=".MainActivity">
            <intent-filter>
                <action android:name="android.intent.action.MAIN"/>
                <category android:name="android.intent.category.LAUNCHER"/>
            </intent-filter>
        </activity>
    </application>
</manifest>`))

	// 添加classes.dex
	classes, err := zipWriter.Create("classes.dex")
	if err != nil {
		return err
	}
	classes.Write([]byte("dex\n"))

	// 添加resources.arsc
	resources, err := zipWriter.Create("resources.arsc")
	if err != nil {
		return err
	}
	resources.Write([]byte("resources"))

	// 添加更多文件以满足最小数量要求
	for i := 0; i < 15; i++ {
		filename := fmt.Sprintf("assets/file%d.txt", i)
		file, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}
		file.Write([]byte("test file content"))
	}

	return nil
}

// 辅助函数：创建测试iOS App目录
func createTestIOSApp(appPath string) error {
	// 创建目录
	err := os.MkdirAll(appPath, 0755)
	if err != nil {
		return err
	}

	// 创建Info.plist
	plistPath := filepath.Join(appPath, "Info.plist")
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>Test App</string>
	<key>CFBundleIdentifier</key>
	<string>com.example.test</string>
</dict>
</plist>`
	err = os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		return err
	}

	// 创建Runner可执行文件
	runnerPath := filepath.Join(appPath, "Runner")
	err = os.WriteFile(runnerPath, []byte("fake executable"), 0755)
	if err != nil {
		return err
	}

	// 创建Frameworks目录（可选）
	frameworksPath := filepath.Join(appPath, "Frameworks")
	err = os.MkdirAll(frameworksPath, 0755)
	if err != nil {
		return err
	}

	return nil
}

// 辅助函数：创建测试IPA文件
func createTestIPA(ipaPath string) error {
	file, err := os.Create(ipaPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建ZIP writer
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 创建Payload目录和App包
	appDir := "Payload/TestApp.app/"

	// 添加Info.plist
	plist, err := zipWriter.Create(appDir + "Info.plist")
	if err != nil {
		return err
	}
	plist.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>Test App</string>
	<key>CFBundleIdentifier</key>
	<string>com.example.test</string>
</dict>
</plist>`))

	// 添加可执行文件
	executable, err := zipWriter.Create(appDir + "TestApp")
	if err != nil {
		return err
	}
	executable.Write([]byte("fake executable"))

	// 添加一些其他文件
	for i := 0; i < 15; i++ {
		filename := appDir + "file" + string(rune('0'+i)) + ".txt"
		file, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}
		file.Write([]byte("test file content"))
	}

	return nil
}