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

package taskman

import (
	"context"
	"fmt"
	"runtime/debug"

	"yunion.io/x/jsonutils"

	api "yunion.io/x/onecloud/pkg/apis/notify"
	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/cloudcommon/notifyclient"
)

var taskWorkMan *appsrv.SWorkerManager
var taskWorkerTable map[string]*appsrv.SWorkerManager

func init() {
	taskWorkMan = appsrv.NewWorkerManager("TaskWorkerManager", 4, 1024, true)
	taskWorkerTable = make(map[string]*appsrv.SWorkerManager)
}

type taskTask struct {
	taskId string
	data   jsonutils.JSONObject
}

func (t *taskTask) Run() {
	TaskManager.execTask(t.taskId, t.data)
}

func (t *taskTask) Dump() string {
	return jsonutils.Marshal(t).PrettyString()
}

func runTask(taskId string, data jsonutils.JSONObject) error {
	taskName := TaskManager.getTaskName(taskId)
	if len(taskName) == 0 {
		return fmt.Errorf("no such task??? task_id=%s", taskId)
	}
	worker := taskWorkMan
	if workerMan, ok := taskWorkerTable[taskName]; ok {
		worker = workerMan
	}

	task := &taskTask{
		taskId: taskId,
		data:   data,
	}

	isOk := worker.Run(task, nil, func(err error) {
		data := jsonutils.NewDict()
		data.Add(jsonutils.NewString(taskName), "task_name")
		data.Add(jsonutils.NewString(taskId), "task_id")
		data.Add(jsonutils.NewString(string(debug.Stack())), "stack")
		data.Add(jsonutils.NewString(err.Error()), "error")
		notifyclient.SystemExceptionNotify(context.TODO(), api.ActionSystemPanic, api.TOPIC_RESOURCE_TASK, data)
	})
	if !isOk {
		return fmt.Errorf("worker %s(%s) not running may be droped", taskName, taskId)
	}
	return nil
}
