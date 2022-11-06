package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tmnhs/common/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func (l *Local) UploadFile(file *multipart.FileHeader) (string, string, error) {
	//读取文件后缀
	ext := path.Ext(file.Filename)
	name := strings.TrimSuffix(file.Filename, ext)
	//拼接新文件名
	filename := name + "_" + time.Now().Format(utils.TimeFormatDateV3) + ext
	//尝试创建路径
	mkdirErr := os.MkdirAll(l.Path, os.ModePerm)
	if mkdirErr != nil {
		return "", "", errors.New("function os.MkdirAll() Filed, err:" + mkdirErr.Error())
	}
	//拼接路径和文件名
	p := l.Path + "/" + filename

	f, openError := file.Open() //读取文件
	if openError != nil {
		return "", "", errors.New("function file.Open() Filed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭

	out, createErr := os.Create(p)
	if createErr != nil {
		return "", "", errors.New("function os.Create() Filed, err:" + createErr.Error())
	}
	defer out.Close() // 创建文件 defer 关闭

	_, copyErr := io.Copy(out, f) // 传输（拷贝）文件
	if copyErr != nil {
		return "", "", errors.New("function io.Copy() Filed, err:" + copyErr.Error())
	}
	return p, filename, nil
}

func (l *Local) DeleteFile(key string) error {
	p := l.Path + "/" + key
	if strings.Contains(p, l.Path) {
		if err := os.Remove(p); err != nil {
			return errors.New("Local file deletion failed，err:" + err.Error())
		}
	}
	return nil
}

//阿里云
func (a *AliyunOSS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	bucket, err := a.NewBucket()
	if err != nil {
		return "", "", errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}

	// 读取本地文件。
	f, openError := file.Open()
	if openError != nil {
		return "", "", errors.New("function file.Open() Failed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭
	// 上传阿里云路径 文件名格式 自己可以改 建议保证唯一性
	// fileTmpPath := filepath.Join("uploads", time.Now().Format("2006-01-02")) + "/" + file.Filename
	fileTmpPath := a.BasePath + "/" + "uploads" + "/" + time.Now().Format(utils.TimeFormatDateV1) + "/" + file.Filename

	// 上传文件流。
	err = bucket.PutObject(fileTmpPath, f)
	if err != nil {
		return "", "", errors.New("function formUploader.Put() Failed, err:" + err.Error())
	}

	return a.BucketUrl + "/" + fileTmpPath, fileTmpPath, nil
}

func (a *AliyunOSS) DeleteFile(key string) error {
	bucket, err := a.NewBucket()
	if err != nil {
		return errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}

	// 删除单个文件。objectName表示删除OSS文件时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	// 如需删除文件夹，请将objectName设置为对应的文件夹名称。如果文件夹非空，则需要将文件夹下的所有object删除后才能删除该文件夹。
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("function bucketManager.Delete() Filed, err:" + err.Error())
	}

	return nil
}
func (a *AliyunOSS) NewBucket() (*oss.Bucket, error) {
	// 创建OSSClient实例。
	client, err := oss.New(a.Endpoint, a.AccessKeyId, a.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(a.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (h *HuaWeiObs) NewHuaWeiObsClient() (client *obs.ObsClient, err error) {
	return obs.New(h.AccessKey, h.SecretKey, h.Endpoint)
}

func (h *HuaWeiObs) UploadFile(file *multipart.FileHeader) (filename string, filepath string, err error) {
	var open multipart.File
	open, err = file.Open()
	if err != nil {
		return filename, filepath, err
	}
	filename = file.Filename
	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			ObjectOperationInput: obs.ObjectOperationInput{
				Bucket: h.Bucket,
				Key:    filename,
			},
			ContentType: file.Header.Get("content-type"),
		},
		Body: open,
	}

	var client *obs.ObsClient
	client, err = h.NewHuaWeiObsClient()
	if err != nil {
		return filepath, filename, errors.New("Failed to get Huawei object storage object,error:" + err.Error())
	}

	_, err = client.PutObject(input)
	if err != nil {
		return filepath, filename, errors.New("File upload failed error:" + err.Error())
	}
	filepath = h.Path + "/" + filename
	return filepath, filename, err
}

func (h *HuaWeiObs) DeleteFile(key string) error {
	client, err := h.NewHuaWeiObsClient()
	if err != nil {
		return errors.New("Failed to get Huawei object storage object,error:" + err.Error())
	}
	input := &obs.DeleteObjectInput{
		Bucket: h.Bucket,
		Key:    key,
	}
	var output *obs.DeleteObjectOutput
	output, err = client.DeleteObject(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to delete object (%s), output:%v,error:%s", key, output, err.Error()))
	}
	return nil
}

//七牛云 上传字节文件
func (q *Qiniu) UploadBytes(data []byte) (string, string, error) {
	putPolicy := storage.PutPolicy{
		Scope: q.Bucket,
	}
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)

	upToken := putPolicy.UploadToken(mac)
	cfg := q.Config()
	formUploader := storage.NewFormUploader(cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": "hust_mall logo",
		},
	}
	uuid, _ := utils.UUID()
	fileKey := fmt.Sprintf("%s", uuid) // 文件名格式 自己可以改 建议保证唯一性

	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, fileKey, bytes.NewReader(data), dataLen, &putExtra)
	if err != nil {
		return "", "", errors.New("function formUploader.Put() Filed, err:" + err.Error())
	}
	return q.ImgPath + "/" + ret.Key, ret.Key, nil
}

//上传文件
func (q *Qiniu) UploadFile(file *multipart.FileHeader) (string, string, error) {
	putPolicy := storage.PutPolicy{Scope: q.Bucket}
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)

	upToken := putPolicy.UploadToken(mac)
	cfg := q.Config()
	formUploader := storage.NewFormUploader(cfg)

	ret := storage.PutRet{}
	putExtra := storage.PutExtra{Params: map[string]string{"x:name": "common logo"}}

	f, openError := file.Open()
	if openError != nil {
		return "", "", errors.New("function file.Open() Filed, err:" + openError.Error())
	}
	defer f.Close()                                                  // 创建文件 defer 关闭
	fileKey := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename) // 文件名格式 自己可以改 建议保证唯一性
	putErr := formUploader.Put(context.Background(), &ret, upToken, fileKey, f, file.Size, &putExtra)
	if putErr != nil {
		return "", "", errors.New("function formUploader.Put() Filed, err:" + putErr.Error())
	}
	return q.ImgPath + "/" + ret.Key, ret.Key, nil
}

