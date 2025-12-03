package prometheus

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	Gin = "gin"
)

var logger *zap.Logger
var defaultMetricPath = "/metrics"

/*
Standard varsayılana metrics
counter, counter_vec, gauge, gauge_vec,
histogram, histogram_vec, summary, summary_vec
*/
var reqCnt = &Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"code", "method", "handler", "host", "url"}}

var reqDur = &Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Type:        "histogram_vec",
	Args:        []string{"code", "method", "url"},
}

var resSz = &Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Type:        "summary"}

var reqSz = &Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Type:        "summary"}

var standardMetrics = []*Metric{
	reqCnt,
	reqDur,
	resSz,
	reqSz,
}

/*
RequestCounterURLLabelMappingFn, middleware'a tedarik edilebilecek bir fonksiyondur.
Bu fonksiyon, istek sayacının "url" etiketinin kardinalitesini kontrol eder,
bu bazı bağlamlarda gerekli olabilir.
Örneğin, "/customer/:name" rotası için her olası müşteri adı için
bir zaman serisi üretmek istemiyorsanız, bu fonksiyonu kullanabilirsiniz:

	func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, p := range c.Params {
			if p.Key == "name" {
				url = strings.Replace(url, p.Value, ":name", 1)
				break
			}
		}
		return url
	}

Bu fonksiyon, "/customer/alice" ve "/customer/bob" URL'lerini
şablon "/customer/:name" ile eşler.
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Metric, her metrik için ad, açıklama, tür, ID ve
// prometheus.Collector türünü (yani CounterVec, Summary, vb.) tanımlar
type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

// Prometheus, örneğin topladığı metrikleri ve yolunu içerir
type Prometheus struct {
	reqCnt        *prometheus.CounterVec
	reqDur        *prometheus.HistogramVec
	reqSz, resSz  prometheus.Summary
	router        *gin.Engine
	listenAddress string
	Ppg           PrometheusPushGateway

	MetricsList []*Metric
	MetricsPath string

	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn

	// gin.Context string'ini prometheus URL etiketi olarak kullan
	URLLabelFromContext string
}

// PrometheusPushGateway, bir Prometheus pushgateway'e veri gönderme konfigürasyonunu içerir (isteğe bağlı)
type PrometheusPushGateway struct {

	// Push aralığı saniye cinsinden
	PushIntervalsn time.Duration

	// Push Gateway URL'si, format http://domain:port
	// JOBNAME istediğiniz herhangi bir dize olabilir
	PushGatewayURL string

	// Yerel metrikler URL'si, bu gelecekte atlanabilir
	// eğer prometheus common/expfmt kullanılarak uygulanmışsa
	MetricsURL string

	// pushgateway iş adı, varsayılan olarak "gin"dir
	Job string
}

// NewPrometheus, belirli bir alt sistem adı ile yeni bir metrik seti oluşturur
func NewPrometheus(subsystem string, customMetricsList ...[]*Metric) *Prometheus {
	var metricsList []*Metric

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus( string, <optional []*Metric> ).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	metricsList = append(metricsList, standardMetrics...)

	p := &Prometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path // varsayılan olarak URL'yi olduğu gibi döndür
		},
	}

	p.registerMetrics(subsystem)

	return p
}

// SetPushGateway, metrikleri belirli aralıklarla pushGatewayURL üzerinden bir uzak pushgateway'e gönderir
// Push aralığı pushIntervalSeconds'dir. Metrikler metricsURL'den alınır
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushIntervalsn time.Duration) {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.MetricsURL = metricsURL
	p.Ppg.PushIntervalsn = pushIntervalsn
	p.startPushTicker()
}

// SetPushGatewayJob iş adı, varsayılan olarak "gin"dir
func (p *Prometheus) SetPushGatewayJob(j string) {
	p.Ppg.Job = j
}

// SetListenAddress, metrikleri bir adreste sergilemek için ayarlar. Ayarlanmamışsa,
// metrikler, kullanılan gin motorunun aynı adresinde sergilenecektir
func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

// SetListenAddressWithRouter, metrikleri sergilemek için ayrı bir router kullanır.
// (bu, GET /metrics gibi şeyleri içeriğin erişim günlüklerinden çıkarır).
func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
}

// SetMetricsPath, metrik yollarını ayarlar
func (p *Prometheus) SetMetricsPath(e *gin.Engine) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

// SetMetricsPathWithAuth, kimlik doğrulaması ile metrik yollarını ayarlar
func (p *Prometheus) SetMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
	}

}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func (p *Prometheus) getMetrics() []byte {
	// HTTP isteği gönderilir ve hata kontrol edilir
	response, err := http.Get(p.Ppg.MetricsURL)
	if err != nil {
		// Hata durumunda Zap logger kullanılarak hata loglanır
		logger.Error("Error fetching metrics", zap.Error(err))
		return nil
	}
	// response.Body kapatma işlemi yapılmalı ancak sadece başarılı bir istekte
	defer func() {
		if err := response.Body.Close(); err != nil {
			logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	// Body okunurken de hata kontrolü yapılır
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("Error reading response body", zap.Error(err))
		return nil
	}

	// Başarılı şekilde metrikler döndürülür
	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "gin"
	}
	return p.Ppg.PushGatewayURL + "/metrics/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	if err != nil {
		logger.Error("Error creating request for push gateway", zap.Error(err))
		return
	}
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		logger.Error("Error sending to push gateway", zap.Error(err))
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Ppg.PushIntervalsn)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

// NewMetric, Metric.Type'a bağlı olarak prometheus.Collector'ı ilişkilendirir
func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

func (p *Prometheus) registerMetrics(subsystem string) {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			logger.Error(
				"Metric could not be registered in Prometheus",
				zap.String("metric_name", metricDef.Name),
				zap.Error(err),
			)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(prometheus.Summary)
		case reqSz:
			p.reqSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}
}

// Gin motoruna ara yazılımı ekler.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
}

// UseWithAuth, BasicAuth ile bir gin motoruna ara yazılımı ekler.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPathWithAuth(e, accounts)
}

// HandlerFunc, ara yazılım için işleyici fonksiyonunu tanımlar
func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.ReqCntURLLabelMappingFn(c)
		// jlambert Oct 2018 - sidecar specific mod
		if len(p.URLLabelFromContext) > 0 {
			u, found := c.Get(p.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}
		p.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// Not: r.Form ve r.MultipartForm'un r.URL'ye dahil olduğu varsayılır.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
