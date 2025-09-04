package certificates

import (
	"fmt"
	"regexp"
	"strings"
)

// IdentifierGenerator 标识符生成器
type IdentifierGenerator struct {
	teamID   string
	bundleID string
}

// NewIdentifierGenerator 创建标识符生成器
func NewIdentifierGenerator(teamID, bundleID string) *IdentifierGenerator {
	return &IdentifierGenerator{
		teamID:   teamID,
		bundleID: bundleID,
	}
}

// Generate 生成唯一标识符
func (ig *IdentifierGenerator) Generate() string {
	// 1. 规范化TeamID和BundleID
	normalizedTeamID := ig.normalizeIdentifier(ig.teamID)
	normalizedBundleID := ig.normalizeIdentifier(ig.bundleID)
	
	// 2. 组合生成唯一标识符
	identifier := fmt.Sprintf("%s_%s", 
		normalizedTeamID, 
		normalizedBundleID)
	
	// 3. 确保长度限制（Keychain名称限制）
	if len(identifier) > 50 {
		identifier = identifier[:50]
	}
	
	return identifier
}

// normalizeIdentifier 规范化标识符
func (ig *IdentifierGenerator) normalizeIdentifier(input string) string {
	// 移除特殊字符，替换为下划线
	normalized := strings.ToLower(input)
	normalized = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "_")
	// 移除连续的下划线
	normalized = regexp.MustCompile(`_+`).ReplaceAllString(normalized, "_")
	// 移除首尾下划线
	normalized = strings.Trim(normalized, "_")
	return normalized
}