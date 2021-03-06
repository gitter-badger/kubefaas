/*
Copyright 2019 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/srcmesh/kubefaas/pkg/controller/client"
	"github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/cli"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/httptrigger"
	"github.com/srcmesh/kubefaas/pkg/cli/console"
	flagkey "github.com/srcmesh/kubefaas/pkg/cli/flag/key"
	"github.com/srcmesh/kubefaas/pkg/cli/util"
)

type TestSubCommand struct {
	cmd.CommandActioner
}

func Test(input cli.Input) error {
	return (&TestSubCommand{}).do(input)
}

func (opts *TestSubCommand) do(input cli.Input) error {
	m := &metav1.ObjectMeta{
		Name:      input.String(flagkey.FnName),
		Namespace: input.String(flagkey.NamespaceFunction),
	}

	// Portforward to the router
	localRouterPort, err := util.SetupPortForward(util.GetFissionNamespace(), "application=kubefaas-router")
	if err != nil {
		return err
	}
	routerURL := "127.0.0.1:" + localRouterPort

	fnUri := m.Name
	if m.Namespace != metav1.NamespaceDefault {
		fnUri = fmt.Sprintf("%v/%v", m.Namespace, m.Name)
	}

	functionUrl, err := url.Parse(fmt.Sprintf("http://%s/kubefaas-function/%s", routerURL, fnUri))
	if err != nil {
		return err
	}

	console.Verbose(2, "Function test url: %v", functionUrl.String())

	queryParams := input.StringSlice(flagkey.FnTestQuery)
	if len(queryParams) > 0 {
		query := url.Values{}
		for _, q := range queryParams {
			queryParts := strings.SplitN(q, "=", 2)
			var key, value string
			if len(queryParts) == 0 {
				continue
			}
			if len(queryParts) > 0 {
				key = queryParts[0]
			}
			if len(queryParts) > 1 {
				value = queryParts[1]
			}
			query.Set(key, value)
		}
		functionUrl.RawQuery = query.Encode()
	}

	var ctx context.Context

	testTimeout := input.Duration(flagkey.FnTestTimeout)
	if testTimeout <= 0*time.Second {
		ctx = context.Background()
	} else {
		var closeCtx context.CancelFunc
		ctx, closeCtx = context.WithTimeout(context.Background(), input.Duration(flagkey.FnTestTimeout))
		defer closeCtx()
	}

	resp, err := doHTTPRequest(ctx, functionUrl.String(),
		input.StringSlice(flagkey.FnTestHeader),
		input.String(flagkey.HtMethod),
		input.String(flagkey.FnTestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response from function")
	}

	if resp.StatusCode < 400 {
		os.Stdout.Write(body)
		return nil
	}

	console.Errorf("Error calling function %s: %d; Please try again or fix the error: %s\n", m.Name, resp.StatusCode, string(body))
	log, err := printPodLogs(opts.Client(), m)
	if err != nil {
		console.Errorf("Error getting function logs from controller: %v. Try to get logs from log database.", err)
		err = Log(input)
		if err != nil {
			return errors.Wrapf(err, "error retrieving function log from log database")
		}
	} else {
		console.Info(log)
	}
	return errors.New("error getting function response")
}

func doHTTPRequest(ctx context.Context, url string, headers []string, method, body string) (*http.Response, error) {
	method, err := httptrigger.GetMethod(method)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	for _, header := range headers {
		headerKeyValue := strings.SplitN(header, ":", 2)
		if len(headerKeyValue) != 2 {
			return nil, errors.New("failed to create request without appropriate headers")
		}
		req.Header.Set(headerKeyValue[0], headerKeyValue[1])
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "error executing HTTP request")
	}

	return resp, nil
}

func printPodLogs(client client.Interface, fnMeta *metav1.ObjectMeta) (string, error) {
	reader, statusCode, err := client.V1().Misc().PodLogs(fnMeta)
	if err != nil {
		return "", errors.Wrap(err, "error executing get logs request")
	}
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", errors.Wrap(err, "error reading the response body")
	}

	if statusCode != http.StatusOK {
		return string(body), errors.Errorf("error getting logs from controller, status code: '%v'", statusCode)
	}

	return string(body), nil
}
