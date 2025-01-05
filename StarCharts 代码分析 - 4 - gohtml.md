## StarCharts 代码分析 - 4 - go html 

### 4.1 go html template 分析

看一个简单的例子，演示如何定义和调用模板：

```go
package main

import (
    "html/template"
    "os"
)

// 定义一个结构体来作为模板的数据
type PageData struct {
    Title   string
    Content string
}

func main() {
    // 定义一个模板，包含主模板和子模板
    const tmpl = `
        {{define "main"}} 
            <html>
                <head><title>{{.Title}}</title></head>
                <body>
                    <h1>{{.Content}}</h1>
                </body>
            </html>
        {{end}}

        {{template "main" .}} 
    `

    // 模板数据
    data := PageData{
        Title:   "Hello, Go Templates",
        Content: "This is a Go HTML Template example.",
    }

    // 解析并执行模板
    t, err := template.New("webpage").Parse(tmpl)
    if err != nil {
        panic(err)
    }

    // 执行模板，将数据传递给模板并输出到 os.Stdout
    err = t.Execute(os.Stdout, data)
    if err != nil {
        panic(err)
    }
}

```

**这段代码可以分为以下几个功能：**

1. **定义子模板**：

```go
{{define "main"}} 
    <html>
        <head><title>{{.Title}}</title></head>
        <body>
            <h1>{{.Content}}</h1>
        </body>
    </html>
{{end}}
```

这里我们用 `{{define "main"}}` 定义了一个子模板 `"main"`，它包含了 HTML 的结构，并使用 `{{.Title}}` 和 `{{.Content}}` 作为模板中的占位符来渲染数据。

2. **调用子模板**：

```go
{{template "main" .}} 
```

这行代码通过 `{{template "main" .}}` 调用名为 `"main"` 的子模板，并将 `.`（即 `PageData` 结构体的数据）传递给子模板。

3. **数据传递**：

在这个例子中，`PageData` 结构体包含了 `Title` 和 `Content` 字段，这些字段的值将被传递给 `"main"` 模板，分别渲染到 `<title>` 标签和 `<h1>` 标签中。

4. **输出结果**：

当模板渲染时，`{{template "main" .}}` 会被替换为子模板 `"main"` 的内容，并且 `.Title` 和 `.Content` 会被替换为对应的数据。最终的输出将是：

```html
<html>
    <head><title>Hello, Go Templates</title></head>
    <body>
        <h1>This is a Go HTML Template example.</h1>
    </body>
</html>
```

**总结：**

- `{{template "main" .}}` 是 Go 模板中的一个嵌套模板调用语法，它调用了名为 `"main"` 的子模板，并将当前的上下文（`.`）传递给该模板进行渲染。
- 通过这种方式，可以在模板中组织结构化的内容，将主要的模板内容拆分成多个可重用的子模板，提高代码的可维护性和可复用性。
- 这类模板嵌套在 web 开发中尤其有用，例如可以为不同的页面渲染不同的内容，但共用一个布局模板。



### 4.2 base.gohtml 分析

```go
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge"/>
    <meta name="theme-color" content="#000000"/>
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="/static/favicon.svg"/>
    <link rel="apple-touch-icon-precomposed" sizes="152x152" href="/static/favicon.svg"/>
    <link rel="icon" type="image/png" href="/static/favicon.png" sizes="32x32"/>
    <link rel="icon" type="image/png" href="/static/favicon.png" sizes="16x16"/>
    <link rel="icon" type="image/svg" href="/static/favicon.svg" sizes="32x32"/>
    <link rel="icon" type="image/svg" href="/static/favicon.svg" sizes="16x16"/>
    <title>{{block "title" .}} {{end}}</title>
    <meta name="description" content="StarCharts"/>
    <meta name="author" content="https://github/caarlos0"/>
    <link rel="stylesheet" href="/static/styles.css?v={{ .Version }}">
    {{block "head" .}} {{end}}
</head>
<body>
{{template "main" .}}
<script defer data-domain="starchart.cc" src="https://plausible.io/js/plausible.js"></script>
</body>
</html>

{{ define "logo" }}
    <a class="title" href="/">
        <img src="/static/stars.svg" alt="Stars">
        <span class="title">starcharts</span>
        <span class="subtitle">Plot your repository stars over time.</span>
    </a>
{{ end }}
```

在 Go 模板中，`{{template "main" .}}` 是一种调用其他模板的语法。让我们详细解释这一部分代码的意义和背景：

1. **模板系统**

