package middleware

import "testing"

// ⚠️ CI 红灯证据用例（RED-LIGHT EVIDENCE）——本文件仅用于采集
// “单元测试失败 → 镜像构建被阻断”的流水线证据，取证提交后立即回滚，
// 严禁合入 dev/main。见 PR #70 与 docs/testing/reports/ci-red-green-evidence-2026-08-26.md
func TestRedLightEvidenceDeliberateFailure(t *testing.T) {
	t.Errorf("故意失败：验证 CI 在单元测试失败时停止后续 api-test 与 docker-build（RED-LIGHT EVIDENCE）")
}
