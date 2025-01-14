package handlers

import (
	"bytes"
	"html/template"
	"io/ioutil"

	"github.com/gofiber/fiber/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// HTML模板
const docTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>API文档</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.2.0/github-markdown.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"></script>
    <style>
        .markdown-body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            margin: 0 auto;
            padding: 45px;
        }
        @media (max-width: 767px) {
            .markdown-body {
                padding: 15px;
            }
        }
        pre code {
            background-color: transparent !important;
        }
    </style>
</head>
<body>
    <article class="markdown-body">
        {{.Content}}
    </article>
    <script>
        document.addEventListener('DOMContentLoaded', (event) => {
            document.querySelectorAll('pre code').forEach((el) => {
                hljs.highlightElement(el);
            });
        });
    </script>
</body>
</html>
`

func HandleApiDoc(c *fiber.Ctx) error {
	// 从根目录读取Markdown文件内容
	content, err := ioutil.ReadFile("./api.md")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error reading Markdown file")
	}

	// 配置 goldmark
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	// 转换 Markdown 为 HTML
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error converting Markdown to HTML")
	}

	// 创建HTML模板
	tmpl, err := template.New("doc").Parse(docTemplate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error parsing template")
	}

	// 准备模板数据
	data := struct {
		Content template.HTML
	}{
		Content: template.HTML(buf.String()),
	}

	// 渲染最终的HTML
	var finalHTML bytes.Buffer
	if err := tmpl.Execute(&finalHTML, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error executing template")
	}

	// 设置响应头
	c.Set("Content-Type", "text/html; charset=utf-8")

	return c.Send(finalHTML.Bytes())
}
