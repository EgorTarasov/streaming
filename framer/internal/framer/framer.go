package framer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/rs/zerolog/log"
)

// This package requires ffmpeg installed on machine

// TaskInfo информация об обработке данных видео потока
type TaskInfo struct {
	VideoId int64  `json:"videoId"`
	Url     string `json:"RtspUrl"`
	Frames  int64  `json:"ProcessedFrames"`
}
type task struct {
	TaskInfo
	cancel context.CancelFunc
}

type framer struct {
	processes map[int64]*task
	mu        sync.RWMutex
}

func New() *framer {
	return &framer{
		processes: make(map[int64]*task),
	}
}

func (f *framer) AddVideoFile(ctx context.Context, videoId int64, video []byte, processFunc ProcessFrameFunc) error {
	panic("not implemented")
	return nil
}

// AddStream Добавление потока в обработку
func (f *framer) AddStream(ctx context.Context, videoId int64, rtspUrl string, processFunc ProcessFrameFunc) error {
	ctx, cancel := context.WithCancel(ctx)

	// тестовый запрос потока, если есть ошибка то завершаем поток
	_, err := f.getImageBytesFromStream(rtspUrl)
	if err != nil {
		cancel()
		return err
	}
	f.mu.Lock()
	f.processes[videoId] = &task{
		TaskInfo: TaskInfo{
			Url:     rtspUrl,
			Frames:  0,
			VideoId: videoId,
		},
		cancel: cancel,
	}
	f.mu.Unlock()

	go f.processRtsp(ctx, videoId, rtspUrl, processFunc)
	return nil
}

// RemoveStream удаление потока их об
func (f *framer) RemoveStream(_ context.Context, videoId int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.processes[videoId]
	if !ok {
		return fmt.Errorf("can't locate stream %d", videoId)
	}

	state.cancel()

	delete(f.processes, videoId)

	return nil
}

// StreamStatus получение информации об обработке одного потока
func (f *framer) StreamStatus(_ context.Context, videoId int64) (TaskInfo, error) {
	f.mu.RLock()
	state, ok := f.processes[videoId]
	defer f.mu.RUnlock()
	if !ok {
		return TaskInfo{}, fmt.Errorf("can't locate stream %d", videoId)
	}
	return TaskInfo{
		VideoId: videoId,
		Url:     state.TaskInfo.Url,
		Frames:  state.TaskInfo.Frames,
	}, nil
}

// Status получения статуса обработки потоков
func (f *framer) Status(_ context.Context) ([]TaskInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var info []TaskInfo

	for _, v := range f.processes {
		info = append(info, TaskInfo{
			VideoId: v.VideoId,
			Url:     v.Url,
			Frames:  v.Frames,
		})

	}
	return info, nil
}

// ProcessFrameFunc функция, которая вызывается при получении нового кадра
type ProcessFrameFunc = func(ctx context.Context, frameId int64, videoId int64, frame []byte)

// processRtsp запуск обработки видео потока
// позволяет запускать процесс в отдельной горутине и останавливать из внешнего контекста
func (f *framer) processRtsp(ctx context.Context, videoId int64, rtspUrl string, processFunc ProcessFrameFunc) {
	errCnt := 0
	prevErr := false

	for {

		select {
		case <-ctx.Done():
			// если прервали операцию, то заканчиваем обработку
			log.Info().Int64("stream", videoId).Msg("finishing")
			return
		default:
			if errCnt > 10 && prevErr {
				// если пропустили много кадров то завершаем горутину
				//slog.Error("failing to retrieve images from", "stream", rtspUrl)
				return
			}
			// получаем кадр и отправляем дальше
			rawImage, err := f.getImageBytesFromStream(rtspUrl)
			if err != nil {
				//slog.Info("err during retrieving", "stream", rtspUrl)
				errCnt += 1
				prevErr = true
			}
			prevErr = false
			f.mu.Lock()
			_, ok := f.processes[videoId]
			if !ok {
				f.mu.Unlock()
				return
			}
			processFunc(ctx, f.processes[videoId].Frames, videoId, rawImage)
			f.processes[videoId].Frames += 1
			f.mu.Unlock()
		}

	}
}

// getImageBytesFromStream получение одного кадра из видео потока
func (f *framer) getImageBytesFromStream(streamUrl string) ([]byte, error) {
	// Extract the image from the RTSP stream using ffmpeg and output it to stdout
	cmd := exec.Command("ffmpeg", "-rtsp_transport", "tcp", "-i", streamUrl, "-frames:v", "1", "-f", "image2pipe", "-")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func splitVideo(videoBytes []byte, frameChan chan<- []byte) error {
	tempFile, err := os.CreateTemp("", "temp-video")
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tempFile.Name())
	_, err = tempFile.Write(videoBytes)
	if err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg", "-i", tempFile.Name(), "-vf", "fps=1", "./out/out%d.png")

	// Run the command
	err = cmd.Run()
	if err != nil {
		return err
	}
	// Read splited frames and send them to the channel
	files, err := os.ReadDir("./out")
	if err != nil {
		return err
	}

	for _, file := range files {
		fileDesc, err := os.Open("./out/" + file.Name())

		frameBytes, err := io.ReadAll(fileDesc)
		if err != nil {
			return err
		}
		frameChan <- frameBytes
		_ = fileDesc.Close()
	}

	return nil
}
