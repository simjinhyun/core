package x

import (
	"context"
	"io"
	"os"
	"syscall"
)

func ReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

func WriteFile(path string, content []byte) {
	if err := os.WriteFile(path, content, 0644); err != nil {
		panic(err)
	}
}

func Debug(msg string) {
	println(msg)
}

func WritePipe(pipePath string, data []byte) error {
	f, err := os.OpenFile(pipePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func ReadPipe(
	ctx context.Context,
	pipePath string,
	handler func(data []byte),
) error {
	// FIFO 생성
	if err := syscall.Mkfifo(pipePath, 0600); err != nil && !os.IsExist(err) {
		return err
	}

	go func() {
		defer os.Remove(pipePath) // 종료 시 파이프 제거

		for {
			select {
			case <-ctx.Done():
				// context 취소 → 고루틴 종료
				return
			default:
				f, err := os.OpenFile(pipePath, os.O_RDONLY, 0600)
				if err != nil {
					// 에러 발생 시 잠시 대기 후 재시도
					continue
				}

				b, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}

				if len(b) > 0 {
					handler(b)
				}
			}
		}
	}()

	return nil
}
