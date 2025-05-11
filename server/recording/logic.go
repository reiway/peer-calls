package recording

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/ivfwriter"
)

func (f *RecorderFactory) GetOrCreateRecorder(roomID string) *RoomRecorder {
	f.mu.Lock()
	defer f.mu.Unlock()

	recorder, exists := f.recorders[roomID]
	if !exists {
		recorder = newRoomRecorder(roomID)
		f.recorders[roomID] = recorder
	}
	return recorder
}

func (f *RecorderFactory) RemoveRecorder(roomID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.recorders, roomID)
}

func (r *RoomRecorder) AddClient(clientID string, track *webrtc.TrackRemote) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[clientID] = &ClientInfo{Track: track}
	r.incomingChan <- track
}

func (r *RoomRecorder) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, clientID)
	if len(r.clients) == 0 {
		close(r.stopChan)
		r.stopped = true
	}
}

func createIVFWriter(roomID string) *ivfwriter.IVFWriter {
	filename := fmt.Sprintf("%s_%s.ivf", roomID, time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create ivf file: %v", err)
	}

	writer, err := ivfwriter.NewWith(file)
	if err != nil {
		log.Fatalf("failed to create ivf writer: %v", err)
	}
	return writer
}

func writeTrackToIVF(track *webrtc.TrackRemote, writer *ivfwriter.IVFWriter) {
	log.Printf("Start recording track: %s", track.ID())

	for {
		packet, _, err := track.ReadRTP()
		if err != nil {
			if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "connection closed") {
				log.Printf("Track %s closed", track.ID())
			} else {
				log.Printf("Error reading RTP: %v", err)
			}
			return
		}

		if err := writer.WriteRTP(packet); err != nil {
			log.Printf("Failed to write RTP: %v", err)
			return
		}
	}
}
