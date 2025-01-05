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
