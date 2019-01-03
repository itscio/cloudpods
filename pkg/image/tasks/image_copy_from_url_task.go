package tasks

import (
	"context"
	"fmt"
	"net/http"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/image/models"
	"yunion.io/x/onecloud/pkg/util/httputils"
)

type ImageCopyFromUrlTask struct {
	taskman.STask
}

func init() {
	taskman.RegisterTask(ImageCopyFromUrlTask{})
}

func (self *ImageCopyFromUrlTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	image := obj.(*models.SImage)

	copyFrom, _ := self.Params.GetString("copy_from")

	log.Infof("Copy image from %s", copyFrom)

	header := http.Header{}
	// header.Set("Content-Type", "application/octet-stream")
	resp, err := httputils.Request(nil, ctx, httputils.GET, copyFrom, header, nil, false)

	if err != nil {
		msg := fmt.Sprintf("copy from url %s request fail %s", copyFrom, err)
		image.OnSaveFailed(ctx, self.UserCred, msg)
		self.SetStageFailed(ctx, msg)
		return
	}

	err = image.SaveImageFromStream(resp.Body)
	if err != nil {
		msg := fmt.Sprintf(" copy from url %s stream fail %s", copyFrom, err)
		image.OnSaveFailed(ctx, self.UserCred, msg)
		self.SetStageFailed(ctx, msg)
		return
	}

	image.OnSaveSuccess(ctx, self.UserCred, "copy from success")

	image.StartImageConvertTask(ctx, self.UserCred, "", true)

	self.SetStageComplete(ctx, nil)
}
