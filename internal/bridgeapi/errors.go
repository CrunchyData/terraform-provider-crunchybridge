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

import "errors"

var (
	ErrorBadRequest = errors.New("invalid request")
	ErrorConflict   = errors.New("non-unique name specified in request")

	ErrorOldSecretFormat = errors.New("unexpected format for api secret, regeneration may be needed")
)
