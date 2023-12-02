package phemex_contract

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Krisa/go-phemex/common"
	"github.com/pkg/errors"
)

func apiGetUnsigned(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http error")
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "can't read")
	}

	defer func() {
		cerr := res.Body.Close()
		// Only overwrite the retured error if the original error was nil and an
		// error occurred while closing the body.
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	if res.StatusCode >= 400 {
		apiTempErr := new(apiTempError)
		e := json.Unmarshal(data, apiTempErr)
		if e != nil {
			return nil, errors.Wrap(e, "failed to unmarshal API error")
		}
		apiErr := new(common.APIError)
		apiErr.Message = apiTempErr.Message
		apiErr.Code, e = strconv.ParseInt(apiTempErr.Code, 10, 64)
		if e != nil {
			return nil, errors.Wrap(e, "failed to parse int")
		}
		return nil, apiErr
	}
	return data, nil
}

type apiTempError struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}
