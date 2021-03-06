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

package misc

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules/compute"
)

func init() {
	type RegionTaskListOptions struct {
		ObjName  string `help:"object name"`
		ObjId    string `help:"object id"`
		TaskName string `help:"task name"`
	}
	R(&RegionTaskListOptions{}, "region-task-list", "List tasks on region server", func(s *mcclient.ClientSession, args *RegionTaskListOptions) error {
		params := jsonutils.Marshal(args)
		result, err := compute.ComputeTasks.List(s, params)
		if err != nil {
			return err
		}
		printList(result, compute.ComputeTasks.GetColumns(s))
		return nil
	})

	type RegionTaskShowOptions struct {
		ID string `help:"ID or name of the task"`
	}
	R(&RegionTaskShowOptions{}, "region-task-show", "Show details of a region task", func(s *mcclient.ClientSession, args *RegionTaskShowOptions) error {
		result, err := compute.ComputeTasks.Get(s, args.ID, nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type TaskListOptions struct {
		ObjName     string `help:"object name"`
		ObjId       string `help:"object id"`
		TaskName    string `help:"task name"`
		ServiceType string `choices:"image|cloudid|cloudevent|devtool|ansible|identity|notify|log|compute|compute_v2"`
	}
	R(&TaskListOptions{}, "task-list", "List tasks", func(s *mcclient.ClientSession, args *TaskListOptions) error {
		params := jsonutils.Marshal(args)
		man := compute.TasksManager{}
		result, err := man.List(s, params)
		if err != nil {
			return err
		}
		printList(result, man.GetColumns(s))
		return nil
	})

	type TaskShowOptions struct {
		ID          string `help:"ID or name of the task"`
		ServiceType string `choices:"image|cloudid|cloudevent|devtool|ansible|identity|notify|log|compute|compute_v2"`
	}

	R(&TaskShowOptions{}, "task-show", "Show details of a task", func(s *mcclient.ClientSession, args *TaskShowOptions) error {
		man := compute.TasksManager{}
		params := jsonutils.Marshal(args)
		result, err := man.Get(s, args.ID, params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

}
