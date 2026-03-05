package novel

import (
	"encoding/json"
	"fmt"
	"novel-generator/pkg/ollama"
	"strings"
)

type Chapter struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Content string `json:"content,omitempty"`
}

type Novel struct {
	Title    string    `json:"title"`
	Intro    string    `json:"intro"`
	Outline  string    `json:"outline"`
	Chapters []Chapter `json:"chapters"`
}

type Generator struct {
	Client *ollama.Client
}

func NewGenerator(model string) *Generator {
	if model == "" {
		model = "qwen2.5:7b-instruct"
	}
	return &Generator{
		Client: ollama.NewClient(model),
	}
}

func (g *Generator) GenerateOutline(title, intro string) (string, error) {
	prompt := fmt.Sprintf(`你是一个专业的小说家。请根据以下信息，为一个长篇小说创作一个详细的大纲。
小说标题：%s
小说简介：%s

要求：
1. 大纲需要包含主线剧情的发展脉络。
2. 设定主要人物的性格和背景。
3. 规划好故事的起承转合。
4. 这是一个长篇小说的大纲，预计写十万字以上。
5. 请直接输出大纲内容，不要包含其他开场白。`, title, intro)

	return g.Client.Generate(prompt)
}

func (g *Generator) GenerateChapters(outline string) ([]Chapter, error) {
	prompt := fmt.Sprintf(`你是一个专业的小说家。请根据以下小说大纲，规划出详细的章节目录。
小说大纲：
%s

要求：
1. 请输出至少50个章节。
2. 每个章节需要包含章节标题和简短的剧情梗概（100字以内）。
3. 剧情要连贯，节奏紧凑。
4. 输出格式必须严格遵守JSON格式，是一个包含对象的数组，每个对象有 "id" (数字), "title" (字符串), "summary" (字符串) 字段。
例如：[{"id": 1, "title": "第一章...", "summary": "..."}]
5. 不要输出JSON以外的任何内容。`, outline)

	resp, err := g.Client.Generate(prompt)
	if err != nil {
		return nil, err
	}

	// Clean up markdown code blocks if present
	resp = strings.TrimPrefix(resp, "```json")
	resp = strings.TrimPrefix(resp, "```")
	resp = strings.TrimSuffix(resp, "```")
	resp = strings.TrimSpace(resp)

	var chapters []Chapter
	if err := json.Unmarshal([]byte(resp), &chapters); err != nil {
		return nil, fmt.Errorf("failed to parse chapters JSON: %v. Response was: %s", err, resp)
	}
	return chapters, nil
}

func (g *Generator) GenerateChapterContent(chapter Chapter, prevSummary string) (string, error) {
	prompt := fmt.Sprintf(`你是一个网文大神。请根据以下信息扩写这一章的正文。
章节标题：%s
本章概要：%s
前情提要：%s

要求：
1. 扩写内容要丰富细节，包括环境描写、人物心理、对话描写等。
2. 保证字数在2000字左右。
3. 语言通顺，风格符合网文阅读习惯。
4. 不要重复输出章节标题。
5. 直接输出正文内容。`, chapter.Title, chapter.Summary, prevSummary)

	return g.Client.Generate(prompt)
}
