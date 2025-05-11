package recording

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type RecorderFactory struct {
	mu        sync.Mutex
	recorders map[string]*RoomRecorder // key = roomID
}

func NewRecorderFactory() *RecorderFactory {
	return &RecorderFactory{
		recorders: make(map[string]*RoomRecorder),
	}
}

type ClientInfo struct {
	ClientId string
	Track    *webrtc.TrackRemote
}

type RoomRecorder struct {
	roomID       string
	clients      map[string]*ClientInfo
	mu           sync.Mutex
	incomingChan chan *webrtc.TrackRemote
	stopChan     chan struct{}
	stopped      bool
}

func newRoomRecorder(roomID string) *RoomRecorder {
	r := &RoomRecorder{
		roomID:       roomID,
		clients:      make(map[string]*ClientInfo),
		incomingChan: make(chan *webrtc.TrackRemote, 10),
		stopChan:     make(chan struct{}),
	}
	go r.run()
	return r
}

func (r *RoomRecorder) run() {
	ivfWriter := createIVFWriter(r.roomID) // Hàm này bạn tự viết theo trước
	defer ivfWriter.Close()

	for {
		select {
		case track := <-r.incomingChan:
			// Ghi dữ liệu từ track
			go writeTrackToIVF(track, ivfWriter)

		case <-r.stopChan:
			return
		}
	}
}
