package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestConsolidatedToolsExposeSevenStrictTools(t *testing.T) {
	tools := consolidatedTools()
	wantNames := []string{
		"configure",
		"zentao_user",
		"zentao_products",
		"zentao_bugs",
		"zentao_builds",
		"zentao_stories",
		"zentao_testcases",
	}
	if len(tools) != len(wantNames) {
		t.Fatalf("consolidatedTools() count = %d, want %d", len(tools), len(wantNames))
	}
	for index, want := range wantNames {
		if tools[index].tool.Name != want {
			t.Errorf("tool[%d].Name = %q, want %q", index, tools[index].tool.Name, want)
		}
		if len(tools[index].tool.RawInputSchema) == 0 {
			t.Errorf("tool %q does not use a strict raw schema", want)
		}
	}
}

func TestActionSchemasRejectExtraPropertiesPerAction(t *testing.T) {
	item := findRegisteredTool(t, "zentao_bugs")
	var schema struct {
		Required             []string    `json:"required"`
		AdditionalProperties interface{} `json:"additionalProperties"`
		OneOf                []struct {
			Title                string      `json:"title"`
			Required             []string    `json:"required"`
			AdditionalProperties interface{} `json:"additionalProperties"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(item.tool.RawInputSchema, &schema); err != nil {
		t.Fatalf("parse zentao_bugs schema: %v", err)
	}
	if len(schema.OneOf) != 6 {
		t.Fatalf("zentao_bugs oneOf branches = %d, want 6", len(schema.OneOf))
	}
	if len(schema.Required) != 1 || schema.Required[0] != "action" || schema.AdditionalProperties != false {
		t.Fatalf("top-level strictness = required %#v, additionalProperties %#v", schema.Required, schema.AdditionalProperties)
	}
	for _, branch := range schema.OneOf {
		if branch.AdditionalProperties != false {
			t.Errorf("branch %q additionalProperties = %#v, want false", branch.Title, branch.AdditionalProperties)
		}
		if len(branch.Required) == 0 || branch.Required[0] != "action" {
			t.Errorf("branch %q required = %#v, want action first", branch.Title, branch.Required)
		}
	}
}

func TestActionHandlerRejectsUnknownActionAndParameters(t *testing.T) {
	item := findRegisteredTool(t, "zentao_bugs")

	result := callTool(t, item.handler, map[string]interface{}{"action": "remove", "id": float64(1)})
	assertToolErrorContains(t, result, "action 不支持")

	result = callTool(t, item.handler, map[string]interface{}{
		"action": "get",
		"id":     float64(1),
		"title":  "不属于 get action 的参数",
	})
	assertToolErrorContains(t, result, "不支持的参数: title")
}

func TestActionHandlerRejectsWrongTypesAndRanges(t *testing.T) {
	item := findRegisteredTool(t, "zentao_bugs")
	result := callTool(t, item.handler, map[string]interface{}{
		"action":   "create",
		"title":    "测试 Bug",
		"severity": float64(5),
		"pri":      float64(1),
		"type":     "codeerror",
	})
	assertToolErrorContains(t, result, "severity 参数错误")

	result = callTool(t, item.handler, map[string]interface{}{
		"action": "get",
		"id":     "1",
	})
	assertToolErrorContains(t, result, "id 参数错误")
}

func TestActionHandlerNormalizesIntegerIDForLegacyHandler(t *testing.T) {
	var received map[string]interface{}
	actions := []actionDefinition{{
		name:        "get",
		description: "测试",
		params:      map[string]parameterSpec{"id": idStringParam("对象 ID")},
		required:    []string{"id"},
		handler: func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			received = request.Params.Arguments
			return mcp.NewToolResultText("ok"), nil
		},
	}}

	result := callTool(t, newActionToolHandler("test_tool", actions), map[string]interface{}{
		"action": "get",
		"id":     float64(42),
	})
	if result.IsError {
		t.Fatalf("handler returned error: %s", toolResultText(t, result))
	}
	if received["id"] != "42" {
		t.Fatalf("normalized id = %#v, want string 42", received["id"])
	}
	if _, exists := received["action"]; exists {
		t.Fatalf("legacy handler unexpectedly received action: %#v", received)
	}
}

func TestTestCaseStepsRequireExactShape(t *testing.T) {
	item := findRegisteredTool(t, "zentao_testcases")
	result := callTool(t, item.handler, map[string]interface{}{
		"action": "create",
		"title":  "登录用例",
		"type":   "feature",
		"steps": []interface{}{
			map[string]interface{}{"desc": "输入账号", "expect": "登录成功", "extra": "不允许"},
		},
	})
	assertToolErrorContains(t, result, "只能包含 desc 和 expect")
}

func TestConfigureRejectsUnknownParameterBeforeChangingConfig(t *testing.T) {
	item := findRegisteredTool(t, "configure")
	result := callTool(t, item.handler, map[string]interface{}{
		"base_url": "https://zentao.example.test/api.php/v1",
		"account":  "tester",
		"password": "secret",
		"token":    "模型不应传入的字段",
	})
	assertToolErrorContains(t, result, "不支持的参数: token")
}

func findRegisteredTool(t *testing.T, name string) registeredTool {
	t.Helper()
	for _, item := range consolidatedTools() {
		if item.tool.Name == name {
			return item
		}
	}
	t.Fatalf("tool %q not found", name)
	return registeredTool{}
}

func callTool(t *testing.T, handler server.ToolHandlerFunc, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Arguments = args
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("handler returned Go error: %v", err)
	}
	return result
}

func assertToolErrorContains(t *testing.T, result *mcp.CallToolResult, substring string) {
	t.Helper()
	if result == nil || !result.IsError {
		t.Fatalf("result = %#v, want tool error containing %q", result, substring)
	}
	text := toolResultText(t, result)
	if !strings.Contains(text, substring) {
		t.Fatalf("tool error = %q, want substring %q", text, substring)
	}
}

func toolResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) != 1 {
		t.Fatalf("result content = %#v, want one text item", result)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("result content type = %T, want mcp.TextContent", result.Content[0])
	}
	return content.Text
}
