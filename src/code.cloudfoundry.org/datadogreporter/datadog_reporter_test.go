package datadogreporter_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"code.cloudfoundry.org/datadogreporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DatadogReporter", func() {
	It("sends data points to datadog on an interval", func() {
		pointBuilder := &spyPointBuilder{}
		httpClient := &spyHTTPClient{}

		reporter := datadogreporter.New(
			"api-key",
			"job-name",
			"instance-id",
			pointBuilder,
			datadogreporter.WithHost("abcdefg"),
			datadogreporter.WithInterval(10*time.Millisecond),
			datadogreporter.WithHTTPClient(httpClient),
		)
		go reporter.Run()

		Eventually(pointBuilder.buildCalled).Should(BeNumerically(">", 1))
		Eventually(httpClient.postCount).Should(BeNumerically(">", 1))
		Eventually(httpClient.url).Should(Equal("https://app.datadoghq.com/api/v1/series?api_key=api-key"))
		Eventually(httpClient.contentType).Should(Equal("application/json"))
		Eventually(httpClient.body).Should(MatchJSON(`{
			"series": [
				{
					"metric": "capacity_planning.sent",
					"points": [[1234, 4321]],
					"type": "gauge",
					"host": "abcdefg",
					"tags": [
						"event_type:metrics",
						"api_version:2",
						"job_name:job-name",
						"instance_index:instance-id"
					]
				},
				{
					"metric": "capacity_planning.read",
					"points": [[1234, 4321]],
					"type": "gauge",
					"host": "abcdefg",
					"tags": [
						"event_type:metrics",
						"api_version:2",
						"job_name:job-name",
						"instance_index:instance-id"
					]
				}
			]
		}`))
	})
})

type spyPointBuilder struct {
	mu           sync.Mutex
	_buildCalled int
}

func (s *spyPointBuilder) BuildPoints() []datadogreporter.Point {
	s.mu.Lock()
	defer s.mu.Unlock()

	s._buildCalled++

	return []datadogreporter.Point{
		{
			Metric: "capacity_planning.sent",
			Points: [][]int64{
				[]int64{int64(1234), int64(4321)},
			},
			Type: "gauge",
			Tags: []string{
				"event_type:metrics",
				"api_version:2",
			},
		},
		{
			Metric: "capacity_planning.read",
			Points: [][]int64{
				[]int64{int64(1234), int64(4321)},
			},
			Type: "gauge",
			Tags: []string{
				"event_type:metrics",
				"api_version:2",
			},
		},
	}
}

func (s *spyPointBuilder) buildCalled() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._buildCalled
}

type spyReadCloser struct{}

func (s *spyReadCloser) Close() error {
	return nil
}

func (s *spyReadCloser) Read([]byte) (int, error) {
	return 0, nil
}

type spyHTTPClient struct {
	mu           sync.Mutex
	_postCount   int
	_url         string
	_contentType string
	_body        string
}

func (s *spyHTTPClient) Post(url string, contentType string, r io.Reader) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s._postCount++

	body, err := ioutil.ReadAll(r)
	Expect(err).ToNot(HaveOccurred())

	s._url = url
	s._contentType = contentType
	s._body = string(body)

	return &http.Response{StatusCode: 201, Body: &spyReadCloser{}}, nil
}

func (s *spyHTTPClient) url() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._url
}

func (s *spyHTTPClient) contentType() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._contentType
}

func (s *spyHTTPClient) body() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._body
}

func (s *spyHTTPClient) postCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._postCount
}
