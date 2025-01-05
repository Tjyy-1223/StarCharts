## StarCharts 代码分析 - 3 - controller pkg

### 3.1 helpers.go

```go
package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"time"
)

const (
	base       = "static/templates/base.gohtml"
	repository = "static/templates/repository.gohtml"
	index      = "static/templates/index.gohtml"
)

var colorExpression = regexp.MustCompile("^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}|[a-fA-F0-9]{8})$")

func extractColor(r *http.Request, name string) (string, error) {
	color := r.URL.Query().Get(name)
	if len(color) == 0 {
		return "", nil
	}
	if colorExpression.MatchString(color) {
		return color, nil
	}
	return "", fmt.Errorf("invalid %s : %s", name, color)
}

type params struct {
	Owner      string
	Repo       string
	Line       string
	Background string
	Axis       string
	Variant    string
}

func extractSvgChartParams(r *http.Request) (*params, error) {
	backgrounndColor, err := extractColor(r, "background")
	if err != nil {
		return nil, err
	}

	axisColor, err := extractColor(r, "axis")
	if err != nil {
		return nil, err
	}

	lineColor, err := extractColor(r, "line")
	if err != nil {
		return nil, err
	}

	vars := mux.Vars(r)
	return &params{
		Owner:      vars["owner"],
		Repo:       vars["repo"],
		Background: backgrounndColor,
		Axis:       axisColor,
		Line:       lineColor,
		Variant:    r.URL.Query().Get("variant"),
	}, nil
}

func writeSvgHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Add("content-type", "image/svg+xml;charset=utf-8")
	header.Add("cache-control", "public, max-age=86400")
	header.Add("date", time.Now().Format(time.RFC1123))
	header.Add("expires", time.Now().Format(time.RFC1123))
}

func chartKey(params *params) string {
	return fmt.Sprintf(
		"%s/%s/[%s][%s][%s][%s]",
		params.Owner,
		params.Repo,
		params.Variant,
		params.Background,
		params.Axis,
		params.Line,
	)
}

```

整体来看，这段代码的主要目的是：

- 提取 HTTP 请求中的参数，尤其是图表配置参数（如颜色、图表所属仓库的 `owner` 和 `repo`）。
- 校验颜色参数的格式，确保它们符合有效的十六进制格式。
- 设置 HTTP 响应头，标明返回内容类型是 SVG 图像，并且配置了缓存策略。
- 使用 `chartKey` 生成一个唯一的键值，这可能用于缓存该图表的图像内容，以避免每次请求都重新生成图表。

该代码是一个生成和处理 SVG 图表请求的基本框架，处理了图表的个性化设置、响应头设置以及缓存控制。



### 3.2 index.go

```go
package controller

import (
	"github.com/caarlos0/httperr"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

func Index(filesystem fs.FS, version string) http.Handler {
	indexTemplate, err := template.ParseFS(filesystem, base, index)
	if err != nil {
		return nil
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		return indexTemplate.Execute(w, map[string]string{"Version": version})
	})
}

func HandleForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")
		http.Redirect(w, r, repo, http.StatusSeeOther)
	}
}
```

这段代码定义了两个 HTTP 处理函数（handlers）：

+ `Index`：用于渲染一个带有版本信息的首页模板，并返回一个处理 HTTP 请求的 handler。
+ `HandleForm`：用于处理提交的 GitHub 仓库 URL，去掉前缀并将用户重定向到对应的 GitHub 仓库页面。

#### 3.2.1 **`Index` 函数**

```go
func Index(filesystem fs.FS, version string) http.Handler {
	indexTemplate, err := template.ParseFS(filesystem, base, index)
	if err != nil {
		return nil
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		return indexTemplate.Execute(w, map[string]string{"Version": version})
	})
}
```

`Index` 函数用于渲染一个 HTML 模板并响应请求。具体步骤和作用如下：

- **输入参数**：
  - `filesystem fs.FS`: 一个 `fs.FS` 类型的文件系统接口，表示存储模板文件的文件系统。
  - `version string`: 一个字符串，通常表示应用的版本号。
