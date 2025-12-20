package x

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// Config : 설정 구조체
type Config struct {
	Data   string
	Values map[string]any
}

var ConfigX *Config

// String : 원본 JSON 문자열 그대로 반환
func (c Config) String() string {
	return c.Data
}

func LoadConfig(path string) *Config {
	Debug(path)
	data := ReadFile(path)
	Debug(string(data))
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		panic(err)
	}
	ConfigX = &Config{
		Data:   string(data),
		Values: values,
	}
	return ConfigX
}

/*
실행중인 인스턴스에게 실제 로딩된 config.json 파일 내용 요청
*/
func ShowConfig() {
	clientPID := os.Getpid()
	pipe := fmt.Sprintf("%d", clientPID)

	// FIFO 생성
	if err := syscall.Mkfifo(pipe, 0600); err != nil && !os.IsExist(err) {
		Debug("파이프 생성 실패: " + err.Error())
		return
	}
	defer os.Remove(pipe)

	done := make(chan struct{})

	go WaitForResponse(pipe, done)
	err := WritePipe(Pipe, []byte(pipe))
	if err != nil {
		Debug(err.Error())
		close(done)
	}

	<-done // 고루틴 종료까지 대기

}

/*
실행중인 인스턴스의 응답을 대기하고 응답이 오면 출력
*/
func WaitForResponse(pipe string, done chan struct{}) {
	defer close(done)
	f, err := os.OpenFile(pipe, os.O_RDONLY, 0600)
	if err != nil {
		Debug("Open 실패: " + err.Error())
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EXIT" {
			break
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		Debug("Scanner 에러: " + err.Error())
	}

	if len(lines) > 0 {
		joined := strings.Join(lines, "\n")
		Debug(joined)
	}
}
