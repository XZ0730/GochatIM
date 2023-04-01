package service

import (
	"chat/internal/vo"
	"chat/util"
	"mime/multipart"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type FileService struct {
}
type filelist struct {
	urls []string
}

var wg = sync.WaitGroup{}

func (f *FileService) SendFile(files []*multipart.FileHeader, uid, room_id string) vo.Response {

	wg.Add(len(files))
	filesUrl := &filelist{
		urls: make([]string, 0),
	}
	go uploadFile(files, uid, room_id, filesUrl)
	wg.Wait()
	return vo.BuildList(filesUrl.urls, len(filesUrl.urls), room_id)
}
func uploadFile(files []*multipart.FileHeader, uid, room_id string, filesUrl *filelist) {
	for _, filehead := range files {
		// util.UploadToQiNiu(,file,file)
		file, _ := filehead.Open()

		source := uid + ":" + uuid.NewV4().String()
		_, url := util.UploadToQiNiu(file, filehead, filehead.Size, source)
		if url == "bad token" {
			wg.Done()
			return
		}
		filesUrl.urls = append(filesUrl.urls, url)
		//		fmt.Println("url:", url)
		wg.Done()
	}
}
