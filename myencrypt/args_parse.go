package myencrypt

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"net/http"
	"os"
	"sync"
	"log"
)

type Target_info struct {
	InstanceId	string
	Excludedbs	string
	Hosts		string
	Port		string
	DB		string
	User		string
	Password	string
}

var (
	logger = log.New(os.Stdout,"",log.Lshortfile | log.Ldate | log.Ltime)
)

func CheckInstall(listenAddress string) Target_info {
	if fi, err := os.Stat("INSTALL"); err == nil || os.IsExist(err) {
		if fi.Size() < 1 {
			logger.Println("INSTALL文件是空的，请手动删除此文件后，重新纳管")
			os.Exit(1)
		}
		data_encrypt, err := os.ReadFile("INSTALL")
		if err != nil {
			logger.Println(err)
			logger.Println("检测到INSTALL文件，但无法打开，可能是权限问题")
			os.Exit(1)
		}
		infohex, deserr := DesDecrypt(string(data_encrypt), getdefaultkey(), "")
		if deserr != nil {
			logger.Println(deserr)
			logger.Println("解密INSTALL文件失败，如有必要，请手动删除此文件后，重新纳管")
			os.Exit(1)
		}

		infobyte, _ := hex.DecodeString(infohex)
		b := bytes.NewBuffer(infobyte)
		infoDecoder := gob.NewDecoder(b)

		var target_info Target_info
		gobdecodeerr := infoDecoder.Decode(&target_info)
		if gobdecodeerr != nil {
			logger.Println(gobdecodeerr)
			os.Exit(1)
		}
		return target_info

	} else {
		dealArgs(listenAddress)
		// test
		return CheckInstall(listenAddress)

	}

}

func dealArgs(listenAddress string) {
	serverDone := &sync.WaitGroup{}
	serverDone.Add(1)
	starthttp(serverDone, listenAddress)
	serverDone.Wait()
	logger.Println("url参数处理完毕")

}

var ctxShutdown, cancel = context.WithCancel(context.Background())

func starthttp(wg *sync.WaitGroup, listenAddress string) {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: listenAddress,Handler: mux}
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctxShutdown.Done():
			logger.Println("ctxShutdown Done, exit")
			return
		default:
		}

		//从Url获取参数
		vars := r.URL.Query()
		logger.Printf("收到url参数: %v\n", string(vars.Encode()))
		instance_id := vars.Get("instance_id")
		if len(instance_id) == 0 {
			logger.Printf("instance id is:%s",instance_id)
		}
		excludedbs := vars.Get("excludedbs")
		if excludedbs == "" {
			logger.Printf("excludedbs is:%s",excludedbs)
		}
		hosts	:= vars.Get("hosts")
		if hosts == "" {
			logger.Printf("excludedbs is:%s",excludedbs)
		}
		port := vars.Get("port")
		if port == "" {
			logger.Printf("port is:%s",excludedbs)
		}
		db := vars.Get("db")
		if db == "" {
			logger.Printf("database is :%s",excludedbs)
		}
		user, _ := DesDecrypt(vars.Get("user"), getdefaultkey(), "")
		if user == ""{
			logger.Printf("user is :%s",excludedbs)
		}
		password, _ := DesDecrypt(vars.Get("password"), getdefaultkey(), "")
		if password == "" {
			logger.Printf("password is :%s",excludedbs)
		}

		if len(hosts) == 0 || len(port) == 0 || len(instance_id) == 0 || len(db) == 0 {
			w.WriteHeader(500)
			w.Write([]byte("url参数不足或解析失败，如果正在重新纳管请稍等，否则请联系支持人员"))
			return
		}

		var b bytes.Buffer
		infoEncoder := gob.NewEncoder(&b)
		infoEncoder.Encode(Target_info{instance_id, excludedbs, hosts, port, db, user, password})
		infohex := hex.EncodeToString(b.Bytes())
		info_des, deserr := DesEncrypt(infohex, getdefaultkey(), "")
		if deserr != nil {
			logger.Println(deserr)
			w.WriteHeader(500)
			w.Write([]byte("对相关参数解密时出现错误，请重试"))
			return
		}

		install_file, err := os.Create("INSTALL")
		if err != nil {
			logger.Println(err)
			w.WriteHeader(500)
			w.Write([]byte("创建INSTALL文件失败，请检查运行目录及其权限"))
			return
		}
		_, err2 := install_file.WriteString(info_des)
		if err2 != nil {
			logger.Println(err2)
			w.WriteHeader(500)
			w.Write([]byte("写入INSTALL文件失败，请检查运行目录及其权限"))
			return
		}
		install_file.Close()
		w.WriteHeader(200)
		w.Write([]byte("登陆信息已保存，下次请求即可看到指标"))

		cancel()
		// graceful-shutdown
		shutdownerr := srv.Shutdown(context.Background())
		if shutdownerr != nil {
			logger.Println("temporary http server shutdown error", err)
		}
	})

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Printf("ListenAndServe(): %v\n", err)
		}

		logger.Println("shutdown over")
	}()
}

func getdefaultkey() string {
	if fi, err := os.Stat("HISTORY"); err == nil || os.IsExist(err) {
		if fi.Size() < 1 {
			logger.Println("HISTORY文件是空的，部署有问题，无法继续")
			os.Exit(1)
		}

		data_encrypt, err := os.ReadFile("HISTORY")

		if err != nil {
			logger.Println(err)
			logger.Println("读取HISTORY文件内容失败，请检查文件权限")
		}
		key, deserr := DesDecrypt(string(data_encrypt), "e.X5T@h"+string([]byte{27}), "")
		if deserr != nil {
			logger.Println(deserr)
			logger.Println("解密HISTORY文件失败，部署步骤可能有问题")
		}
		return key
	} else {
		if err != nil {
			logger.Println(err)
			logger.Println("HISTORY文件找不到，请检查部署步骤")
			os.Exit(1)
		}
	}
	return ""
}
