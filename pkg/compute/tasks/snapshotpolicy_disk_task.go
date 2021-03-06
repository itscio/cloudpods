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

package tasks

import (
	"context"
	"fmt"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/cloudcommon/notifyclient"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/util/logclient"
)

func init() {
	taskman.RegisterTask(SnapshotPolicyApplyTask{})
	taskman.RegisterTask(SnapshotPolicyCancelTask{})
}

type SnapshotPolicyApplyTask struct {
	taskman.STask
}

func (self *SnapshotPolicyApplyTask) taskFail(ctx context.Context, disk *models.SDisk,
	spd *models.SSnapshotPolicyDisk, reason jsonutils.JSONObject) {

	if spd != nil {
		err := spd.RealDetach(ctx, self.UserCred)
		if err != nil {
			log.Errorf("Delete snapshotpolicydisk %s failed, need to delete", spd.GetId())
		}
		snapshotpolicy, err := models.SnapshotPolicyManager.FetchSnapshotPolicyById(spd.SnapshotpolicyId)
		if err == nil {
			db.OpsLog.LogEvent(snapshotpolicy, db.ACT_SNAPSHOT_POLICY_BIND_DISK_FAIL, "", self.UserCred)
			logclient.AddActionLogWithStartable(self, snapshotpolicy, logclient.ACT_BIND_DISK, "", self.UserCred, false)
		}
	}

	disk.SetStatus(self.UserCred, compute.DISK_APPLY_SNAPSHOT_FAIL, reason.String())

	db.OpsLog.LogEvent(disk, db.ACT_APPLY_SNAPSHOT_POLICY_FAILED, reason, self.UserCred)
	logclient.AddActionLogWithStartable(self, disk, logclient.ACT_APPLY_SNAPSHOT_POLICY, reason, self.UserCred, false)
	notifyclient.NotifySystemErrorWithCtx(ctx, disk.GetId(), disk.Name, compute.DISK_APPLY_SNAPSHOT_FAIL, reason.String())
	self.SetStageFailed(ctx, reason)
}

func (self *SnapshotPolicyApplyTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	disk := obj.(*models.SDisk)
	spd := models.SSnapshotPolicyDisk{}
	data.Unmarshal(&spd, "snapshotPolicyDisk")

	var snapshotPolicy *models.SSnapshotPolicy
	if data.Contains("need_detach") {
		snapshotPolicyID, _ := data.GetString("need_detach")
		model, err := models.SnapshotPolicyManager.FetchById(snapshotPolicyID)
		if err != nil {
			self.taskFail(ctx, disk, &spd, jsonutils.NewString(err.Error()))
			return
		}
		snapshotPolicy = model.(*models.SSnapshotPolicy)
	}
	self.Params.Add(jsonutils.NewString(spd.SnapshotpolicyId), "snapshotpolicy_id")

	self.SetStage("OnPreSnapshotPolicyApplyComplete", nil)
	// pass data to next Stage without inserting database through this way
	storage, _ := disk.GetStorage()
	region, _ := storage.GetRegion()
	if err := region.GetDriver().RequestPreSnapshotPolicyApply(ctx, self.UserCred, self, disk, snapshotPolicy,
		data); err != nil {

		self.taskFail(ctx, disk, &spd, jsonutils.NewString(err.Error()))
		return
	}
}

func (self *SnapshotPolicyApplyTask) OnPreSnapshotPolicyApplyCompleteFailed(ctx context.Context, disk *models.SDisk,
	reason jsonutils.JSONObject) {

	spId, _ := self.Params.GetString("snapshotpolicy_id")
	spd, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(spId, disk.GetId())
	if err != nil {
		self.taskFail(ctx, disk, nil, reason)
		return
	}
	self.taskFail(ctx, disk, spd, reason)
}

func (self *SnapshotPolicyApplyTask) OnPreSnapshotPolicyApplyComplete(ctx context.Context, disk *models.SDisk,
	data jsonutils.JSONObject) {

	snapshotPolicy := models.SSnapshotPolicy{}
	spd := models.SSnapshotPolicyDisk{}
	data.Unmarshal(&snapshotPolicy, "snapshotPolicy")
	data.Unmarshal(&spd, "snapshotPolicyDisk")

	if data.Contains("need_detach") {
		snapshotPolicyID, _ := data.GetString("need_detach")
		spd1, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(snapshotPolicyID, disk.GetId())
		if err != nil {
			self.taskFail(ctx, disk, &spd, jsonutils.NewString(err.Error()))
			return
		}
		if spd1 != nil {
			spd1.RealDetach(ctx, self.UserCred)
		}
	}
	self.SetStage("OnSnapshotPolicyApply", nil)

	// pass data to next Stage without inserting database through this way
	storage, _ := disk.GetStorage()
	region, _ := storage.GetRegion()
	if err := region.GetDriver().
		RequestApplySnapshotPolicy(ctx, self.UserCred, self, disk, &snapshotPolicy, data); err != nil {

		self.taskFail(ctx, disk, &spd, jsonutils.NewString(err.Error()))
	}
}

func (self *SnapshotPolicyApplyTask) OnSnapshotPolicyApplyFailed(ctx context.Context, disk *models.SDisk,
	reason jsonutils.JSONObject) {

	spId, _ := self.Params.GetString("snapshotpolicy_id")
	spd, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(spId, disk.GetId())
	if err != nil {
		self.taskFail(ctx, disk, nil, reason)
		return
	}
	self.taskFail(ctx, disk, spd, reason)
}

