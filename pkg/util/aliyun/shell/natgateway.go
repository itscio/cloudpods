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

package shell

import (
	"yunion.io/x/onecloud/pkg/util/aliyun"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

func init() {
	type NatGatewayListOptions struct {
		Limit  int `help:"page size"`
		Offset int `help:"page offset"`
	}
	shellutils.R(&NatGatewayListOptions{}, "natgateway-list", "List NAT gateways", func(cli *aliyun.SRegion, args *NatGatewayListOptions) error {
		gws, total, e := cli.GetNatGateways("", "", args.Offset, args.Limit)
		if e != nil {
			return e
		}
		printList(gws, total, args.Offset, args.Limit, []string{})
		return nil
	})

	type SNatEntryListOptions struct {
		ID     string `help:"SNat Table ID"`
		Limit  int    `help:"page size"`
		Offset int    `help:"page offset"`
	}
	shellutils.R(&SNatEntryListOptions{}, "snat-entry-list", "List SNAT entries", func(cli *aliyun.SRegion, args *SNatEntryListOptions) error {
		entries, total, e := cli.GetSNATEntries(args.ID, args.Offset, args.Limit)
		if e != nil {
			return e
		}
		printList(entries, total, args.Offset, args.Limit, []string{})
		return nil
	})

	type DNatEntryListOptions struct {
		ID     string `help:"DNat Table ID"`
		Limit  int    `help:"page size"`
		Offset int    `help:"page offset"`
	}
	shellutils.R(&DNatEntryListOptions{}, "dnat-entry-list", "List DNAT entries", func(cli *aliyun.SRegion, args *DNatEntryListOptions) error {
		entries, total, e := cli.GetForwardTableEntries(args.ID, args.Offset, args.Limit)
		if e != nil {
			return e
		}
		printList(entries, total, args.Offset, args.Limit, []string{})
		return nil
	})

}
