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

package options

type NatGatewayListOptions struct {
	Vpc         string `help:"vpc id or name"`
	Cloudregion string `help:"cloudreigon id or name"`

	BaseListOptions
}

type NatDTableListOptions struct {
	Natgateway string `help:"Natgateway name or id"`

	BaseListOptions
}

type NatSTableListOptions struct {
	Natgateway string `help:"Natgateway name or id"`
	Network    string `help:"Network id or name"`

	BaseListOptions
}