- **操作流程**：
  - 首先，通过 `template.ParseFS` 解析模板文件，模板文件的路径由 `base` 和 `index` 提供（`base` 和 `index` 很可能是两个常量或者变量，表示模板文件的路径）。如果解析失败（`err != nil`），函数会返回 `nil`，表示处理失败。
  - 然后，返回一个使用 httperr.NewF 创建的 HTTP 处理器。这个处理器会在请求到来时：
    - 使用 `indexTemplate.Execute` 渲染模板，并传递一个 `map[string]string`，其中 `Version` 键对应的值是函数的输入参数 `version`。这个值通常用于在 HTML 中显示应用的版本号。
  - 这个返回的处理器会在执行时向响应写入渲染后的 HTML 内容。
- **用途**：该函数通常用于处理访问根路径或首页的请求，渲染一个带有应用版本信息的首页模板。

#### 3.2.2 **`HandleForm` 函数**

```go
func HandleForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")
		http.Redirect(w, r, repo, http.StatusSeeOther)
	}
}
```

`HandleForm` 函数用于处理表单提交，并重定向到 GitHub 上的一个项目仓库。具体步骤和作用如下：

- **返回值**：
  - 这个函数返回一个 `http.HandlerFunc`，这是一个用于处理 HTTP 请求的函数，通常用于处理表单提交。
- **操作流程**：
  - `r.FormValue("repository")` 从请求中获取名为 `repository` 的表单字段值。这个值预计是一个 URL，如 `https://github.com/username/repo`。
  - `strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")` 会去掉输入 URL 的前缀部分 `https://github.com/`，保留下来的部分（如 `username/repo`）被认为是 GitHub 上的仓库名。
  - 然后，使用 `http.Redirect` 将用户重定向到该 GitHub 仓库的 URL（`repo`）。重定向使用了 HTTP 状态码 `http.StatusSeeOther`（通常是 303 状态码），指示客户端应该进行 GET 请求。
- **用途**：该函数通常用于处理包含 GitHub 仓库信息的表单提交，将用户重定向到对应的 GitHub 仓库页面。

两者结合，可以实现一个基本的 Web 应用，显示一个版本号，并允许用户输入仓库 URL 来重定向到 GitHub 上的实际仓库页面。



### 3.3 repository.go

```go
package controller

import (
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/github"
	"fmt"
	"github.com/caarlos0/httperr"
	"github.com/gorilla/mux"
	"html/template"
	"io/fs"
	"net/http"
)

const (
	CHART_WIDTH  = 1024
	CHART_HEIGHT = 400
)

// GetRepo shows the given repo chart
func GetRepo(fsys fs.FS, gh *github.GitHub, cache *cache.Redis, version string) http.Handler {
	repositoryTemplate, err := template.ParseFS(fsys, base, repository)
	if err != nil {
		panic(err)
	}

	indexTemplate, err := template.ParseFS(fsys, base, index)
	if err != nil {
		panic(err)
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		name := fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		details, err := gh.RepoDetails(r.Context(), name)
		if err != nil{
			return indexTemplate.Execute(w, map[string]error{
				"Error": err,
			})
		}
		
		return repositoryTemplate.Execute(w, map[string]interface{}{
			"Version": version,
			"Details": details,
		})
	})
}

```

`GetRepo` 函数的主要作用是：

- 接受一个 GitHub 仓库的 `owner` 和 `repo`（仓库名）作为 URL 参数。
- 使用 `GitHub` API 获取该仓库的详细信息。
- 使用 HTML 模板渲染该仓库的详细信息并返回给用户。
- 如果发生错误（例如仓库不存在或 API 请求失败），会显示一个错误信息页面

#### **3.3.1 模板解析**：

```go
repositoryTemplate, err := template.ParseFS(fsys, base, repository)
if err != nil {
    panic(err)
}

indexTemplate, err := template.ParseFS(fsys, base, index)
if err != nil {
    panic(err)
}
```

这段代码会从 `fsys` 中加载两个 HTML 模板文件，一个是 `repositoryTemplate`，另一个是 `indexTemplate`。`base` 和 `repository`、`index` 都是模板文件的路径（可能是常量或变量），`template.ParseFS` 会将这些文件解析为模板对象。

如果模板加载失败，程序会调用 `panic`，这会导致程序崩溃并显示错误信息。

