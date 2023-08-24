package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func IsDir(dir string) bool {
	return false
}

func DirExist(dir string) bool {
	return false
}

func IsFile(file string) bool {
	return false
}

func FileExist(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

// DownloadFile 下载文件, 提供网址和目的地
func DownloadFile(url string, filename string) error {
	if info, err := os.Stat(filename); err == nil && info.Size() > 0 {
		return nil
	}

	Logger.InfoF("download %s to %s", url, filename)

	resp, err := http.Get(url)
	if err != nil {
		Logger.ErrorF("download: %s err: %v", url, err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			Logger.WarningF("download file, close body err: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		Logger.ErrorF("download: %s status code failed: %d", url, resp.StatusCode)
		return nil
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		Logger.ErrorF("download: %s open_file err: %v", url, err)
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			Logger.WarningF("download file, close file err: %v", err)
		}
	}(f)

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		Logger.ErrorF("save content error: %s err: %v", filename, err)
		return err
	}

	return nil
}

func CopyFile(dstName, srcName string) (writeen int64, err error) {
	src, err := os.Open(dstName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(srcName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
