package chatbot

import (
	"strings"
	"testing"
)

func TestKnowledgeBaseIsNotEmpty(t *testing.T) {
	if len(KnowledgeBase) == 0 {
		t.Fatal("KnowledgeBase constant should not be empty")
	}
}

func TestKnowledgeBaseContainsRequiredSections(t *testing.T) {
	requiredSections := []string{
		"## Navigation",
		"## Features",
		"### Payment Pipeline",
		"### Payment Statuses",
		"## Risk Levels",
		"## Document Upload",
		"## REA",
		"## Contracts",
		"### CLIN Types",
		"### CLIN Statuses",
		"## Roles and Permissions",
		"## Financial Terms",
		"## Routing Thresholds",
	}

	for _, section := range requiredSections {
		if !strings.Contains(KnowledgeBase, section) {
			t.Errorf("KnowledgeBase missing required section: %s", section)
		}
	}
}

func TestBuildSystemPromptIncludesUserContext(t *testing.T) {
	prompt := BuildSystemPrompt("CONTRACTING_OFFICER", "ORG-001", "/contracts")

	if !strings.Contains(prompt, "CONTRACTING_OFFICER") {
		t.Error("System prompt should contain the user role")
	}
	if !strings.Contains(prompt, "ORG-001") {
		t.Error("System prompt should contain the organization ID")
	}
	if !strings.Contains(prompt, "/contracts") {
		t.Error("System prompt should contain the current page")
	}
}

func TestBuildSystemPromptIncludesKnowledgeBase(t *testing.T) {
	prompt := BuildSystemPrompt("CONTRACTOR", "ORG-002", "/dashboard")

	if !strings.Contains(prompt, "Federal Payment Processing Platform - Knowledge Base") {
		t.Error("System prompt should contain the knowledge base")
	}
}

func TestBuildSystemPromptRoleInstructions(t *testing.T) {
	tests := []struct {
		role     string
		expected string
	}{
		{"CONTRACTING_OFFICER", "full access to all contracts"},
		{"COR", "view-only access"},
		{"PROCURING_CONTRACTING_OFFICER", "organization's contracts"},
		{"CONTRACTOR", "limited to their organization's contracts"},
		{"PROGRAM_MANAGER", "Program Manager"},
		{"UNKNOWN_ROLE", "role is not recognized"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			prompt := BuildSystemPrompt(tt.role, "ORG-001", "/")
			if !strings.Contains(prompt, tt.expected) {
				t.Errorf("System prompt for role %s should contain %q", tt.role, tt.expected)
			}
		})
	}
}