Go 的 `html/template` 和 `text/template` 包提供了模板功能，允许我们动态地生成 HTML、文本或其他格式的输出。模板文件包含了一些占位符和控制结构，Go 语言的模板引擎会根据传入的数据渲染这些占位符。

在模板中，你可以定义一些子模板，然后通过 `{{template "name" .}}` 语法来调用其他模板。这里的 `"name"` 是你定义的子模板的名字，`.` 是当前的上下文数据。

2. **语法解析**

```go
{{template "main" .}}
```

这段代码的含义是：调用名为 `"main"` 的子模板，并将当前模板的数据传递给它。具体细节如下：

- **`template`**: `template` 是一个模板函数，允许在当前模板中嵌套并调用其他模板。你可以通过这个函数引用和渲染其他模板。
- **`"main"`**: 这是你希望调用的模板的名称。模板的名称通常是通过 `{{define "name"}}...{{end}}` 语法定义的。你可以理解为 `"main"` 是一个已定义的模板块的名字。模板名称不一定是文件名，它可以是任何字符串。
- **`.`**: 这是 Go 模板中的 **dot**（点）符号，它表示当前的上下文或数据。模板渲染时会传入一些数据，并通过 `.` 引用这些数据。当前的上下文可以是一个简单的变量、结构体或复杂的数据结构。通过 `.`，你可以访问传入模板的数据。



### 4.3 index.gohtml 分析

```go
{{define "title"}}Star Charts{{end}}

{{define "main"}}
    <div class="container index">
        {{ with .Error }}
            <a class="title" href="/">
                <img src="/static/error.svg" alt="Stars">
                <span class="title">starcharts</span>
                <span class="subtitle">Plot your repository stars over time.</span>
            </a>
        {{ else }}
            {{template "logo" .}}
        {{ end }}
        <hr />
        <div class="main">
            {{ with .Error }}
                <p class="error">{{ . }}</p>
            {{ end }}
            <form method="POST" action="/">
                <label for="repository">Repository:</label><br>
                <input type="text" id="repository" name="repository" value="caarlos0/starcharts"
                       placeholder="caarlos0/starcharts" autofocus="autofocus"><br>
                <button type="submit" class="full-width">Submit</button>
            </form>
        </div>
    </div>
    <script type="text/javascript">
        const repository = document.querySelector('input#repository');
        if (!repository) {
            throw new Error('repo input not found');
        }

        const lastRepoKey = 'last-repo';
        if (localStorage && localStorage.getItem(lastRepoKey)) {
            repository.value = localStorage.getItem(lastRepoKey);
        }

        repository.select();
        document.querySelector('form').addEventListener('submit', () => {
            localStorage && localStorage.setItem(lastRepoKey, repository.value);
        });
    </script>
{{end}}
```

这段代码是一个用 Go 模板引擎（Go templates）编写的 Web 页面，主要功能是允许用户输入一个 GitHub 仓库的名称，提交后生成该仓库的星标数（Stars）随时间变化的图表。代码中包含 HTML、CSS、Go 模板语法和 JavaScript 逻辑。

其中：

```go
<form method="POST" action="/">
    <label for="repository">Repository:</label><br>
    <input type="text" id="repository" name="repository" value="caarlos0/starcharts"
           placeholder="caarlos0/starcharts" autofocus="autofocus"><br>
    <button type="submit" class="full-width">Submit</button>
</form>
```

提供了一个表单，允许用户输入 GitHub 仓库的名称，默认值是 `caarlos0/starcharts`。

输入框有一个 `autofocus` 属性，页面加载后会自动将焦点放在该输入框上。

表单的提交方式是 `POST`，提交后会通过 `action="/"` 向服务器发送请求，进行仓库星标图表的生成。

**JavaScript 逻辑如下**

- **获取输入框**：

  ```javascript
  const repository = document.querySelector('input#repository');
  if (!repository) {
      throw new Error('repo input not found');
  }
  ```

  这段代码通过 `querySelector` 获取页面上的输入框元素（`input#repository`）。如果没有找到该元素，会抛出一个错误。

- **保存和加载最近使用的仓库名**：

  ```javascript
  const lastRepoKey = 'last-repo';
  if (localStorage && localStorage.getItem(lastRepoKey)) {
      repository.value = localStorage.getItem(lastRepoKey);
  }
  ```

  这段代码检查浏览器的 `localStorage` 是否有保存上次输入的仓库名称（`last-repo`）。如果存在，则将其填充到输入框中。`localStorage` 可以在页面刷新后依然保持数据。