#### 3.3.2 **返回一个 HTTP 处理器**：

```go
return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
```

`httperr.NewF` 创建了一个自定义的错误处理 HTTP 处理器。在该处理器中，代码将处理 HTTP 请求并返回响应。

#### 3.3.3 **获取 GitHub 仓库信息**：

```go
name := fmt.Sprintf(
    "%s/%s",
    mux.Vars(r)["owner"],
    mux.Vars(r)["repo"],
)
details, err := gh.RepoDetails(r.Context(), name)
```

通过 `mux.Vars(r)` 获取 URL 路径参数 `owner` 和 `repo`，拼接成一个 GitHub 仓库的名字（例如：`octocat/Hello-World`）。然后，调用 `gh.RepoDetails` 方法获取该仓库的详细信息。

`gh.RepoDetails(r.Context(), name)` 可能会向 GitHub API 发送请求，获取仓库的详细信息，如仓库的描述、星标数、贡献者等。

#### 3.3.4 错误处理与模版渲染

```go
if err != nil {
    return indexTemplate.Execute(w, map[string]error{
        "Error": err,
    })
}
```

如果在获取仓库详细信息时发生错误（例如，仓库不存在或 GitHub API 请求失败），函数会渲染 `indexTemplate` 模板并返回错误信息。

#### 3.3.5 渲染仓库信息

```go
return repositoryTemplate.Execute(w, map[string]interface{}{
    "Version": version,
    "Details": details,
})
```

如果成功获取到仓库的详细信息，函数将使用 `repositoryTemplate` 模板渲染页面。模板的数据包括：

- `Version`: 应用的版本号，用于显示在页面上。
- `Details`: 仓库的详细信息（可能包含名称、描述、星标数等）。

最终，渲染后的页面会作为响应返回给用户。



### 3.4 chart.go

#### 3.4.1 代码讲解

这段代码是一个 Go 语言编写的 HTTP 请求处理程序，它定义了一个函数 `GetRepoChart`，该函数返回一个生成 GitHub 仓库星标数（stargazers）随时间变化的 SVG 图表的 HTTP handler。

该 handler 会从缓存中读取图表数据，若缓存中没有，则会通过调用 GitHub API 获取星标信息，并生成图表。生成的图表会被缓存，以便下次请求可以更快地响应。