func (q *Qiniu) DeleteFile(key string) error {
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)
	cfg := q.Config()
	bucketManager := storage.NewBucketManager(mac, cfg)
	if err := bucketManager.Delete(q.Bucket, key); err != nil {
		return errors.New("function bucketManager.Delete() Filed, err:" + err.Error())
	}
	return nil
}

func (q *Qiniu) Config() *storage.Config {
	cfg := storage.Config{
		UseHTTPS:      q.UseHTTPS,
		UseCdnDomains: q.UseCdnDomains,
	}
	switch q.Zone { // 根据配置文件进行初始化空间对应的机房
	case "ZoneHuadong":
		cfg.Zone = &storage.ZoneHuadong
	case "ZoneHuabei":
		cfg.Zone = &storage.ZoneHuabei
	case "ZoneHuanan":
		cfg.Zone = &storage.ZoneHuanan
	case "ZoneBeimei":
		cfg.Zone = &storage.ZoneBeimei
	case "ZoneXinjiapo":
		cfg.Zone = &storage.ZoneXinjiapo
	}
	return &cfg
}

// UploadFile upload file to COS
func (t *TencentCOS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	client := t.NewClient()
	f, openError := file.Open()
	if openError != nil {
		return "", "", errors.New("function file.Open() Filed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭
	fileKey := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename)

	_, err := client.Object.Put(context.Background(), t.PathPrefix+"/"+fileKey, f, nil)
	if err != nil {
		panic(err)
	}
	return t.BaseURL + "/" + t.PathPrefix + "/" + fileKey, fileKey, nil
}

// DeleteFile delete file form COS
func (t *TencentCOS) DeleteFile(key string) error {
	client := t.NewClient()
	name := t.PathPrefix + "/" + key
	_, err := client.Object.Delete(context.Background(), name)
	if err != nil {
		return errors.New("function bucketManager.Delete() Filed, err:" + err.Error())
	}
	return nil
}

// NewClient init COS client
func (t *TencentCOS) NewClient() *cos.Client {
	urlStr, _ := url.Parse("https://" + t.Bucket + ".cos." + t.Region + ".myqcloud.com")
	baseURL := &cos.BaseURL{BucketURL: urlStr}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  t.SecretID,
			SecretKey: t.SecretKey,
		},
	})
	return client
}
