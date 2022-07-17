package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	command = "cmd"
)

func main() {
	genProto("auth")
	genProto("rental")
}

func genProto(domain string) {
	_, p, _, ok := runtime.Caller(1)
	if !ok {
		panic("error: cannot read abs path")
	}
	ps := strings.Split(p, "/")
	ps = ps[:len(ps)-2]
	fullPath := strings.Join(ps, "/")

	protoPath := fmt.Sprintf("./%s/api", domain)
	goOutPath := fmt.Sprintf("./%s/api/gen/v1", domain)

	err := os.RemoveAll(goOutPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = pathCreate(goOutPath)
	if err != nil {
		log.Fatalln(err)
	}

	showError(exec.Command(command, "/C", fmt.Sprintf("protoc -I %s --go_out %s --go_opt paths=source_relative --go-grpc_out %s --go-grpc_opt paths=source_relative %s", protoPath, goOutPath, goOutPath, fmt.Sprintf("%s.proto", domain))))

	showError(exec.Command(command, "/C", fmt.Sprintf("protoc -I %s --grpc-gateway_out %s --grpc-gateway_opt paths=source_relative --grpc-gateway_opt %s %s", protoPath, goOutPath, fmt.Sprintf("grpc_api_configuration=%s/%s.yaml", protoPath, domain), fmt.Sprintf("%s.proto", domain))))

	pdtsBinDir := fmt.Sprintf("%s/wx/miniprogram/node_modules/.bin", fullPath)
	pbtsOutDir := fmt.Sprintf("%s/wx/miniprogram/service/proto_gen/%s", fullPath, domain)

	err = os.RemoveAll(pbtsOutDir)
	if err != nil {
		log.Fatalln(err)
	}

	err = pathCreate(pbtsOutDir)
	if err != nil {
		log.Fatalln(err)
	}

	showError(exec.Command(command, "/C", fmt.Sprintf("%s -t static -w es6 %s --no-cache --no-create --no-verify --no-delimited -o %s", fmt.Sprintf("%s/pbjs", pdtsBinDir), fmt.Sprintf("%s/%s.proto", protoPath, domain), fmt.Sprintf("%s/%s_pb_tmp.js", pbtsOutDir, domain))))

	err = tracefile(fmt.Sprintf("%s/%s_pb.js", pbtsOutDir, domain), []byte(`import * as $protobuf from "protobufjs";`))
	if err != nil {
		log.Fatalln(err)
	}

	content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s_pb_tmp.js", pbtsOutDir, domain))
	if err != nil {
		log.Fatalln(err)
	}
	err = tracefile(fmt.Sprintf("%s/%s_pb.js", pbtsOutDir, domain), content)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Remove(fmt.Sprintf("%s/%s_pb_tmp.js", pbtsOutDir, domain))
	if err != nil {
		log.Fatalln(err)
	}

	showError(exec.Command(command, "/C", fmt.Sprintf("%s -o %s %s", fmt.Sprintf("%s/pbts", pdtsBinDir), fmt.Sprintf("%s/%s_pb.d.ts", pbtsOutDir, domain), fmt.Sprintf("%s/%s_pb.js", pbtsOutDir, domain))))
}

func pathCreate(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	return err
}

func tracefile(name string, content []byte) error {
	fd, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fd.Write(content)
	fd.Close()
	return nil
}

func showError(cmd *exec.Cmd) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(fmt.Sprint(err) + ": " + stderr.String())
	}
}
