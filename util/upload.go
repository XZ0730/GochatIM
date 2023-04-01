package util

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// 封装上传图片到七牛云然后返回状态和图片的url  source-->uid:roomid
func UploadToQiNiu(file multipart.File, fileheader *multipart.FileHeader, fileSize int64, source string) (int, string) {
	var AccessKey = "bvNjY2qeTRz1rNw9NfdEEFUV9yBRlDSfwoMtAWHB"
	var SerectKey = "pcovzI0wcNQO_DUTVv97PpkrWqLxz7gmSJK0EluR"
	var Bucket = "yamany"
	var ImgUrl = "rsdnd23ji.hn-bkt.clouddn.com"

	putPlicy := storage.PutPolicy{
		Scope: Bucket,
	}
	mac := qbox.NewMac(AccessKey, SerectKey)
	upToken := putPlicy.UploadToken(mac)
	cfg := storage.Config{
		Zone:          &storage.ZoneHuanan,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}
	formUploader := storage.NewResumeUploaderV2(&cfg)
	//recoder用来配置断点续传
	recorder, err := storage.NewFileRecorder(os.TempDir())
	if err != nil {
		return 300121, err.Error()
	}
	ret := storage.PutRet{}
	putExtra := storage.RputV2Extra{
		Recorder: recorder,
	}
	var filebox string
	var suffix string
	filename := strings.Split(fileheader.Filename, ".")
	if len(filename) > 1 {
		suffix = filename[len(filename)-1]
	} else {
		suffix = filename[0]
	}
	fmt.Println("filesuffix:", suffix)
	fileheader.Filename = source + "." + suffix
	key := filebox + fileheader.Filename
	fmt.Println("time:", time.Now())
	err = formUploader.Put(context.Background(), &ret, upToken, key, file, fileSize, &putExtra)
	if err != nil {
		log.Println("upload err:", err)
		return 22222, err.Error()
	}
	fmt.Println("time:", time.Now())
	url := ImgUrl + ret.Key
	return 200, url
}
