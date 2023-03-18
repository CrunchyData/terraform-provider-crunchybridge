/*
Copyright 2022 Crunchy Data Solutions, Inc.

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
package bridgeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrorBadRequest = errors.New("invalid request")
	ErrorConflict   = errors.New("non-unique name specified in request")

	ErrorOldSecretFormat = errors.New("unexpected format for api secret, regeneration may be needed")
)

func errorFromAPIMessageResponse(resp *http.Response) error {
	// APIMessage is the default response format when the API function doesn't
	// return the documented response type
	var mesg APIMessage
	if resp.StatusCode != http.StatusCreated {
		err := json.NewDecoder(resp.Body).Decode(&mesg)
		if err != nil {
			// Move forward with errors based on http code
			mesg.Message = "unable to retrieve further error details"
		}
	}

	if mesg.RequestID == "" {
		return fmt.Errorf("server responded with %s: %s",
			resp.Status, mesg.Message)
	}

	return fmt.Errorf("server responded with %s, request_id: %s: %s",
		resp.Status, mesg.RequestID, mesg.Message)
}

func safeClose(outErr *error, c io.Closer, nameFormat string, a ...any) {
	err := c.Close()
	if err != nil {
		name := fmt.Sprintf(nameFormat, a...)

		*outErr = errors.Join(*outErr, fmt.Errorf(
			"failed to close %s: %w", name, err))
	}
}