- **表单提交时保存仓库名称**：

  ```javascript
  document.querySelector('form').addEventListener('submit', () => {
      localStorage && localStorage.setItem(lastRepoKey, repository.value);
  });
  ```

  这段代码为表单提交事件添加了监听器。在表单提交时，当前输入框的值（仓库名称）会保存到 `localStorage` 中，以便下次加载页面时可以记住上次输入的仓库名。



### 4.4 repository.gohtml 分析

```go
{{define "title"}}Star Charts | {{ .Details.FullName }} {{end}}

{{define "head"}}
    <link rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.1.0/styles/base16/dracula.min.css"
          integrity="sha512-oDvVpANXrKQ6R5B25VO6DooEQWA7jUXleyD6oUWHChC0fjv8wAANSX7lKXtp5D6HbZ7EUxd0wjMibtpCQ+aCDw=="
          crossorigin="anonymous" referrerpolicy="no-referrer"/>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/mdbassit/Coloris@latest/dist/coloris.min.css"/>
{{end}}

{{define "main"}}
    <div class="container">
        {{template "logo" .}}
        <hr/>
        {{ with .Details }}
            <div class="main">
                <p>
                    {{ if gt .StargazersCount 0 }}
                        <b>Awesome!</b>
                    {{ else }}
                        <b>Hang in there!</b>
                    {{ end }}
                    <a href="https://github.com/{{ .FullName }}">{{ .FullName }}</a>
                    was created
                    <time datetime="{{ .CreatedAt }}"></time>
                    and now has <b>{{ .StargazersCount }}</b> stars.
                </p>
            </div>

            {{ if gt .StargazersCount 0 }}
                <div class="chart-review">
                    <div class="chart-selection">
                        <div class="button-group">
                            <button data-variant="adaptive" class="active">Adaptive</button>
                            <button data-variant="light">Light</button>
                            <button data-variant="dark">Dark</button>
                            <button data-variant="custom">Custom</button>
                        </div>
                        <div class="customisation">
                            <label for="background">Background Color</label>
                            <input id="background" name="background" type="text" value="#FFFFFF" data-coloris>
                            <label for="axis">Axis Color</label>
                            <input id="axis" name="axis" type="text" value="#333333" data-coloris>
                            <label for="line">Line Color</label>
                            <input id="line" name="line" type="text" value="#6b63ff" data-coloris>
                        </div>
                    </div>
                    <div class="chart">
                        <img src="/{{ .FullName }}.svg?variant=adaptive"
                             id="chart"
                             data-src="/{{ .FullName }}.svg"
                             alt="Please try again in a few minutes. This might not work for very famous repository.">
                    </div>
                </div>
                <noscript id="code-template">## Stargazers over time
                    [![Stargazers over time]($URL)](https://starchart.cc/{{ .FullName }})</noscript>
                <p>
                    You can include the chart on your repository's
                    <code>README.md</code>
                    as follows:
                </p>
                <div class="code-block">
                    <pre class="markdown" id="code">
                        <code></code>
                    </pre>
                    <button class="copy-btn full-width" data-clipboard-target="#code">Copy</button>
                </div>
            {{ end }}

            <div class="footer">
                <a href="https://www.digitalocean.com/?refcode=7e8e9efb2f77&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge">
                    <img src="https://web-platforms.sfo2.cdn.digitaloceanspaces.com/WWW/Badge%201.svg"
                         alt="DigitalOcean Referral Badge" width="150px"/>
                </a>
            </div>
        {{end}}
    </div>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/timeago.js/4.0.2/timeago.min.js"
            integrity="sha512-SVDh1zH5N9ChofSlNAK43lcNS7lWze6DTVx1JCXH1Tmno+0/1jMpdbR8YDgDUfcUrPp1xyE53G42GFrcM0CMVg=="
            crossorigin="anonymous" referrerpolicy="no-referrer"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.1.0/highlight.min.js"
            integrity="sha512-z+/WWfyD5tccCukM4VvONpEtLmbAm5LDu7eKiyMQJ9m7OfPEDL7gENyDRL3Yfe8XAuGsS2fS4xSMnl6d30kqGQ=="
            crossorigin="anonymous" referrerpolicy="no-referrer"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.8/clipboard.min.js"
            integrity="sha512-sIqUEnRn31BgngPmHt2JenzleDDsXwYO+iyvQ46Mw6RL+udAUZj2n/u/PGY80NxRxynO7R9xIGx5LEzw4INWJQ=="
            crossorigin="anonymous" referrerpolicy="no-referrer"></script>
    <script src="https://cdn.jsdelivr.net/gh/mdbassit/Coloris@latest/dist/coloris.min.js"
            crossorigin="anonymous" referrerpolicy="no-referrer"></script>
    <script src="/static/scripts.js"></script>
{{end}}
```

