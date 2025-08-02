package obsidian

import (
	"fmt"
	"time"
)

// Template represents a note generation template
type Template struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Content     string             `json:"content"`
	Variables   []TemplateVariable `json:"variables"`
}

// TemplateVariable represents a variable used in templates
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, date, number, boolean
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
}

// getDefaultTemplates returns the default set of templates
func getDefaultTemplates() map[string]*Template {
	return map[string]*Template{
		"daily":   getDailyTemplate(),
		"summary": getSummaryTemplate(),
		"report":  getReportTemplate(),
	}
}

// getDailyTemplate returns the daily note template
func getDailyTemplate() *Template {
	return &Template{
		Name:        "daily",
		Description: "日記スタイルのテンプレート - 日付ベースの記憶を時系列で整理",
		Content: `---
date: {{date}}
tags: [daily, mory-generated]
category: {{category}}
---

# {{title}}

*Generated on {{date}} from Mory memories*

## 今日の記憶

{{#memories}}
{{/memories}}

## 関連する記憶

{{#related_memories}}
{{/related_memories}}

## 統計情報

- 記憶数: {{stats.total_memories}}
- 関連記憶数: {{stats.related_memories}}

---
*このノートはMoryによって自動生成されました*`,
		Variables: []TemplateVariable{
			{
				Name:        "title",
				Type:        "string",
				Description: "ノートのタイトル",
				Required:    true,
			},
			{
				Name:        "category",
				Type:        "string",
				Description: "フィルタするカテゴリ",
				Required:    false,
			},
			{
				Name:        "date",
				Type:        "date",
				Description: "生成日付",
				Required:    false,
				Default:     time.Now().Format("2006-01-02"),
			},
		},
	}
}

// getSummaryTemplate returns the summary note template
func getSummaryTemplate() *Template {
	return &Template{
		Name:        "summary",
		Description: "サマリースタイルのテンプレート - 特定カテゴリの記憶をまとめる",
		Content: `---
title: {{title}}
date: {{date}}
tags: [summary, mory-generated]
category: {{category}}
---

# {{title}}

*Generated on {{date}} from Mory memories*

## 概要

このサマリーは{{category}}カテゴリの記憶をまとめたものです。

## 記憶一覧

{{#memories}}
{{/memories}}

## 関連情報

{{#related_memories}}
{{/related_memories}}

## 統計

| 項目 | 数 |
|------|-----|
| 総記憶数 | {{stats.total_memories}} |
| 関連記憶数 | {{stats.related_memories}} |

## まとめ

- 主要なトピック: {{category}}
- 記録された記憶の数: {{stats.total_memories}}
- 生成日時: {{date}}

---
*このサマリーはMoryによって自動生成されました*`,
		Variables: []TemplateVariable{
			{
				Name:        "title",
				Type:        "string",
				Description: "サマリーのタイトル",
				Required:    true,
			},
			{
				Name:        "category",
				Type:        "string",
				Description: "サマリー対象のカテゴリ",
				Required:    true,
			},
			{
				Name:        "date",
				Type:        "date",
				Description: "生成日付",
				Required:    false,
				Default:     time.Now().Format("2006-01-02"),
			},
		},
	}
}

// getReportTemplate returns the report note template
func getReportTemplate() *Template {
	return &Template{
		Name:        "report",
		Description: "レポートスタイルのテンプレート - プロジェクト/テーマ別の記憶を整理",
		Content: `---
title: {{title}}
date: {{date}}
tags: [report, mory-generated]
category: {{category}}
type: report
---

# {{title}}

*Report generated on {{date}} from Mory memories*

## エグゼクティブサマリー

このレポートは{{category}}に関連する記憶とデータをまとめたものです。

## 主要な発見事項

{{#memories}}
{{/memories}}

## 関連データ

{{#related_memories}}
{{/related_memories}}

## 詳細分析

### データ概要
- 分析対象記憶数: {{stats.total_memories}}
- 関連記憶数: {{stats.related_memories}}
- 生成日: {{date}}

### カテゴリ分析
- 主要カテゴリ: {{category}}
- データ収集期間: 〜{{date}}

## 推奨事項

1. 継続的なデータ収集の実施
2. 関連記憶の定期的な見直し
3. カテゴリ分類の最適化

## 次のステップ

- [ ] データの追加収集
- [ ] 関連記憶の詳細分析
- [ ] カテゴリの見直しと整理

---
*このレポートはMoryによって自動生成されました*
*レポート作成日時: {{date}}*`,
		Variables: []TemplateVariable{
			{
				Name:        "title",
				Type:        "string",
				Description: "レポートのタイトル",
				Required:    true,
			},
			{
				Name:        "category",
				Type:        "string",
				Description: "レポート対象のカテゴリ",
				Required:    true,
			},
			{
				Name:        "date",
				Type:        "date",
				Description: "生成日付",
				Required:    false,
				Default:     time.Now().Format("2006-01-02"),
			},
		},
	}
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*Template, bool) {
	templates := getDefaultTemplates()
	template, exists := templates[name]
	return template, exists
}

// ListTemplates returns all available templates
func ListTemplates() map[string]*Template {
	return getDefaultTemplates()
}

// ValidateTemplate validates a template structure
func ValidateTemplate(template *Template) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Content == "" {
		return fmt.Errorf("template content is required")
	}
	return nil
}

// ValidateTemplateData validates template data against template variables
func ValidateTemplateData(template *Template, data map[string]string) error {
	for _, variable := range template.Variables {
		if variable.Required {
			if _, exists := data[variable.Name]; !exists {
				return fmt.Errorf("required variable '%s' is missing", variable.Name)
			}
		}
	}
	return nil
}
