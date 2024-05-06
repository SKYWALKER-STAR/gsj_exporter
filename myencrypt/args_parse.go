package myencrypt

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sync"
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

func CheckInstall(listenAddress string) Target_info {
	if fi, err := os.Stat("INSTALL"); err == nil || os.IsExist(err) {
		if fi.Size() < 1 {
			fmt.Println("INSTALL文件是空的，请手动删除此文件后，重新纳管")
			os.Exit(1)
		}
		data_encrypt, err := os.ReadFile("INSTALL")
		if err != nil {
			fmt.Println(err)
			fmt.Println("检测到INSTALL文件，但无法打开，可能是权限问题")
			os.Exit(1)
		}
		infohex, deserr := DesDecrypt(string(data_encrypt), getdefaultkey(), "")
		if deserr != nil {
			fmt.Println(deserr)
			fmt.Println("解密INSTALL文件失败，如有必要，请手动删除此文件后，重新纳管")
			os.Exit(1)
		}

		infobyte, _ := hex.DecodeString(infohex)
		b := bytes.NewBuffer(infobyte)
		infoDecoder := gob.NewDecoder(b)

		var target_info Target_info
		gobdecodeerr := infoDecoder.Decode(&target_info)
		if gobdecodeerr != nil {
			fmt.Println(gobdecodeerr)
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
	fmt.Println("url参数处理完毕")

}

var ctxShutdown, cancel = context.WithCancel(context.Background())

func starthttp(wg *sync.WaitGroup, listenAddress string) {
	srv := &http.Server{Addr: listenAddress}
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctxShutdown.Done():
			fmt.Println("ctxShutdown Done, exit")
			return
		default:
		}

		vars := r.URL.Query()
		fmt.Printf("收到url参数: %v\n", string(vars.Encode()))

		instance_id := vars.Get("instance_id")
		excludedbs := vars.Get("Excludedbs")
		hosts	:= vars.Get("hosts")
		port := vars.Get("port")
		db := vars.Get("db")
		user, _ := DesDecrypt(vars.Get("user"), getdefaultkey(), "")
		password, _ := DesDecrypt(vars.Get("password"), getdefaultkey(), "")

		// redis可能没有用户名密码
		//if len(ip) == 0 || len(port) == 0 || len(user) == 0 || len(password) == 0 {
		if len(hosts) == 0 || len(port) == 0 {
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
			fmt.Println(deserr)
			w.WriteHeader(500)
			w.Write([]byte("对相关参数解密时出现错误，请重试"))
			return
		}

		install_file, err := os.Create("INSTALL")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			w.Write([]byte("创建INSTALL文件失败，请检查运行目录及其权限"))
			return
		}
		_, err2 := install_file.WriteString(info_des)
		if err2 != nil {
			fmt.Println(err2)
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
			fmt.Println("temporary http server shutdown error", err)
		}
	})

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %v\n", err)
		}

		fmt.Println("shutdown over")
	}()
}

func getdefaultkey() string {
	if fi, err := os.Stat("HISTORY"); err == nil || os.IsExist(err) {
		if fi.Size() < 1 {
			fmt.Println("HISTORY文件是空的，部署有问题，无法继续")
			os.Exit(1)
		}

		data_encrypt, err := os.ReadFile("HISTORY")

		if err != nil {
			fmt.Println(err)
			fmt.Println("读取HISTORY文件内容失败，请检查文件权限")
		}
		key, deserr := DesDecrypt(string(data_encrypt), "e.X5T@h"+string([]byte{27}), "")
		if deserr != nil {
			fmt.Println(deserr)
			fmt.Println("解密HISTORY文件失败，部署步骤可能有问题")
		}
		return key
	} else {
		if err != nil {
			fmt.Println(err)
			fmt.Println("HISTORY文件找不到，请检查部署步骤")
			os.Exit(1)
		}
	}
	return ""
}