这段代码用于生成一个网页，该网页展示了 GitHub 仓库的星标数（Stars）随时间变化的图表，并允许用户自定义该图表的样式。具体来说，页面展示了以下内容和功能：

#### 4.4.1仓库信息展示

如果 .Details存在，它包含有关 GitHub 仓库的信息，包括：

- 仓库的全名（`FullName`）
- 仓库的创建时间（`CreatedAt`）
- 仓库的星标数（`StargazersCount`）

```go
<p>
    {{ if gt .StargazersCount 0 }}
        <b>Awesome!</b>
    {{ else }}
        <b>Hang in there!</b>
    {{ end }}
    <a href="https://github.com/{{ .FullName }}">{{ .FullName }}</a>
    was created
    <time datetime="{{ .CreatedAt }}"></time>
    and now has <b>{{ .StargazersCount }}</b> stars.
</p>
```

#### 4.4.2 **图表显示与自定义功能**

如果仓库有星标数（即 `StargazersCount > 0`），页面将展示一个星标变化图表，并提供自定义功能，允许用户选择图表的配色方案。

```go
<div class="chart-review">
    <div class="chart-selection">
        <div class="button-group">
            <button data-variant="adaptive" class="active">Adaptive</button>
            <button data-variant="light">Light</button>
            <button data-variant="dark">Dark</button>
            <button data-variant="custom">Custom</button>
        </div>
        <div class="customisation">
            <label for="background">Background Color</label>
            <input id="background" name="background" type="text" value="#FFFFFF" data-coloris>
            <label for="axis">Axis Color</label>
            <input id="axis" name="axis" type="text" value="#333333" data-coloris>
            <label for="line">Line Color</label>
            <input id="line" name="line" type="text" value="#6b63ff" data-coloris>
        </div>
    </div>
    <div class="chart">
        <img src="/{{ .FullName }}.svg?variant=adaptive" id="chart" data-src="/{{ .FullName }}.svg" alt="Please try again in a few minutes. This might not work for very famous repository.">
    </div>
</div>
```

**图表展示**：图表的图片来自 URL `/{{ .FullName }}.svg`，通过 `variant=adaptive` 参数来设置图表的适应性风格。

**图表自定义**：提供了多个配色方案（自适应、浅色、深色、自定义）供用户选择，并允许用户通过颜色选择器修改背景色、轴线颜色和图表线条颜色。

#### 4.4.3 Markdown 代码生成

```html
<noscript id="code-template">## Stargazers over time
    [![Stargazers over time]($URL)](https://starchart.cc/{{ .FullName }})</noscript>
<p>
    You can include the chart on your repository's
    <code>README.md</code>
    as follows:
</p>
<div class="code-block">
    <pre class="markdown" id="code">
        <code></code>
    </pre>
    <button class="copy-btn full-width" data-clipboard-target="#code">Copy</button>
</div>
```

**Markdown 代码生成**：提供了一段 Markdown 代码，用户可以将这段代码复制到他们的 GitHub 仓库的 `README.md` 文件中，以便显示该仓库的星标变化图表。

用户可以通过点击 "Copy" 按钮将这段代码复制到剪贴板。

#### 4.4.4 **页脚**

```go
<div class="footer">
    <a href="https://www.digitalocean.com/?refcode=7e8e9efb2f77&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge">
        <img src="https://web-platforms.sfo2.cdn.digitaloceanspaces.com/WWW/Badge%201.svg"
             alt="DigitalOcean Referral Badge" width="150px"/>
    </a>
</div>
```

- 页脚包含一个 DigitalOcean 的推荐链接和徽章。

#### 4.4.5 外部 JavaScript 引用

```go
<script src="https://cdnjs.cloudflare.com/ajax/libs/timeago.js/4.0.2/timeago.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.1.0/highlight.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.8/clipboard.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/mdbassit/Coloris@latest/dist/coloris.min.js"></script>
<script src="/static/scripts.js"></script>
```

**功能**：

- `timeago.js` 用于将时间格式化为“相对时间”，例如“2 days ago”。
- `highlight.js` 用于代码高亮显示。
- `clipboard.js` 用于实现复制按钮功能。
- `Coloris` 用于颜色选择器。
- `/static/scripts.js` 可能包含自定义的 JavaScript 逻辑，处理用户交互、图表自定义等功能。