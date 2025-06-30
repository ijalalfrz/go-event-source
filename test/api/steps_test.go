//go:build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/thedevsaddam/gojsonq/v2"
)

type feature struct {
	headers  map[string]string
	response *http.Response
	body     string
	host     string
	testingT *testing.T
}

func (f *feature) iUseHeader(header, value string) {
	if header != "" {
		f.headers[header] = value
	}
}

func (f *feature) iUseDefaultTimestamp() {
	timestamp := time.Now().Format(time.RFC3339)
	f.headers["x-timestamp"] = timestamp
}

func (f *feature) buildHeader(req *http.Request) {
	for key, val := range f.headers {
		req.Header.Add(key, val)
	}
}

func (f *feature) iSendARequestTo(method, path string) error {
	req, err := http.NewRequestWithContext(context.Background(), method, fmt.Sprintf("http://%s%s", f.host, path), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	f.buildHeader(req)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	f.response = resp
	f.body = "" // reset

	return nil
}

func (f *feature) iSendARequestToWithJSON(method, path string, payload *godog.DocString) error {
	req, err := http.NewRequestWithContext(context.Background(), method, fmt.Sprintf("http://%s%s", f.host, path), strings.NewReader(payload.Content))
	if err != nil {
		return fmt.Errorf("create request with JSON: %w", err)
	}

	f.buildHeader(req)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request with JSON: %w", err)
	}

	f.response = resp
	f.body = "" // reset

	return nil
}

func (f *feature) theResponseCodeShouldBe(code int) error {
	if code != f.response.StatusCode {
		f.readBody()

		return fmt.Errorf("response code is not matched. Want %d, got %d\nBody: %s", code, f.response.StatusCode, f.body)
	}

	return nil
}

func (f *feature) theResponseErrorMessageShouldContain(errorMsg string) error {
	f.readBody()

	respErrorMsg := f.body
	if !strings.Contains(respErrorMsg, errorMsg) {
		return fmt.Errorf("response error message is not matched. Need to contain %s, got %s", errorMsg, respErrorMsg)
	}

	return nil
}

func (f *feature) theResponseMessageShouldContain(msg string) error {
	f.readBody()

	respMsg := f.body
	if !strings.Contains(respMsg, msg) {
		return fmt.Errorf("response message is not matched. Need to contain %s, got %s", msg, respMsg)
	}

	return nil
}

func (f *feature) theResponseBodyShouldMatchJSONSchema(path string) error {
	f.readBody()

	err := validateJSONSchema("../../"+path, f.body)
	if err != nil {
		return fmt.Errorf("validate response body: %w\nBody: %s", err, f.body)
	}

	return nil
}

func (f *feature) theNumberObjectMatchingPatternShouldEqualTo(pattern string, expected int) error {
	f.readBody()

	got := gojsonq.New().FromString(f.body).From(pattern).Count()
	if got != expected {
		return fmt.Errorf("object count is not matched. Want %d, got %d", expected, got)
	}

	return nil
}

func (f *feature) theErrorMessageShouldBe(want string) error {
	f.readBody()

	got, _ := gojsonq.New().FromString(f.body).Find("error").(string)
	if got != want {
		return fmt.Errorf("error message doesn't match. Want %s, got %s", want, got)
	}

	return nil
}

func (f *feature) readBody() error {
	if f.body != "" {
		return nil
	}

	defer f.response.Body.Close()

	body, err := io.ReadAll(f.response.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	f.body = string(body)

	return nil
}

// minifyJSON will remove whitespace in json string.
func (f *feature) minifyJSON(body string) (string, error) {

	minifiedBody := &bytes.Buffer{}
	if err := json.Compact(minifiedBody, []byte(body)); err != nil {
		return "", fmt.Errorf("minifyJSON: %w", err)
	}

	return minifiedBody.String(), nil
}

type testFeatures struct {
	fixtures *testfixtures.Loader
	testingT *testing.T
}

func (tf *testFeatures) initializeTestSuite(sc *godog.TestSuiteContext) {
	// To be run once before suite runner
	sc.BeforeSuite(func() {
		dbDSN := os.Getenv("DB_DSN")

		db, err := sql.Open("postgres", dbDSN)
		assert.Nil(tf.testingT, err)

		tf.fixtures, err = testfixtures.New(
			testfixtures.Database(db),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory("fixtures"),
		)
		assert.Nil(tf.testingT, err)
	})
}

func (tf *testFeatures) initializeScenario(ctx *godog.ScenarioContext) {
	feat := &feature{
		host:     "0.0.0.0:3001",
		headers:  make(map[string]string),
		testingT: tf.testingT,
	}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := tf.fixtures.Load()
		assert.Nil(tf.testingT, err)

		return ctx, nil
	})

	ctx.Step(`^I send a ([^"]*) with path "([^"]*)" with JSON:$`, feat.iSendARequestToWithJSON)
	ctx.Step(`^I send a ([^"]*) with path "([^"]*)"$`, feat.iSendARequestTo)
	ctx.Step(`^the response body should match JSON schema "([^"]*)"$`, feat.theResponseBodyShouldMatchJSONSchema)
	ctx.Step(`^the response code should be (\d+)$`, feat.theResponseCodeShouldBe)
	ctx.Step(`^the response error message should contain "([^"]*)"`, feat.theResponseErrorMessageShouldContain)
	ctx.Step(`^the response error message should be "([^"]*)"$`, feat.theErrorMessageShouldBe)
	ctx.Step(`^the response message should contain "([^"]*)"`, feat.theResponseMessageShouldContain)
	ctx.Step(`^the number of object matching "([^"]*)" should equal to (\d+)$`, feat.theNumberObjectMatchingPatternShouldEqualTo)
	ctx.Step(`^I set a header key "([^"]*)" with value "([^"]*)"$`, feat.iUseHeader)
	ctx.Step(`^I use default timestamp$`, feat.iUseDefaultTimestamp)

}

func TestFeatures(t *testing.T) {
	test := &testFeatures{testingT: t}

	suite := godog.TestSuite{
		TestSuiteInitializer: test.initializeTestSuite,
		ScenarioInitializer:  test.initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
