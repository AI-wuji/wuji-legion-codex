package core

import "strings"

var officerContracts = map[string]OfficerContract{
	"internal-quality-check": {
		Role: "internal-quality-check", TaskTypes: []string{"medium"}, Stages: []string{"deterministic-general-staff-review"},
		RiskSignals: []string{"test", "review"}, ArtifactTypes: []string{"code", "artifact"}, EvidenceGaps: []string{"acceptance-evidence"}, RequiresUserConfirmation: false,
	},
	"composite-moe-officer": {
		Role: "composite-moe-officer", TaskTypes: []string{"large", "high-risk"}, Stages: []string{"independent-review", "quality-inspection"},
		RiskSignals: []string{"architecture", "migration", "governance"}, ArtifactTypes: []string{"code", "artifact", "receipt"}, EvidenceGaps: []string{"acceptance-evidence"}, RequiresUserConfirmation: false,
	},
}

// SelectOfficerRecommendations keeps routine checks inside General Staff.
// It returns at most one composite officer recommendation and never starts one.
func SelectOfficerRecommendations(query string) []OfficerRecommendation {
	query = strings.ToLower(strings.TrimSpace(query))
	if containsAny(query, "test", "review", "verify", "regression", "验收", "测试", "审查", "回归", "质检") && !containsAny(query, "architecture", "migration", "routing", "model policy", "dependency upgrade", "permission", "security policy", "架构", "迁移", "路由", "模型策略", "依赖升级", "权限", "安全策略") {
		contract := officerContracts["internal-quality-check"]
		return []OfficerRecommendation{{Role: contract.Role, Decision: "internal-quality-check", Reason: "medium task receives internal quality checking; staff tracks the evidence but does not accept completion", Contract: contract}}
	}
	if !containsAny(query, "architecture", "migration", "routing", "model policy", "dependency upgrade", "permission", "security policy", "架构", "迁移", "路由", "模型策略", "依赖升级", "权限", "安全策略") {
		return nil
	}
	contract := officerContracts["composite-moe-officer"]
	decision := "independent-composite-quality-inspection"
	reason := "large or high-risk task permits one independent composite MoE quality-inspection call"
	if containsAny(query, "authorization", "source", "compliance", "release", "publish", "payment", "delete", "sensitive data", "dispute evidence", "授权", "来源", "合规", "发布", "付款", "删除", "敏感数据", "争议证据") {
		contract.Stages = append(contract.Stages, "audit")
		decision = "independent-composite-quality-inspection-with-audit"
		reason = "governance risk adds an audit section to the single independent composite MoE call"
	}
	return []OfficerRecommendation{{Role: contract.Role, Decision: decision, Reason: reason, Contract: contract}}
}