```go
package controller

import (
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/chart"
	"B1-StarCharts/internal/chart/svg"
	"B1-StarCharts/internal/github"
	"fmt"
	"github.com/apex/log"
	"github.com/caarlos0/httperr"
	"io"
	"net/http"
	"strings"
	"time"
)

var stylesMap = map[string]string{
	"light":    chart.LightStyles,
	"dark":     chart.DarkStyles,
	"adaptive": chart.AdaptiveStyles,
}

// GetRepoChart returns the svg chart for the given repository.
func GetRepoChart(gh *github.GitHub, cache *cache.Redis) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
    // 1. 提取请求参数和处理错误
		params, err := extractSvgChartParams(r)
		if err != nil {
			log.WithError(err).Error("failed to extract params")
			return err
		}
	
    // 2. 从缓存中读取图表
		cacheKey := chartKey(params)
		name := fmt.Sprintf("%s/%s", params.Owner, params.Repo)
		log := log.WithField("repo", name).WithField("variant", params.Variant)

		cachedChart := ""
		if err = cache.Get(cacheKey, &cachedChart); err == nil {
			writeSvgHeaders(w)
			log.Debugf("using cached chart")
			_, err := fmt.Fprintf(w, cachedChart)
			return err
		}
		
    // 3.从 GitHub 获取仓库和星标数据
		defer log.Trace("collect_stars").Stop(nil)
		repo, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return httperr.Wrap(err, http.StatusBadRequest)
		}

		stargazers, err := gh.Stargazers(r.Context(), repo)
		if err != nil {
			log.WithError(err).Error("failed to get stars")
			writeSvgHeaders(w)
			_, err := w.Write([]byte(errSvg(err)))
			return err
		}

    // 4.生成图表数据
		series := chart.Series{
			StrokeWidth: 2,
			Color:       params.Line,
		}
		for i, star := range stargazers {
			series.XValues = append(series.XValues, star.StarredAt)
			series.YValues = append(series.YValues, float64(i+1))
		}

		if len(series.XValues) < 2 {
			log.Info("not enough results, adding some fake ones")
			series.XValues = append(series.XValues, time.Now())
			series.YValues = append(series.YValues, 1)
		}
		
    // 5. 图表样式和渲染
		graph := &chart.Chart{
			Width:      CHART_WIDTH,
			Height:     CHART_HEIGHT,
			Styles:     stylesMap[params.Variant],
			Background: params.Background,
			XAxis: chart.XAxis{
				Name:        "Time",
				Color:       params.Axis,
				StrokeWidth: 2,
			},
			YAxis: chart.YAxis{
				Name:        "Stargazers",
				Color:       params.Axis,
				StrokeWidth: 2,
			},
			Series: series,
		}
		
    // 缓存生成的图表
		defer log.Trace("chart").Stop(&err)

		writeSvgHeaders(w)
		cacheBuffer := &strings.Builder{}
		graph.Render(io.MultiWriter(w, cacheBuffer))
		err = cache.Put(cacheKey, cacheBuffer.String())
		if err != nil {
			log.WithError(err).Error("failed to cache chart")
		}
		return nil
	})
}

// 7.如果获取星标数据失败，会生成一个包含错误信息的 SVG 图表。错误信息会显示为红色文本，居中显示在图表中。
func errSvg(err error) string {
	return svg.SVG().
		Attr("width", svg.Px(CHART_WIDTH)).
		Attr("height", svg.Px(CHART_HEIGHT)).
		ContentFunc(func(writer io.Writer) {
			svg.Text().
				Attr("fill", "red").
				Attr("x", svg.Px(CHART_WIDTH/2)).
				Attr("y", svg.Px(CHART_HEIGHT/2)).
				Content(err.Error()).
				Render(writer)
		}).String()
}
```

#### 3.4.2 defer log.Trace 作用

```go
package main

import (
	"errors"
	"fmt"
	"github.com/apex/log"
	"time"
)

func fetchData() (string, error) {
	// 模拟从数据库获取数据
	time.Sleep(1 * time.Second)
	return "some chart data", nil
}

func generateChart(data string) (string, error) {
	// 模拟图表生成
	time.Sleep(2 * time.Second)
	if data == "" {
		return "", errors.New("no data available")
	}
	return "chart generated", nil
}

func createChart() error {
	var err error

	// 延迟执行 Trace 的 Stop，结束时将 err 的状态报告给 Stop
	defer log.Trace("chart").Stop(&err)

	// 步骤 1: 获取数据
	data, err := fetchData()
	if err != nil {
		return fmt.Errorf("fetch data failed: %w", err)
	}

	defer log.Trace("chart - 2").Stop(&err)
	// 步骤 2: 生成图表
	chart, err := generateChart(data)
	if err != nil {
		return fmt.Errorf("generate chart failed: %w", err)
	}

	// 假设生成成功，输出图表内容
	fmt.Println(chart)
	return nil
}

func main() {
	createChart()
}


/*
2025/01/05 16:19:27  info chart                    
2025/01/05 16:19:28  info chart - 2                
chart generated
2025/01/05 16:19:30  info chart - 2                 duration=2001
2025/01/05 16:19:30  info chart                     duration=3002
*/
```

在 `createChart` 函数中，使用 `defer log.Trace("chart").Stop(&err)` 来确保在函数执行结束时，跟踪记录被停止，并根据 `err` 变量的状态记录成功或失败。

+ 首先通过调用 `fetchData` 获取数据，如果出错，则立即返回并记录错误信息。

+ 然后调用 `generateChart` 生成图表，如果出错，同样会返回并记录错误信息。

如果两步都成功完成，则输出生成的图表内容，并最终返回 `nil`。

在这个例子中，`defer log.Trace("chart").Stop(&err)` 确保了函数 `createChart` 的执行过程被跟踪，并根据 `err` 变量的值（错误或成功）记录相关的日志信息。**这种方式对于监控和调试复杂的业务流程非常有用，能够清晰地了解操作是否成功、在哪个环节失败，以及失败的具体原因。**