func (self *SnapshotPolicyApplyTask) OnSnapshotPolicyApply(ctx context.Context, disk *models.SDisk, data jsonutils.JSONObject) {
	sp_id, _ := data.GetString("snapshotpolicy_id")
	spd, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(sp_id, disk.GetId())
	if err != nil {
		log.Errorf("Fetch snapshotpolicy disk failed")
	} else if spd != nil {
		spd.SetStatus(self.UserCred, compute.SNAPSHOT_POLICY_DISK_READY, "")
		snapshotpolicy, err := models.SnapshotPolicyManager.FetchSnapshotPolicyById(spd.SnapshotpolicyId)
		if err == nil {
			db.OpsLog.LogEvent(snapshotpolicy, db.ACT_SNAPSHOT_POLICY_BIND_DISK, "", self.UserCred)
			logclient.AddActionLogWithStartable(self, snapshotpolicy, logclient.ACT_BIND_DISK, "", self.UserCred, true)
		}
	}
	db.OpsLog.LogEvent(disk, db.ACT_APPLY_SNAPSHOT_POLICY, "", self.UserCred)
	logclient.AddActionLogWithStartable(self, disk, logclient.ACT_APPLY_SNAPSHOT_POLICY, "", self.UserCred, true)
	self.SetStageComplete(ctx, nil)
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

type SnapshotPolicyCancelTask struct {
	taskman.STask
}

func (self *SnapshotPolicyCancelTask) taskFail(ctx context.Context, disk *models.SDisk, spd *models.SSnapshotPolicyDisk, reason jsonutils.JSONObject) {
	if spd != nil {
		spd.SetStatus(self.UserCred, compute.SNAPSHOT_POLICY_DISK_DELETE_FAILED, "")
		snapshotpolicy, err := models.SnapshotPolicyManager.FetchSnapshotPolicyById(spd.SnapshotpolicyId)
		if err == nil {
			db.OpsLog.LogEvent(snapshotpolicy, db.ACT_SNAPSHOT_POLICY_UNBIND_DISK_FAIL, "", self.UserCred)
			logclient.AddActionLogWithStartable(self, snapshotpolicy, logclient.ACT_UNBIND_DISK, "", self.UserCred, false)
		}
	}
	disk.SetStatus(self.UserCred, compute.DISK_CALCEL_SNAPSHOT_FAIL, reason.String())

	db.OpsLog.LogEvent(disk, db.ACT_CANCEL_SNAPSHOT_POLICY_FAILED, reason, self.UserCred)
	logclient.AddActionLogWithStartable(self, disk, logclient.ACT_CANCEL_SNAPSHOT_POLICY, reason, self.UserCred, false)
	self.SetStageFailed(ctx, reason)
}

func (self *SnapshotPolicyCancelTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	disk := obj.(*models.SDisk)
	snapshotPolicyID, _ := data.GetString("snapshot_policy_id")
	spd := models.SSnapshotPolicyDisk{}
	data.Unmarshal(&spd, "snapshotPolicyDisk")
	self.Params.Add(jsonutils.NewString(snapshotPolicyID), "snapshotpolicy_id")

	model, err := models.SnapshotPolicyManager.FetchById(snapshotPolicyID)
	if err != nil {
		self.taskFail(ctx, disk, &spd, jsonutils.NewString(fmt.Sprintf("failed to fetch disk by id %s: %s", snapshotPolicyID, err.Error())))
		return
	}
	snapshotPolicy := model.(*models.SSnapshotPolicy)
	self.SetStage("OnSnapshotPolicyCancel", nil)
	storage, _ := disk.GetStorage()
	region, _ := storage.GetRegion()
	if err := region.GetDriver().RequestCancelSnapshotPolicy(ctx, self.UserCred, self, disk, snapshotPolicy, data); err != nil {

		self.taskFail(ctx, disk, &spd, jsonutils.NewString(err.Error()))
		// fmt.Sprintf("faile to detach snapshot policy %s and disk %s: %s", snapshotPolicy.Id, disk.Id, err.Error()))
	}
}

func (self *SnapshotPolicyCancelTask) OnSnapshotPolicyCancelFailed(ctx context.Context, disk *models.SDisk,
	reason jsonutils.JSONObject) {

	spId, _ := self.Params.GetString("snapshotpolicy_id")
	spd, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(spId, disk.GetId())
	if err != nil {
		self.taskFail(ctx, disk, nil, reason)
		return
	}
	self.taskFail(ctx, disk, spd, reason)
}

func (self *SnapshotPolicyCancelTask) OnSnapshotPolicyCancel(ctx context.Context, disk *models.SDisk, data jsonutils.JSONObject) {
	sp_id, _ := data.GetString("snapshotpolicy_id")
	spd, err := models.SnapshotPolicyDiskManager.FetchBySnapshotPolicyDisk(sp_id, disk.GetId())
	if err != nil {
		log.Errorf("Fechsnapshotpolicy disk failed")
	}
	if spd != nil {
		//real detach
		spd.RealDetach(ctx, self.UserCred)
		snapshotpolicy, err := models.SnapshotPolicyManager.FetchSnapshotPolicyById(spd.SnapshotpolicyId)
		if err == nil {
			db.OpsLog.LogEvent(snapshotpolicy, db.ACT_SNAPSHOT_POLICY_UNBIND_DISK, "", self.UserCred)
			logclient.AddActionLogWithStartable(self, snapshotpolicy, logclient.ACT_UNBIND_DISK, "", self.UserCred, true)
		}
	}

	db.OpsLog.LogEvent(disk, db.ACT_CANCEL_SNAPSHOT_POLICY, "", self.UserCred)
	logclient.AddActionLogWithStartable(self, disk, logclient.ACT_CANCEL_SNAPSHOT_POLICY, "", self.UserCred, true)
	self.SetStageComplete(ctx, nil)
}
