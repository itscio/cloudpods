// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient/options"
)

type MongoDBListOptions struct {
	options.BaseListOptions
}

func (opts *MongoDBListOptions) Params() (jsonutils.JSONObject, error) {
	return options.ListStructToParams(opts)
}

type MongoDBUpdateOptions struct {
	options.BaseIdOptions
	Name        string
	Description string
	Delete      string `help:"Lock or not lock dbinstance" choices:"enable|disable"`
}

func (opts *MongoDBUpdateOptions) Params() (jsonutils.JSONObject, error) {
	params, err := options.StructToParams(opts)
	if err != nil {
		return nil, err
	}
	if len(opts.Delete) > 0 {
		if opts.Delete == "disable" {
			params.Add(jsonutils.JSONTrue, "disable_delete")
		} else {
			params.Add(jsonutils.JSONFalse, "disable_delete")
		}
	}
	return params, nil
}

type MongoDBBackupListOptions struct {
	options.BaseIdOptions
}

func (opts *MongoDBBackupListOptions) Property() string {
	return "backups"
}
