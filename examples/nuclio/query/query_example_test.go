// +build integration

package query

import (
	"fmt"
	"github.com/nuclio/nuclio-test-go"
	"github.com/v3io/v3io-tsdb/pkg/config"
	"github.com/v3io/v3io-tsdb/pkg/tsdb/tsdbtest"
	"os"
	"path/filepath"
	"testing"
)

func TestQueryIntegration(t *testing.T) {
	v3ioConfig, err := config.LoadConfig(filepath.Join("..", "..", "..", config.DefaultConfigurationFileName))
	defer tsdbtest.SetUp(t, v3ioConfig)()
	tsdbConfig = fmt.Sprintf(`path: "%v"`, v3ioConfig.Path)

	url := os.Getenv("V3IO_SERVICE_URL")
	if url == "" {
		url = v3ioConfig.V3ioUrl
	}

	data := nutest.DataBind{
		Name: "db0", Url: url, Container: v3ioConfig.Container, User: v3ioConfig.Username, Password: v3ioConfig.Password}
	tc, err := nutest.NewTestContext(Handler, false, &data)
	if err != nil {
		t.Fatal(err)
	}

	err = tc.InitContext(InitContext)
	if err != nil {
		t.Fatal(err)
	}
	testEvent := nutest.TestEvent{
		Body: []byte(queryEvent),
	}
	resp, err := tc.Invoke(&testEvent)
	tc.Logger.InfoWith("Run complete", "resp", resp, "err", err)
	fmt.Println(resp)

}